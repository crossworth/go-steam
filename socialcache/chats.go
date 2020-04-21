package socialcache

import (
	"errors"
	"sync"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

// Chats list is a thread safe map
// They can be iterated over like so:
// 	for id, chat := range client.Social.Chats.GetCopy() {
// 		log.Println(id, chat.Name)
// 	}
type ChatsList struct {
	mutex sync.RWMutex
	byID  map[steamid.SteamID]*Chat
}

// Returns a new chats list
func NewChatsList() *ChatsList {
	return &ChatsList{byID: make(map[steamid.SteamID]*Chat)}
}

// Adds a chat to the chat list
func (list *ChatsList) Add(chat Chat) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	_, exists := list.byID[chat.SteamID]
	if !exists { //make sure this doesnt already exist
		list.byID[chat.SteamID] = &chat
	}
}

// Removes a chat from the chat list
func (list *ChatsList) Remove(id steamid.SteamID) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byID, id)
}

// Adds a chat member to a given chat
func (list *ChatsList) AddChatMember(id steamid.SteamID, member ChatMember) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	chat := list.byID[id]
	if chat == nil { //Chat doesn't exist
		chat = &Chat{SteamID: id}
		list.byID[id] = chat
	}
	if chat.ChatMembers == nil { //New chat
		chat.ChatMembers = make(map[steamid.SteamID]ChatMember)
	}
	chat.ChatMembers[member.SteamID] = member
}

// Removes a chat member from a given chat
func (list *ChatsList) RemoveChatMember(id steamid.SteamID, member steamid.SteamID) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	chat := list.byID[id]
	if chat == nil { //Chat doesn't exist
		return
	}
	if chat.ChatMembers == nil { //New chat
		return
	}
	delete(chat.ChatMembers, member)
}

// Returns a copy of the chats map
func (list *ChatsList) GetCopy() map[steamid.SteamID]Chat {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	glist := make(map[steamid.SteamID]Chat)
	for key, chat := range list.byID {
		glist[key] = *chat
	}
	return glist
}

// Returns a copy of the chat of a given steamid.SteamID
func (list *ChatsList) ByID(id steamid.SteamID) (Chat, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if val, ok := list.byID[id]; ok {
		return *val, nil
	}
	return Chat{}, errors.New("Chat not found")
}

// Returns the number of chats
func (list *ChatsList) Count() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.byID)
}

// A Chat
type Chat struct {
	SteamID     steamid.SteamID `json:",string"`
	GroupID     steamid.SteamID `json:",string"`
	ChatMembers map[steamid.SteamID]ChatMember
}

// A Chat Member
type ChatMember struct {
	SteamID         steamid.SteamID `json:",string"`
	ChatPermissions steamlang.EChatPermission
	ClanPermissions steamlang.EClanPermission
}
