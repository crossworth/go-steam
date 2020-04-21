package steam

import (
	"time"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

type FriendsListEvent struct{}

type FriendStateEvent struct {
	SteamId      steamid.SteamID `json:",string"`
	Relationship steamlang.EFriendRelationship
}

func (f *FriendStateEvent) IsFriend() bool {
	return f.Relationship == steamlang.EFriendRelationship_Friend
}

type GroupStateEvent struct {
	SteamId      steamid.SteamID `json:",string"`
	Relationship steamlang.EClanRelationship
}

func (g *GroupStateEvent) IsMember() bool {
	return g.Relationship == steamlang.EClanRelationship_Member
}

// Fired when someone changing their friend details
type PersonaStateEvent struct {
	StatusFlags            steamlang.EClientPersonaStateFlag
	FriendId               steamid.SteamID `json:",string"`
	State                  steamlang.EPersonaState
	StateFlags             steamlang.EPersonaStateFlag
	GameAppId              uint32
	GameId                 uint64 `json:",string"`
	GameName               string
	GameServerIp           uint32
	GameServerPort         uint32
	QueryPort              uint32
	SourceSteamId          steamid.SteamID `json:",string"`
	GameDataBlob           []byte
	Name                   string
	Avatar                 string
	LastLogOff             uint32
	LastLogOn              uint32
	ClanRank               uint32
	ClanTag                string
	OnlineSessionInstances uint32
	PublishedSessionId     uint32
	PersonaSetByUser       bool
	FacebookName           string
	FacebookId             uint64 `json:",string"`
}

// Fired when a clan's state has been changed
type ClanStateEvent struct {
	ClandId             steamid.SteamID `json:",string"`
	StateFlags          steamlang.EClientPersonaStateFlag
	AccountFlags        steamlang.EAccountFlags
	ClanName            string
	Avatar              string
	MemberTotalCount    uint32
	MemberOnlineCount   uint32
	MemberChattingCount uint32
	MemberInGameCount   uint32
	Events              []ClanEventDetails
	Announcements       []ClanEventDetails
}

type ClanEventDetails struct {
	Id         uint64 `json:",string"`
	EventTime  uint32
	Headline   string
	GameId     uint64 `json:",string"`
	JustPosted bool
}

// Fired in response to adding a friend to your friends list
type FriendAddedEvent struct {
	Result      steamlang.EResult
	SteamId     steamid.SteamID `json:",string"`
	PersonaName string
}

// Fired when the client receives a message from either a friend or a chat room
type ChatMsgEvent struct {
	ChatRoomId steamid.SteamID `json:",string"` // not set for friend messages
	ChatterId  steamid.SteamID `json:",string"`
	Message    string
	EntryType  steamlang.EChatEntryType
	Timestamp  time.Time
	Offline    bool
}

// Whether the type is ChatMsg
func (c *ChatMsgEvent) IsMessage() bool {
	return c.EntryType == steamlang.EChatEntryType_ChatMsg
}

// Fired in response to joining a chat
type ChatEnterEvent struct {
	ChatRoomId    steamid.SteamID `json:",string"`
	FriendId      steamid.SteamID `json:",string"`
	ChatRoomType  steamlang.EChatRoomType
	OwnerId       steamid.SteamID `json:",string"`
	ClanId        steamid.SteamID `json:",string"`
	ChatFlags     byte
	EnterResponse steamlang.EChatRoomEnterResponse
	Name          string
}

// Fired in response to a chat member's info being received
type ChatMemberInfoEvent struct {
	ChatRoomId      steamid.SteamID `json:",string"`
	Type            steamlang.EChatInfoType
	StateChangeInfo StateChangeDetails
}

type StateChangeDetails struct {
	ChatterActedOn steamid.SteamID `json:",string"`
	StateChange    steamlang.EChatMemberStateChange
	ChatterActedBy steamid.SteamID `json:",string"`
}

// Fired when a chat action has completed
type ChatActionResultEvent struct {
	ChatRoomId steamid.SteamID `json:",string"`
	ChatterId  steamid.SteamID `json:",string"`
	Action     steamlang.EChatAction
	Result     steamlang.EChatActionResult
}

// Fired when a chat invite is received
type ChatInviteEvent struct {
	InvitedId    steamid.SteamID `json:",string"`
	ChatRoomId   steamid.SteamID `json:",string"`
	PatronId     steamid.SteamID `json:",string"`
	ChatRoomType steamlang.EChatRoomType
	FriendChatId steamid.SteamID `json:",string"`
	ChatRoomName string
	GameId       uint64 `json:",string"`
}

// Fired in response to ignoring a friend
type IgnoreFriendEvent struct {
	Result steamlang.EResult
}

// Fired in response to requesting profile info for a user
type ProfileInfoEvent struct {
	Result      steamlang.EResult
	SteamId     steamid.SteamID `json:",string"`
	TimeCreated uint32
	RealName    string
	CityName    string
	StateName   string
	CountryName string
	Headline    string
	Summary     string
}
