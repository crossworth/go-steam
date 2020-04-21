package socialcache

import (
	"errors"
	"sync"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

// FriendsList is a thread safe map
// They can be iterated over like so:
// 	for id, friend := range client.Social.Friends.GetCopy() {
// 		log.Println(id, friend.Name)
// 	}
type FriendsList struct {
	mutex sync.RWMutex
	byID  map[steamid.SteamID]*Friend
}

// NewFriendsList builds a new friends list
func NewFriendsList() *FriendsList {
	return &FriendsList{byID: make(map[steamid.SteamID]*Friend)}
}

// Add adds a friend to the friend list
func (list *FriendsList) Add(friend Friend) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	_, exists := list.byID[friend.SteamID]
	if !exists { //make sure this doesnt already exist
		list.byID[friend.SteamID] = &friend
	}
}

// Remove removes a friend from the friend list
func (list *FriendsList) Remove(id steamid.SteamID) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byID, id)
}

// Returns a copy of the friends map
func (list *FriendsList) GetCopy() map[steamid.SteamID]Friend {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	flist := make(map[steamid.SteamID]Friend)
	for key, friend := range list.byID {
		flist[key] = *friend
	}
	return flist
}

// Returns a copy of the friend of a given SteamID
func (list *FriendsList) ByID(id steamid.SteamID) (Friend, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if val, ok := list.byID[id]; ok {
		return *val, nil
	}
	return Friend{}, errors.New("Friend not found")
}

// Returns the number of friends
func (list *FriendsList) Count() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.byID)
}

//Setter methods
func (list *FriendsList) SetName(id steamid.SteamID, name string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.Name = name
	}
}

func (list *FriendsList) SetAvatar(id steamid.SteamID, hash string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.Avatar = hash
	}
}

func (list *FriendsList) SetRelationship(id steamid.SteamID, relationship steamlang.EFriendRelationship) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.Relationship = relationship
	}
}

func (list *FriendsList) SetPersonaState(id steamid.SteamID, state steamlang.EPersonaState) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.PersonaState = state
	}
}

func (list *FriendsList) SetPersonaStateFlags(id steamid.SteamID, flags steamlang.EPersonaStateFlag) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.PersonaStateFlags = flags
	}
}

func (list *FriendsList) SetGameAppID(id steamid.SteamID, gameAppID uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.GameAppID = gameAppID
	}
}

func (list *FriendsList) SetGameID(id steamid.SteamID, gameID uint64) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.GameID = gameID
	}
}

func (list *FriendsList) SetGameName(id steamid.SteamID, name string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byID[id]; ok {
		val.GameName = name
	}
}

// A Friend
type Friend struct {
	SteamID           steamid.SteamID `json:",string"`
	Name              string
	Avatar            string
	Relationship      steamlang.EFriendRelationship
	PersonaState      steamlang.EPersonaState
	PersonaStateFlags steamlang.EPersonaStateFlag
	GameAppID         uint32
	GameID            uint64 `json:",string"`
	GameName          string
}
