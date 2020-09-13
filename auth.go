package steam

import (
	"errors"
	"time"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"google.golang.org/protobuf/proto"

	"github.com/13k/go-steam/cryptoutil"
	"github.com/13k/go-steam/protocol"
	"github.com/13k/go-steam/steamid"
)

type SentryHash []byte

type LogOnDetails struct {
	Username string
	// Can be omitted if using a LoginKey.
	Password string
	// Previously saved login key.
	LoginKey string
	// Steam Guard email code.
	AuthCode string
	// Steam Guard two-factor authentication code.
	TwoFactorCode string
	// Tells Steam to generate a login key to be used on subsequent logins without a password.
	// A `LoginKeyEvent` event will be emitted with the LoginKey to be saved.
	ShouldRememberPassword bool
	// Previously saved machine identification hash.
	// It can be saved when the event `MachineAuthUpdateEvent` is emitted.
	SentryFileHash SentryHash
	// Previously saved CellID from a `LoggedOnEvent` event.
	CellID uint32
	// LoginID uniquely identifies a logon session (required if establishing more than one active
	// session to the given account).
	LoginID uint32
}

type Auth struct {
	client *Client
}

var _ protocol.PacketHandler = (*Auth)(nil)

func NewAuth(client *Client) *Auth {
	return &Auth{client: client}
}

// LogOn logs on with the given details.
//
// You must always specify username and password OR username and loginkey. For the first login,
// don't set an authcode or a hash and you'll receive an error (EResult_AccountLogonDenied) and
// Steam will send you an authcode. Then you have to login again, this time with the authcode.
//
// Shortly after logging in, you'll receive a `MachineAuthUpdateEvent` with a hash which allows you
// to login without using an authcode in the future.
//
// If you don't use Steam Guard, username and password are enough.
//
// After the event EMsg_ClientNewLoginKey is received you can use the LoginKey to login instead of
// using the password.
func (a *Auth) LogOn(details *LogOnDetails) error {
	if details.Username == "" {
		return errors.New("steam/auth: username must be set")
	}

	if details.Password == "" && details.LoginKey == "" {
		return errors.New("steam/auth: Password or LoginKey must be set")
	}

	machineID, err := NewMachineID()

	if err != nil {
		return err
	}

	machineIDAuth, err := machineID.Auth()

	if err != nil {
		return err
	}

	steamID := steamid.New(
		steamlang.EAccountType_Individual,
		steamlang.EUniverse_Public,
		0,
		steamid.DesktopInstance,
	)

	logon := &pb.CMsgClientLogon{
		ProtocolVersion:           proto.Uint32(steamlang.MsgClientLogon_CurrentProtocol),
		AccountName:               proto.String(details.Username),
		Password:                  proto.String(details.Password),
		ShouldRememberPassword:    proto.Bool(details.ShouldRememberPassword),
		ClientLanguage:            proto.String("english"),
		ShaSentryfile:             details.SentryFileHash,
		EresultSentryfile:         proto.Int32(int32(steamlang.EResult_FileNotFound)),
		SupportsRateLimitResponse: proto.Bool(true),
		ChatMode:                  proto.Uint32(2),
		MachineId:                 machineIDAuth,
	}

	if details.AuthCode != "" {
		logon.AuthCode = proto.String(details.AuthCode)
	}

	if details.TwoFactorCode != "" {
		logon.TwoFactorCode = proto.String(details.TwoFactorCode)
	}

	if details.LoginKey != "" {
		logon.LoginKey = proto.String(details.LoginKey)
	}

	if details.SentryFileHash != nil {
		logon.EresultSentryfile = proto.Int32(int32(steamlang.EResult_OK))
	}

	if details.CellID != 0 {
		logon.CellId = proto.Uint32(details.CellID)
	}

	if details.LoginID != 0 {
		logon.ObfuscatedPrivateIp = &pb.CMsgIPAddress{
			Ip: &pb.CMsgIPAddress_V4{V4: details.LoginID},
		}
	}

	msg := protocol.NewProtoMessage(steamlang.EMsg_ClientLogon, logon)

	msg.SetSessionID(0)
	msg.SetSteamID(steamID)

	a.client.setSteamID(steamID)
	a.client.Write(msg)

	return nil
}

// LogOnAnonymous logs on with an anonymous user account on the global cell id
// https://github.com/SteamDatabase/SteamTracking/blob/master/ClientExtracted/steam/cached/CellMap.vdf
func (a *Auth) LogOnAnonymousOnGlobalCellID() {
	steamID := steamid.New(
		steamlang.EAccountType_AnonUser,
		steamlang.EUniverse_Public,
		0,
		steamid.UnknownInstance,
	)

	logon := &pb.CMsgClientLogon{
		ProtocolVersion:           proto.Uint32(steamlang.MsgClientLogon_CurrentProtocol),
		EresultSentryfile:         proto.Int32(int32(steamlang.EResult_FileNotFound)),
		SupportsRateLimitResponse: proto.Bool(false),
		AnonUserTargetAccountName: proto.String("anonymous"),
		ChatMode:                  proto.Uint32(2),
		CellId:                    proto.Uint32(0),
	}

	msg := protocol.NewProtoMessage(steamlang.EMsg_ClientLogon, logon)

	msg.SetSessionID(0)
	msg.SetSteamID(steamID)

	a.client.setSteamID(steamID)
	a.client.Write(msg)
}

func (a *Auth) HandlePacket(packet *protocol.Packet) {
	switch packet.EMsg() {
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
	if !packet.IsProto() {
		a.client.Fatalf("auth/LogOnResponse: received non-proto message")
		return
	}

	body := &pb.CMsgClientLogonResponse{}
	msg, err := packet.ReadProtoMsg(body)

	if err != nil {
		a.client.Fatalf("auth/LogOnResponse: error reading message: %v", err)
		return
	}

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
			ClientSteamID:  steamid.SteamID(body.GetClientSuppliedSteamid()),
			Body:           body,
		})
	case steamlang.EResult_AccountLogonDenied:
		fallthrough
	case steamlang.EResult_TwoFactorCodeMismatch:
		fallthrough
	case steamlang.EResult_AccountLoginDeniedNeedTwoFactor:
		authCode := result == steamlang.EResult_AccountLogonDenied
		twoFactorCode := result == steamlang.EResult_AccountLoginDeniedNeedTwoFactor
		lastCodeWrong := false

		if result == steamlang.EResult_TwoFactorCodeMismatch {
			lastCodeWrong = true
		}

		a.client.Emit(&SteamGuardEvent{
			AuthCode:      authCode,
			TwoFactorCode: twoFactorCode,
			Domain:        body.GetEmailDomain(),
			LastCodeWrong: lastCodeWrong,
		})
	case steamlang.EResult_Fail, steamlang.EResult_ServiceUnavailable, steamlang.EResult_TryAnotherCM:
		// some error on Steam's side, we'll get an EOF later
		a.client.Emit(&FailureEvent{Result: result})
	default:
		a.client.Emit(&LogOnFailedEvent{Result: result})
		a.client.Disconnect()
	}
}

func (a *Auth) handleLoginKey(packet *protocol.Packet) {
	body := &pb.CMsgClientNewLoginKey{}

	if _, err := packet.ReadProtoMsg(body); err != nil {
		a.client.Errorf("auth/LoginKey: error reading message: %v", err)
		return
	}

	pbAccepted := &pb.CMsgClientNewLoginKeyAccepted{
		UniqueId: proto.Uint32(body.GetUniqueId()),
	}

	a.client.Write(protocol.NewProtoMessage(steamlang.EMsg_ClientNewLoginKeyAccepted, pbAccepted))

	a.client.Emit(&LoginKeyEvent{
		UniqueID: body.GetUniqueId(),
		LoginKey: body.GetLoginKey(),
	})
}

func (a *Auth) handleLoggedOff(packet *protocol.Packet) {
	var result steamlang.EResult

	if packet.IsProto() {
		body := &pb.CMsgClientLoggedOff{}

		if _, err := packet.ReadProtoMsg(body); err != nil {
			a.client.Errorf("auth/LoggedOff: error reading message: %v", err)
			return
		}

		result = steamlang.EResult(body.GetEresult())
	} else {
		body := &steamlang.MsgClientLoggedOff{}

		if _, err := packet.ReadClientMsg(body); err != nil {
			a.client.Errorf("auth/LoggedOff: error reading message: %v", err)
			return
		}

		result = body.Result
	}

	a.client.Emit(&LoggedOffEvent{Result: result})
}

func (a *Auth) handleUpdateMachineAuth(packet *protocol.Packet) {
	body := &pb.CMsgClientUpdateMachineAuth{}

	if _, err := packet.ReadProtoMsg(body); err != nil {
		a.client.Errorf("auth/UpdateMachineAuth: error reading message: %v", err)
		return
	}

	sha1sum := cryptoutil.SHA1Sum(packet.Data)

	pbAuthRes := &pb.CMsgClientUpdateMachineAuthResponse{
		ShaFile: sha1sum,
	}

	msg := protocol.NewProtoMessage(steamlang.EMsg_ClientUpdateMachineAuthResponse, pbAuthRes)

	msg.SetTargetJobID(packet.SourceJobID())

	a.client.Write(msg)
	a.client.Emit(&MachineAuthUpdateEvent{Hash: sha1sum})
}

func (a *Auth) handleAccountInfo(packet *protocol.Packet) {
	body := &pb.CMsgClientAccountInfo{}

	if _, err := packet.ReadProtoMsg(body); err != nil {
		a.client.Errorf("auth/AccountInfo: error reading message: %v", err)
		return
	}

	a.client.Emit(&AccountInfoEvent{
		PersonaName:          body.GetPersonaName(),
		Country:              body.GetIpCountry(),
		CountAuthedComputers: body.GetCountAuthedComputers(),
		AccountFlags:         steamlang.EAccountFlags(body.GetAccountFlags()),
		FacebookID:           body.GetFacebookId(),
		FacebookName:         body.GetFacebookName(),
	})
}
