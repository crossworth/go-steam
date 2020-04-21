package steam

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/cryptoutil"
	"github.com/13k/go-steam/protocol"
	"google.golang.org/protobuf/proto"
)

type Web struct {
	// 64 bit alignment
	relogOnNonce uint32

	// The `sessionid` cookie required to use the steam website.
	// This cookie may contain a characters that will need to be URL-escaped, otherwise
	// Steam (probably) interprets is as a string.
	// When used as an URL parameter this is automatically escaped by the Go HTTP package.
	SessionID string
	// The `steamLogin` cookie required to use the steam website. Already URL-escaped.
	// This is only available after calling LogOn().
	SteamLogin string
	// The `steamLoginSecure` cookie required to use the steam website over HTTPs. Already URL-escaped.
	// This is only availbile after calling LogOn().
	SteamLoginSecure string

	webLoginKey string

	client *Client
}

func (w *Web) HandlePacket(packet *protocol.Packet) {
	switch packet.EMsg {
	case steamlang.EMsg_ClientNewLoginKey:
		w.handleNewLoginKey(packet)
	case steamlang.EMsg_ClientRequestWebAPIAuthenticateUserNonceResponse:
		w.handleAuthNonceResponse(packet)
	}
}

// LogOn fetches the `steamLogin` cookie.
//
// Returns an error if called before the first WebSessionIdEvent.
func (w *Web) LogOn() error {
	if w.webLoginKey == "" {
		return errors.New("steam/web: session not initialized")
	}

	go func() {
		// retry three times. yes, I know about loops.
		err := w.apiLogOn()

		if err != nil {
			err = w.apiLogOn()

			if err != nil {
				err = w.apiLogOn()
			}
		}

		if err != nil {
			w.client.Emit(WebLogOnErrorEvent(err))
			return
		}
	}()

	return nil
}

func (w *Web) apiLogOn() error {
	sessionKey := make([]byte, 32)

	if _, err := rand.Read(sessionKey); err != nil {
		return err
	}

	cryptedSessionKey, err := cryptoutil.RSAEncrypt(GetPublicKey(steamlang.EUniverse_Public), sessionKey)

	if err != nil {
		return err
	}

	ciph, err := aes.NewCipher(sessionKey)

	if err != nil {
		return err
	}

	cryptedLoginKey, err := cryptoutil.SymmetricEncrypt(ciph, []byte(w.webLoginKey))

	if err != nil {
		return err
	}

	data := make(url.Values)
	data.Add("format", "json")
	data.Add("steamid", w.client.SteamID().FormatString())
	data.Add("sessionkey", string(cryptedSessionKey))
	data.Add("encrypted_loginkey", string(cryptedLoginKey))

	resp, err := http.PostForm("https://api.steampowered.com/ISteamUserAuth/AuthenticateUser/v0001", data)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		// our web login key has expired, request a new one
		atomic.StoreUint32(&w.relogOnNonce, 1)
		msg := &pb.CMsgClientRequestWebAPIAuthenticateUserNonce{}
		w.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientRequestWebAPIAuthenticateUserNonce, msg))
		return nil
	} else if resp.StatusCode != 200 {
		return errors.New("steam.Web.apiLogOn: request failed with status " + resp.Status)
	}

	result := &struct {
		Authenticateuser struct {
			Token       string
			TokenSecure string
		}
	}{}

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	w.SteamLogin = result.Authenticateuser.Token
	w.SteamLoginSecure = result.Authenticateuser.TokenSecure

	w.client.Emit(&WebLoggedOnEvent{})

	return nil
}

func (w *Web) handleNewLoginKey(packet *protocol.Packet) {
	msg := &pb.CMsgClientNewLoginKey{}
	packet.ReadProtoMsg(msg)

	acceptMsg := &pb.CMsgClientNewLoginKeyAccepted{
		UniqueId: proto.Uint32(msg.GetUniqueId()),
	}

	w.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientNewLoginKeyAccepted, acceptMsg))

	// number -> string -> bytes -> base64
	uniqueIDStr := strconv.FormatUint(uint64(msg.GetUniqueId()), 10)
	w.SessionID = base64.StdEncoding.EncodeToString([]byte(uniqueIDStr))

	w.client.Emit(&WebSessionIDEvent{})
}

func (w *Web) handleAuthNonceResponse(packet *protocol.Packet) {
	// this has to be the best name for a message yet.
	msg := &pb.CMsgClientRequestWebAPIAuthenticateUserNonceResponse{}
	packet.ReadProtoMsg(msg)
	w.webLoginKey = msg.GetWebapiAuthenticateUserNonce()

	// if the nonce was specifically requested in apiLogOn(),
	// don't emit an event.
	if atomic.CompareAndSwapUint32(&w.relogOnNonce, 1, 0) {
		w.LogOn()
	}
}
