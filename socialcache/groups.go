package socialcache

import (
	"errors"
	"sync"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

// Groups list is a thread safe map
// They can be iterated over like so:
// 	for id, group := range client.Social.Groups.GetCopy() {
// 		log.Println(id, group.Name)
// 	}
type GroupsList struct {
	mutex sync.RWMutex
	byID  map[steamid.SteamID]*Group
}

// Returns a new groups list
func NewGroupsList() *GroupsList {
	return &GroupsList{byID: make(map[steamid.SteamID]*Group)}
}

// Adds a group to the group list
func (list *GroupsList) Add(group Group) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	_, exists := list.byID[group.SteamID]
	if !exists { //make sure this doesnt already exist
		list.byID[group.SteamID] = &group
	}
}

// Remove removes a group from the group list
func (list *GroupsList) Remove(id steamid.SteamID) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byID, id)
}

// GetCopy returns a copy of the groups map
func (list *GroupsList) GetCopy() map[steamid.SteamID]Group {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	glist := make(map[steamid.SteamID]Group)

	for key, group := range list.byID {
		glist[key] = *group
	}

	return glist
}

// Returns a copy of the group of a given SteamID
func (list *GroupsList) ByID(id steamid.SteamID) (Group, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		return *val, nil
	}
	return Group{}, errors.New("Group not found")
}

// Returns the number of groups
func (list *GroupsList) Count() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.byID)
}

//Setter methods
func (list *GroupsList) SetName(id steamid.SteamID, name string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.Name = name
	}
}

func (list *GroupsList) SetAvatar(id steamid.SteamID, hash string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.Avatar = hash
	}
}

func (list *GroupsList) SetRelationship(id steamid.SteamID, relationship steamlang.EClanRelationship) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.Relationship = relationship
	}
}

func (list *GroupsList) SetMemberTotalCount(id steamid.SteamID, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.MemberTotalCount = count
	}
}

func (list *GroupsList) SetMemberOnlineCount(id steamid.SteamID, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.MemberOnlineCount = count
	}
}

func (list *GroupsList) SetMemberChattingCount(id steamid.SteamID, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.MemberChattingCount = count
	}
}

func (list *GroupsList) SetMemberInGameCount(id steamid.SteamID, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	// id = id.ChatToClan()
	if val, ok := list.byID[id]; ok {
		val.MemberInGameCount = count
	}
}

// A Group
type Group struct {
	SteamID             steamid.SteamID `json:",string"`
	Name                string
	Avatar              string
	Relationship        steamlang.EClanRelationship
	MemberTotalCount    uint32
	MemberOnlineCount   uint32
	MemberChattingCount uint32
	MemberInGameCount   uint32
}
