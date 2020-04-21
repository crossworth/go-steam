package steam

import (
	"time"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

type FriendsListEvent struct{}

type FriendStateEvent struct {
	SteamID      steamid.SteamID `json:",string"`
	Relationship steamlang.EFriendRelationship
}

func (f *FriendStateEvent) IsFriend() bool {
	return f.Relationship == steamlang.EFriendRelationship_Friend
}

type GroupStateEvent struct {
	SteamID      steamid.SteamID `json:",string"`
	Relationship steamlang.EClanRelationship
}

func (g *GroupStateEvent) IsMember() bool {
	return g.Relationship == steamlang.EClanRelationship_Member
}

// Fired when someone changing their friend details
type PersonaStateEvent struct {
	StatusFlags            steamlang.EClientPersonaStateFlag
	FriendID               steamid.SteamID `json:",string"`
	State                  steamlang.EPersonaState
	StateFlags             steamlang.EPersonaStateFlag
	GameAppID              uint32
	GameID                 uint64 `json:",string"`
	GameName               string
	GameServerIP           uint32
	GameServerPort         uint32
	QueryPort              uint32
	SourceSteamID          steamid.SteamID `json:",string"`
	GameDataBlob           []byte
	Name                   string
	Avatar                 string
	LastLogOff             uint32
	LastLogOn              uint32
	ClanRank               uint32
	ClanTag                string
	OnlineSessionInstances uint32
	PublishedSessionID     uint32
	PersonaSetByUser       bool
	FacebookName           string
	FacebookID             uint64 `json:",string"`
}

// Fired when a clan's state has been changed
type ClanStateEvent struct {
	ClandID             steamid.SteamID `json:",string"`
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
	ID         uint64 `json:",string"`
	EventTime  uint32
	Headline   string
	GameID     uint64 `json:",string"`
	JustPosted bool
}

// Fired in response to adding a friend to your friends list
type FriendAddedEvent struct {
	Result      steamlang.EResult
	SteamID     steamid.SteamID `json:",string"`
	PersonaName string
}

// Fired when the client receives a message from either a friend or a chat room
type ChatMsgEvent struct {
	ChatRoomID steamid.SteamID `json:",string"` // not set for friend messages
	ChatterID  steamid.SteamID `json:",string"`
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
	ChatRoomID    steamid.SteamID `json:",string"`
	FriendID      steamid.SteamID `json:",string"`
	ChatRoomType  steamlang.EChatRoomType
	OwnerID       steamid.SteamID `json:",string"`
	ClanID        steamid.SteamID `json:",string"`
	ChatFlags     byte
	EnterResponse steamlang.EChatRoomEnterResponse
	Name          string
}

// Fired in response to a chat member's info being received
type ChatMemberInfoEvent struct {
	ChatRoomID      steamid.SteamID `json:",string"`
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
	ChatRoomID steamid.SteamID `json:",string"`
	ChatterID  steamid.SteamID `json:",string"`
	Action     steamlang.EChatAction
	Result     steamlang.EChatActionResult
}

// Fired when a chat invite is received
type ChatInviteEvent struct {
	InvitedID    steamid.SteamID `json:",string"`
	ChatRoomID   steamid.SteamID `json:",string"`
	PatronID     steamid.SteamID `json:",string"`
	ChatRoomType steamlang.EChatRoomType
	FriendChatID steamid.SteamID `json:",string"`
	ChatRoomName string
	GameID       uint64 `json:",string"`
}

// Fired in response to ignoring a friend
type IgnoreFriendEvent struct {
	Result steamlang.EResult
}

// Fired in response to requesting profile info for a user
type ProfileInfoEvent struct {
	Result      steamlang.EResult
	SteamID     steamid.SteamID `json:",string"`
	TimeCreated uint32
	RealName    string
	CityName    string
	StateName   string
	CountryName string
	Headline    string
	Summary     string
}
