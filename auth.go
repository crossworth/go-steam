package steam

import (
	"crypto/sha1"
	"errors"
	"time"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

type Auth struct {
	client *Client
}

type SentryHash []byte

type LogOnDetails struct {
	Username string

	// If logging into an account without a login key, the account's password.
	Password string

	// If you have a Steam Guard email code, you can provide it here.
	AuthCode string

	// If you have a Steam Guard mobile two-factor authentication code, you can provide it here.
	TwoFactorCode  string
	SentryFileHash SentryHash
	LoginKey       string

	// true if you want to get a login key which can be used in lieu of
	// a password for subsequent logins. false or omitted otherwise.
	ShouldRememberPassword bool
}

// Log on with the given details. You must always specify username and
// password OR username and loginkey. For the first login, don't set an authcode or a hash and you'll
//  receive an error (EResult_AccountLogonDenied)
// and Steam will send you an authcode. Then you have to login again, this time with the authcode.
// Shortly after logging in, you'll receive a MachineAuthUpdateEvent with a hash which allows
// you to login without using an authcode in the future.
//
// If you don't use Steam Guard, username and password are enough.
//
// After the event EMsg_ClientNewLoginKey is received you can use the LoginKey
// to login instead of using the password.
func (a *Auth) LogOn(details *LogOnDetails) error {
	if details.Username == "" {
		return errors.New("steam/auth: username must be set")
	}

	if details.Password == "" && details.LoginKey == "" {
		return errors.New("steam/auth: Password or LoginKey must be set")
	}

	logon := &pb.CMsgClientLogon{}
	logon.AccountName = &details.Username
	logon.Password = &details.Password

	if details.AuthCode != "" {
		logon.AuthCode = proto.String(details.AuthCode)
	}

	if details.TwoFactorCode != "" {
		logon.TwoFactorCode = proto.String(details.TwoFactorCode)
	}

	logon.ClientLanguage = proto.String("english")
	logon.ProtocolVersion = proto.Uint32(steamlang.MsgClientLogon_CurrentProtocol)
	logon.ShaSentryfile = details.SentryFileHash

	if details.LoginKey != "" {
		logon.LoginKey = proto.String(details.LoginKey)
	}

	if details.ShouldRememberPassword {
		logon.ShouldRememberPassword = proto.Bool(details.ShouldRememberPassword)
	}

	a.client.setSteamID(steamid.New(
		steamlang.EAccountType_Individual,
		steamlang.EUniverse_Public,
		0,
		steamid.DesktopInstance,
	))

	a.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientLogon, logon))

	return nil
}

func (a *Auth) HandlePacket(packet *protocol.Packet) {
	switch packet.EMsg {
	case steamlang.EMsg_ClientLogOnResponse:
		a.handleLogOnResponse(packet)
	case steamlang.EMsg_ClientNewLoginKey:
		a.handleLoginKey(packet)
	case steamlang.EMsg_ClientLoggedOff:
		a.handleLoggedOff(packet)
	case steamlang.EMsg_ClientUpdateMachineAuth:
		a.handleUpdateMachineAuth(packet)
	case steamlang.EMsg_ClientAccountInfo:
		a.handleAccountInfo(packet)
	}
}

func (a *Auth) handleLogOnResponse(packet *protocol.Packet) {
	if !packet.IsProto {
		a.client.Fatalf("received non-proto logon response")
		return
	}

	body := &pb.CMsgClientLogonResponse{}
	msg := packet.ReadProtoMsg(body)

	switch result := steamlang.EResult(body.GetEresult()); result {
	case steamlang.EResult_OK:
		a.client.setSessionID(msg.Header.Proto.GetClientSessionid())
		a.client.setSteamID(steamid.SteamID(msg.Header.Proto.GetSteamid()))
		a.client.Web.webLoginKey = *body.WebapiAuthenticateUserNonce

		go a.client.heartbeatLoop(time.Duration(body.GetOutOfGameHeartbeatSeconds()))

		a.client.Emit(&LoggedOnEvent{
			Result:         result,
			ExtendedResult: steamlang.EResult(body.GetEresultExtended()),
			AccountFlags:   steamlang.EAccountFlags(body.GetAccountFlags()),
			ClientSteamId:  steamid.SteamID(body.GetClientSuppliedSteamid()),
			Body:           body,
		})
	case steamlang.EResult_Fail, steamlang.EResult_ServiceUnavailable, steamlang.EResult_TryAnotherCM:
		// some error on Steam's side, we'll get an EOF later
		a.client.Emit(&SteamFailureEvent{Result: result})
	default:
		a.client.Emit(&LogOnFailedEvent{Result: result})
		a.client.Disconnect()
	}
}

func (a *Auth) handleLoginKey(packet *protocol.Packet) {
	body := &pb.CMsgClientNewLoginKey{}
	packet.ReadProtoMsg(body)

	pbAccepted := &pb.CMsgClientNewLoginKeyAccepted{
		UniqueId: proto.Uint32(body.GetUniqueId()),
	}

	a.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientNewLoginKeyAccepted, pbAccepted))

	a.client.Emit(&LoginKeyEvent{
		UniqueId: body.GetUniqueId(),
		LoginKey: body.GetLoginKey(),
	})
}

func (a *Auth) handleLoggedOff(packet *protocol.Packet) {
	var result steamlang.EResult

	if packet.IsProto {
		body := &pb.CMsgClientLoggedOff{}
		packet.ReadProtoMsg(body)
		result = steamlang.EResult(body.GetEresult())
	} else {
		body := &steamlang.MsgClientLoggedOff{}
		packet.ReadClientMsg(body)
		result = body.Result
	}

	a.client.Emit(&LoggedOffEvent{Result: result})
}

func (a *Auth) handleUpdateMachineAuth(packet *protocol.Packet) {
	body := &pb.CMsgClientUpdateMachineAuth{}

	packet.ReadProtoMsg(body)

	hash := sha1.New()

	if _, err := hash.Write(packet.Data); err != nil {
		a.client.Fatalf("auth: error generating sha1 hash of machine auth: %v", err)
		return
	}

	sha := hash.Sum(nil)

	pbAuthRes := &pb.CMsgClientUpdateMachineAuthResponse{
		ShaFile: sha,
	}

	msg := protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientUpdateMachineAuthResponse, pbAuthRes)

	msg.SetTargetJobId(packet.SourceJobId)

	a.client.Write(msg)
	a.client.Emit(&MachineAuthUpdateEvent{sha})
}

func (a *Auth) handleAccountInfo(packet *protocol.Packet) {
	body := &pb.CMsgClientAccountInfo{}
	packet.ReadProtoMsg(body)
	a.client.Emit(&AccountInfoEvent{
		PersonaName:          body.GetPersonaName(),
		Country:              body.GetIpCountry(),
		CountAuthedComputers: body.GetCountAuthedComputers(),
		AccountFlags:         steamlang.EAccountFlags(body.GetAccountFlags()),
		FacebookId:           body.GetFacebookId(),
		FacebookName:         body.GetFacebookName(),
	})
}
