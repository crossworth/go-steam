package steam

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"sync"
	"time"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/protocol"
	"github.com/13k/go-steam/rwu"
	"github.com/13k/go-steam/socialcache"
	"github.com/13k/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

// Social provides access to social aspects of Steam.
type Social struct {
	mutex sync.RWMutex

	name         string
	avatar       string
	personaState steamlang.EPersonaState

	Friends *socialcache.FriendsList
	Groups  *socialcache.GroupsList
	Chats   *socialcache.ChatsList

	client *Client
}

func newSocial(client *Client) *Social {
	return &Social{
		Friends: socialcache.NewFriendsList(),
		Groups:  socialcache.NewGroupsList(),
		Chats:   socialcache.NewChatsList(),
		client:  client,
	}
}

// GetAvatar the local user's avatar
func (s *Social) GetAvatar() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.avatar
}

// GetPersonaName the local user's persona name
func (s *Social) GetPersonaName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.name
}

// SetPersonaName the local user's persona name and broadcasts it over the network
func (s *Social) SetPersonaName(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.name = name
	s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientChangeStatus, &pb.CMsgClientChangeStatus{
		PersonaState: proto.Uint32(uint32(s.personaState)),
		PlayerName:   proto.String(name),
	}))
}

// GetPersonaState the local user's persona state
func (s *Social) GetPersonaState() steamlang.EPersonaState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.personaState
}

// SetPersonaState the local user's persona state and broadcasts it over the network
func (s *Social) SetPersonaState(state steamlang.EPersonaState) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.personaState = state
	s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientChangeStatus, &pb.CMsgClientChangeStatus{
		PersonaState: proto.Uint32(uint32(state)),
	}))
}

// SendMessage a chat message to ether a room or friend
func (s *Social) SendMessage(to steamid.SteamID, entryType steamlang.EChatEntryType, message string) {
	switch to.AccountType().Enum() {
	case steamlang.EAccountType_Individual, steamlang.EAccountType_ConsoleUser: // Friend
		s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendMsg, &pb.CMsgClientFriendMsg{
			Steamid:       proto.Uint64(to.Uint64()),
			ChatEntryType: proto.Int32(int32(entryType)),
			Message:       []byte(message),
		}))
	case steamlang.EAccountType_Clan, steamlang.EAccountType_Chat: // Chat room
		chatID := to.ClanToChat()
		s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientChatMsg{
			ChatMsgType:     entryType,
			SteamIdChatRoom: chatID.Uint64(),
			SteamIdChatter:  s.client.SteamID().Uint64(),
		}, []byte(message)))
	}
}

// AddFriend adds a friend to your friends list or accepts a friend. You'll receive a
// FriendStateEvent for every new/changed friend.
func (s *Social) AddFriend(id steamid.SteamID) {
	s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientAddFriend, &pb.CMsgClientAddFriend{
		SteamidToAdd: proto.Uint64(id.Uint64()),
	}))
}

// RemoveFriend removes a friend from your friends list
func (s *Social) RemoveFriend(id steamid.SteamID) {
	s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientRemoveFriend, &pb.CMsgClientRemoveFriend{
		Friendid: proto.Uint64(id.Uint64()),
	}))
}

// IgnoreFriend ignores or unignores a friend on Steam
func (s *Social) IgnoreFriend(id steamid.SteamID, setIgnore bool) {
	ignore := uint8(1) //True
	if !setIgnore {
		ignore = uint8(0) //False
	}
	s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientSetIgnoreFriend{
		MySteamId:     s.client.SteamID().Uint64(),
		SteamIdFriend: id.Uint64(),
		Ignore:        ignore,
	}, make([]byte, 0)))
}

// RequestFriendListInfo requests persona state for a list of specified SteamIDs
func (s *Social) RequestFriendListInfo(ids []steamid.SteamID, requestedInfo steamlang.EClientPersonaStateFlag) {
	friends := make([]uint64, len(ids))

	for i, id := range ids {
		friends[i] = id.Uint64()
	}

	s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientRequestFriendData, &pb.CMsgClientRequestFriendData{
		PersonaStateRequested: proto.Uint32(uint32(requestedInfo)),
		Friends:               friends,
	}))
}

// RequestFriendInfo requests persona state for a specified SteamID
func (s *Social) RequestFriendInfo(id steamid.SteamID, requestedInfo steamlang.EClientPersonaStateFlag) {
	s.RequestFriendListInfo([]steamid.SteamID{id}, requestedInfo)
}

// RequestProfileInfo requests profile information for a specified SteamID
func (s *Social) RequestProfileInfo(id steamid.SteamID) {
	s.client.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendProfileInfo, &pb.CMsgClientFriendProfileInfo{
		SteamidFriend: proto.Uint64(id.Uint64()),
	}))
}

// RequestOfflineMessages requests all offline messages and marks them as read
/* TODO: Determine if this is possible to re-implement
func (s *Social) RequestOfflineMessages() {
	pbmsg := &CMsgClientFSGetFriendMessageHistoryForOfflineMessages{}
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientFSGetFriendMessageHistoryForOfflineMessages, pbmsg))
}
*/

// JoinChat attempts to join a chat room
func (s *Social) JoinChat(id steamid.SteamID) {
	chatID := id.ClanToChat()
	s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientJoinChat{
		SteamIdChat: chatID.Uint64(),
	}, make([]byte, 0)))
}

// LeaveChat attempts to leave a chat room
// TODO: handle errors
func (s *Social) LeaveChat(id steamid.SteamID) {
	chatID := id.ClanToChat()
	payload := &bytes.Buffer{}

	_ = binary.Write(payload, binary.LittleEndian, s.client.SteamID().Uint64())                   // ChatterActedOn
	_ = binary.Write(payload, binary.LittleEndian, uint32(steamlang.EChatMemberStateChange_Left)) // StateChange
	_ = binary.Write(payload, binary.LittleEndian, s.client.SteamID().Uint64())                   // ChatterActedBy

	s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientChatMemberInfo{
		SteamIdChat: chatID.Uint64(),
		Type:        steamlang.EChatInfoType_StateChange,
	}, payload.Bytes()))
}

// KickChatMember the specified chat member from the given chat room
func (s *Social) KickChatMember(room steamid.SteamID, user steamid.SteamID) {
	chatID := room.ClanToChat()
	s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientChatAction{
		SteamIdChat:        chatID.Uint64(),
		SteamIdUserToActOn: user.Uint64(),
		ChatAction:         steamlang.EChatAction_Kick,
	}, make([]byte, 0)))
}

// BanChatMember the specified chat member from the given chat room
func (s *Social) BanChatMember(room steamid.SteamID, user steamid.SteamID) {
	chatID := room.ClanToChat()
	s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientChatAction{
		SteamIdChat:        chatID.Uint64(),
		SteamIdUserToActOn: user.Uint64(),
		ChatAction:         steamlang.EChatAction_Ban,
	}, make([]byte, 0)))
}

// UnbanChatMember the specified chat member from the given chat room
func (s *Social) UnbanChatMember(room steamid.SteamID, user steamid.SteamID) {
	chatID := room.ClanToChat()
	s.client.Write(protocol.NewClientMsg(&steamlang.MsgClientChatAction{
		SteamIdChat:        chatID.Uint64(),
		SteamIdUserToActOn: user.Uint64(),
		ChatAction:         steamlang.EChatAction_UnBan,
	}, make([]byte, 0)))
}

// HandlePacket handles a Steam packet.
func (s *Social) HandlePacket(packet *protocol.Packet) {
	switch packet.EMsg {
	case steamlang.EMsg_ClientPersonaState:
		s.handlePersonaState(packet)
	case steamlang.EMsg_ClientClanState:
		s.handleClanState(packet)
	case steamlang.EMsg_ClientFriendsList:
		s.handleFriendsList(packet)
	case steamlang.EMsg_ClientFriendMsgIncoming:
		s.handleFriendMsg(packet)
	case steamlang.EMsg_ClientAccountInfo:
		s.handleAccountInfo(packet)
	case steamlang.EMsg_ClientAddFriendResponse:
		s.handleFriendResponse(packet)
	case steamlang.EMsg_ClientChatEnter:
		s.handleChatEnter(packet)
	case steamlang.EMsg_ClientChatMsg:
		s.handleChatMsg(packet)
	case steamlang.EMsg_ClientChatMemberInfo:
		s.handleChatMemberInfo(packet)
	case steamlang.EMsg_ClientChatActionResult:
		s.handleChatActionResult(packet)
	case steamlang.EMsg_ClientChatInvite:
		s.handleChatInvite(packet)
	case steamlang.EMsg_ClientSetIgnoreFriendResponse:
		s.handleIgnoreFriendResponse(packet)
	case steamlang.EMsg_ClientFriendProfileInfoResponse:
		s.handleProfileInfoResponse(packet)
		// case EMsg_ClientFSGetFriendMessageHistoryResponse:
		// s.handleFriendMessageHistoryResponse(packet)
	}
}

func (s *Social) handleAccountInfo(_ *protocol.Packet) {
	// Just fire the personainfo, Auth handles the callback
	flags := steamlang.EClientPersonaStateFlag_PlayerName |
		steamlang.EClientPersonaStateFlag_Presence |
		steamlang.EClientPersonaStateFlag_SourceID

	s.RequestFriendInfo(s.client.SteamID(), flags)
}

func (s *Social) handleFriendsList(packet *protocol.Packet) {
	list := &pb.CMsgClientFriendsList{}

	packet.ReadProtoMsg(list)

	var friends []steamid.SteamID

	for _, friend := range list.GetFriends() {
		steamID := steamid.SteamID(friend.GetUlfriendid())

		if steamID.AccountType().IsClan() {
			rel := steamlang.EClanRelationship(friend.GetEfriendrelationship())

			if rel == steamlang.EClanRelationship_None {
				s.Groups.Remove(steamID)
			} else {
				s.Groups.Add(socialcache.Group{
					SteamID:      steamID,
					Relationship: rel,
				})
			}

			if list.GetBincremental() {
				s.client.Emit(&GroupStateEvent{SteamID: steamID, Relationship: rel})
			}
		} else {
			rel := steamlang.EFriendRelationship(friend.GetEfriendrelationship())

			if rel == steamlang.EFriendRelationship_None {
				s.Friends.Remove(steamID)
			} else {
				s.Friends.Add(socialcache.Friend{
					SteamID:      steamID,
					Relationship: rel,
				})
			}

			if list.GetBincremental() {
				s.client.Emit(&FriendStateEvent{steamID, rel})
			}
		}

		if !list.GetBincremental() {
			friends = append(friends, steamID)
		}
	}

	if !list.GetBincremental() {
		s.RequestFriendListInfo(friends, protocol.DefaultPersonaStateFlagInfoRequest)
		s.client.Emit(&FriendsListEvent{})
	}
}

func (s *Social) handlePersonaState(packet *protocol.Packet) {
	list := &pb.CMsgClientPersonaState{}

	packet.ReadProtoMsg(list)

	flags := steamlang.EClientPersonaStateFlag(list.GetStatusFlags())

	for _, friend := range list.GetFriends() {
		id := steamid.SteamID(friend.GetFriendid())

		if id == s.client.SteamID() { // this is our client id
			s.mutex.Lock()

			if friend.GetPlayerName() != "" {
				s.name = friend.GetPlayerName()
			}

			avatar := hex.EncodeToString(friend.GetAvatarHash())

			if protocol.ValidAvatar(avatar) {
				s.avatar = avatar
			}

			s.mutex.Unlock()
		} else if id.AccountType().IsIndividual() {
			if (flags & steamlang.EClientPersonaStateFlag_PlayerName) == steamlang.EClientPersonaStateFlag_PlayerName {
				if friend.GetPlayerName() != "" {
					s.Friends.SetName(id, friend.GetPlayerName())
				}
			}
			if (flags & steamlang.EClientPersonaStateFlag_Presence) == steamlang.EClientPersonaStateFlag_Presence {
				avatar := hex.EncodeToString(friend.GetAvatarHash())
				if protocol.ValidAvatar(avatar) {
					s.Friends.SetAvatar(id, avatar)
				}
				s.Friends.SetPersonaState(id, steamlang.EPersonaState(friend.GetPersonaState()))
				s.Friends.SetPersonaStateFlags(id, steamlang.EPersonaStateFlag(friend.GetPersonaStateFlags()))
			}
			if (flags & steamlang.EClientPersonaStateFlag_GameDataBlob) == steamlang.EClientPersonaStateFlag_GameDataBlob {
				s.Friends.SetGameAppID(id, friend.GetGamePlayedAppId())
				s.Friends.SetGameID(id, friend.GetGameid())
				s.Friends.SetGameName(id, friend.GetGameName())
			}
		} else if id.AccountType().IsClan() {
			if (flags & steamlang.EClientPersonaStateFlag_PlayerName) == steamlang.EClientPersonaStateFlag_PlayerName {
				if friend.GetPlayerName() != "" {
					s.Groups.SetName(id, friend.GetPlayerName())
				}
			}
			if (flags & steamlang.EClientPersonaStateFlag_Presence) == steamlang.EClientPersonaStateFlag_Presence {
				avatar := hex.EncodeToString(friend.GetAvatarHash())
				if protocol.ValidAvatar(avatar) {
					s.Groups.SetAvatar(id, avatar)
				}
			}
		}

		// TODO: update with current protobuf (CMsgClientPersonaState.Friend) fields
		s.client.Emit(&PersonaStateEvent{
			StatusFlags:            flags,
			FriendID:               id,
			State:                  steamlang.EPersonaState(friend.GetPersonaState()),
			StateFlags:             steamlang.EPersonaStateFlag(friend.GetPersonaStateFlags()),
			GameAppID:              friend.GetGamePlayedAppId(),
			GameID:                 friend.GetGameid(),
			GameName:               friend.GetGameName(),
			GameServerIP:           friend.GetGameServerIp(),
			GameServerPort:         friend.GetGameServerPort(),
			QueryPort:              friend.GetQueryPort(),
			SourceSteamID:          steamid.SteamID(friend.GetSteamidSource()),
			GameDataBlob:           friend.GetGameDataBlob(),
			Name:                   friend.GetPlayerName(),
			Avatar:                 hex.EncodeToString(friend.GetAvatarHash()),
			LastLogOff:             friend.GetLastLogoff(),
			LastLogOn:              friend.GetLastLogon(),
			ClanRank:               friend.GetClanRank(),
			ClanTag:                friend.GetClanTag(),
			OnlineSessionInstances: friend.GetOnlineSessionInstances(),
			PersonaSetByUser:       friend.GetPersonaSetByUser(),
			// PublishedSessionID:     friend.GetPublishedInstanceId(),
			// FacebookName:           friend.GetFacebookName(),
			// FacebookID:             friend.GetFacebookId(),
		})
	}
}

func (s *Social) handleClanState(packet *protocol.Packet) {
	body := &pb.CMsgClientClanState{}
	packet.ReadProtoMsg(body)

	var name, avatar string

	if body.GetNameInfo() != nil {
		name = body.GetNameInfo().GetClanName()
		avatar = hex.EncodeToString(body.GetNameInfo().GetShaAvatar())
	}

	var totalCount, onlineCount, chattingCount, ingameCount uint32

	if body.GetUserCounts() != nil {
		usercounts := body.GetUserCounts()
		totalCount = usercounts.GetMembers()
		onlineCount = usercounts.GetOnline()
		chattingCount = usercounts.GetChatting()
		ingameCount = usercounts.GetInGame()
	}

	var events, announcements []ClanEventDetails

	for _, event := range body.GetEvents() {
		events = append(events, ClanEventDetails{
			ID:         event.GetGid(),
			EventTime:  event.GetEventTime(),
			Headline:   event.GetHeadline(),
			GameID:     event.GetGameId(),
			JustPosted: event.GetJustPosted(),
		})
	}

	for _, announce := range body.GetAnnouncements() {
		announcements = append(announcements, ClanEventDetails{
			ID:         announce.GetGid(),
			EventTime:  announce.GetEventTime(),
			Headline:   announce.GetHeadline(),
			GameID:     announce.GetGameId(),
			JustPosted: announce.GetJustPosted(),
		})
	}

	clanid := steamid.SteamID(body.GetSteamidClan())

	// TODO: investigate what changed and re-enable this
	// flags := steamlang.EClientPersonaStateFlag(body.GetMUnStatusFlags())

	/*
		if (flags & steamlang.EClientPersonaStateFlag_PlayerName) == steamlang.EClientPersonaStateFlag_PlayerName {
			if name != "" {
				s.Groups.SetName(clanid, name)
			}
		}

		if (flags & steamlang.EClientPersonaStateFlag_Presence) == steamlang.EClientPersonaStateFlag_Presence {
			if protocol.ValidAvatar(avatar) {
				s.Groups.SetAvatar(clanid, avatar)
			}
		}
	*/

	if body.GetUserCounts() != nil {
		s.Groups.SetMemberTotalCount(clanid, totalCount)
		s.Groups.SetMemberOnlineCount(clanid, onlineCount)
		s.Groups.SetMemberChattingCount(clanid, chattingCount)
		s.Groups.SetMemberInGameCount(clanid, ingameCount)
	}

	s.client.Emit(&ClanStateEvent{
		// StateFlags:          steamlang.EClientPersonaStateFlag(body.GetMUnStatusFlags()),
		AccountFlags:        steamlang.EAccountFlags(body.GetClanAccountFlags()),
		ClandID:             clanid,
		ClanName:            name,
		Avatar:              avatar,
		MemberTotalCount:    totalCount,
		MemberOnlineCount:   onlineCount,
		MemberChattingCount: chattingCount,
		MemberInGameCount:   ingameCount,
		Events:              events,
		Announcements:       announcements,
	})
}

func (s *Social) handleFriendResponse(packet *protocol.Packet) {
	body := &pb.CMsgClientAddFriendResponse{}

	packet.ReadProtoMsg(body)

	s.client.Emit(&FriendAddedEvent{
		Result:      steamlang.EResult(body.GetEresult()),
		SteamID:     steamid.SteamID(body.GetSteamIdAdded()),
		PersonaName: body.GetPersonaNameAdded(),
	})
}

func (s *Social) handleFriendMsg(packet *protocol.Packet) {
	body := &pb.CMsgClientFriendMsgIncoming{}

	packet.ReadProtoMsg(body)

	message := string(bytes.Split(body.GetMessage(), []byte{0x0})[0])

	s.client.Emit(&ChatMsgEvent{
		ChatterID: steamid.SteamID(body.GetSteamidFrom()),
		Message:   message,
		EntryType: steamlang.EChatEntryType(body.GetChatEntryType()),
		Timestamp: time.Unix(int64(body.GetRtime32ServerTimestamp()), 0),
	})
}

func (s *Social) handleChatMsg(packet *protocol.Packet) {
	body := &steamlang.MsgClientChatMsg{}
	payload := packet.ReadClientMsg(body).Payload
	message := string(bytes.Split(payload, []byte{0x0})[0])

	s.client.Emit(&ChatMsgEvent{
		ChatRoomID: steamid.SteamID(body.SteamIdChatRoom),
		ChatterID:  steamid.SteamID(body.SteamIdChatter),
		Message:    message,
		EntryType:  body.ChatMsgType,
	})
}

func (s *Social) handleChatEnter(packet *protocol.Packet) {
	body := &steamlang.MsgClientChatEnter{}
	payload := packet.ReadClientMsg(body).Payload
	reader := bytes.NewBuffer(payload)
	name, _ := rwu.ReadString(reader)
	_, _ = rwu.ReadByte(reader) //0
	count := body.NumMembers
	chatID := steamid.SteamID(body.SteamIdChat)
	clanID := steamid.SteamID(body.SteamIdClan)

	s.Chats.Add(socialcache.Chat{SteamID: chatID, GroupID: clanID})

	for i := 0; i < int(count); i++ {
		id, chatPerm, clanPerm := readChatMember(reader)
		_, _ = rwu.ReadBytes(reader, 6) //No idea what this is

		s.Chats.AddChatMember(chatID, socialcache.ChatMember{
			SteamID:         id,
			ChatPermissions: chatPerm,
			ClanPermissions: clanPerm,
		})
	}

	s.client.Emit(&ChatEnterEvent{
		ChatRoomID:    steamid.SteamID(body.SteamIdChat),
		FriendID:      steamid.SteamID(body.SteamIdFriend),
		ChatRoomType:  body.ChatRoomType,
		OwnerID:       steamid.SteamID(body.SteamIdOwner),
		ClanID:        steamid.SteamID(body.SteamIdClan),
		ChatFlags:     body.ChatFlags,
		EnterResponse: body.EnterResponse,
		Name:          name,
	})
}

func (s *Social) handleChatMemberInfo(packet *protocol.Packet) {
	body := &steamlang.MsgClientChatMemberInfo{}
	payload := packet.ReadClientMsg(body).Payload
	reader := bytes.NewBuffer(payload)
	chatID := steamid.SteamID(body.SteamIdChat)

	if body.Type == steamlang.EChatInfoType_StateChange {
		actedOn, _ := rwu.ReadUint64(reader)
		state, _ := rwu.ReadInt32(reader)
		actedBy, _ := rwu.ReadUint64(reader)
		_, _ = rwu.ReadByte(reader) // 0
		stateChange := steamlang.EChatMemberStateChange(state)

		switch stateChange {
		case steamlang.EChatMemberStateChange_Entered:
			_, chatPerm, clanPerm := readChatMember(reader)

			s.Chats.AddChatMember(chatID, socialcache.ChatMember{
				SteamID:         steamid.SteamID(actedOn),
				ChatPermissions: chatPerm,
				ClanPermissions: clanPerm,
			})
		case steamlang.EChatMemberStateChange_Banned,
			steamlang.EChatMemberStateChange_Kicked,
			steamlang.EChatMemberStateChange_Disconnected,
			steamlang.EChatMemberStateChange_Left:
			s.Chats.RemoveChatMember(chatID, steamid.SteamID(actedOn))
		}

		stateInfo := StateChangeDetails{
			ChatterActedOn: steamid.SteamID(actedOn),
			StateChange:    stateChange,
			ChatterActedBy: steamid.SteamID(actedBy),
		}

		s.client.Emit(&ChatMemberInfoEvent{
			ChatRoomID:      steamid.SteamID(body.SteamIdChat),
			Type:            body.Type,
			StateChangeInfo: stateInfo,
		})
	}
}

// TODO: handle errors
func readChatMember(r io.Reader) (steamid.SteamID, steamlang.EChatPermission, steamlang.EClanPermission) {
	_, _ = rwu.ReadString(r) // MessageObject
	_, _ = rwu.ReadByte(r)   // 7
	_, _ = rwu.ReadString(r) //steamid
	id, _ := rwu.ReadUint64(r)
	_, _ = rwu.ReadByte(r)   // 2
	_, _ = rwu.ReadString(r) //Permissions
	chat, _ := rwu.ReadInt32(r)
	_, _ = rwu.ReadByte(r)   // 2
	_, _ = rwu.ReadString(r) //Details
	clan, _ := rwu.ReadInt32(r)

	return steamid.SteamID(id), steamlang.EChatPermission(chat), steamlang.EClanPermission(clan)
}

func (s *Social) handleChatActionResult(packet *protocol.Packet) {
	body := &steamlang.MsgClientChatActionResult{}

	packet.ReadClientMsg(body)

	s.client.Emit(&ChatActionResultEvent{
		ChatRoomID: steamid.SteamID(body.SteamIdChat),
		ChatterID:  steamid.SteamID(body.SteamIdUserActedOn),
		Action:     body.ChatAction,
		Result:     body.ActionResult,
	})
}

func (s *Social) handleChatInvite(packet *protocol.Packet) {
	body := &pb.CMsgClientChatInvite{}

	packet.ReadProtoMsg(body)

	s.client.Emit(&ChatInviteEvent{
		InvitedID:    steamid.SteamID(body.GetSteamIdInvited()),
		ChatRoomID:   steamid.SteamID(body.GetSteamIdChat()),
		PatronID:     steamid.SteamID(body.GetSteamIdPatron()),
		ChatRoomType: steamlang.EChatRoomType(body.GetChatroomType()),
		FriendChatID: steamid.SteamID(body.GetSteamIdFriendChat()),
		ChatRoomName: body.GetChatName(),
		GameID:       body.GetGameId(),
	})
}

func (s *Social) handleIgnoreFriendResponse(packet *protocol.Packet) {
	body := &steamlang.MsgClientSetIgnoreFriendResponse{}

	packet.ReadClientMsg(body)

	s.client.Emit(&IgnoreFriendEvent{
		Result: body.Result,
	})
}

func (s *Social) handleProfileInfoResponse(packet *protocol.Packet) {
	body := &pb.CMsgClientFriendProfileInfoResponse{}

	packet.ReadProtoMsg(body)

	s.client.Emit(&ProfileInfoEvent{
		Result:      steamlang.EResult(body.GetEresult()),
		SteamID:     steamid.SteamID(body.GetSteamidFriend()),
		TimeCreated: body.GetTimeCreated(),
		RealName:    body.GetRealName(),
		CityName:    body.GetCityName(),
		StateName:   body.GetStateName(),
		CountryName: body.GetCountryName(),
		Headline:    body.GetHeadline(),
		Summary:     body.GetSummary(),
	})
}
