package pulsesms

import (
	"fmt"
	"sync"
	"time"
)

type Store struct {
	sync.Mutex
	Contacts map[PID]Contact
	Chats    map[PID]Chat
}

type Contact struct {
	PID    PID
	Notify string
	Name   string
	Short  string
}

type Chat struct {
	PID             PID
	ConversationID  ConversationID
	Name            string
	ModifyTag       string
	UnreadCount     int
	LastMessageTime int64
	MutedUntil      int64
	IsMarkedSpam    bool
	IsArchived      bool
	IsPinned        bool
	Members         []PhoneNumber
	// Source          map[string]string
	ReceivedAt time.Time
}

func newChat(conv Conversation) Chat {
	c := Chat{
		// PID: fmt.Sprint(conv.DeviceId),
		PID: fmt.Sprint(conv.ID),

		ConversationID: conv.ID,

		Name:            conv.Title,
		Members:         conv.members(),
		LastMessageTime: conv.Timestamp,
	}
	return c
}

func newStore() *Store {
	return &Store{
		Contacts: make(map[PID]Contact),
		Chats:    make(map[PID]Chat),
	}
}

func (s *Store) setContact(phone PhoneNumber, contact Contact) {
	// fmt.Println("setting contact", phone, contact.PID, contact.Name)
	s.Lock()
	s.Contacts[phone] = contact
	s.Unlock()
}

func (s *Store) getContactByPhone(phone PhoneNumber) (Contact, bool) {
	c, ok := s.Contacts[phone]
	return c, ok
}

func (s *Store) getContactByName(name string) (Contact, bool) {
	for _, c := range s.Contacts {
		if c.Name == name {
			return c, true
		}
	}
	return Contact{}, false
}

func (s *Store) getChatByConversationID(convoID ConversationID) (Chat, bool) {
	fmt.Println("gettin chat by convo", convoID)
	for _, c := range s.Chats {
		if c.ConversationID == convoID {
			fmt.Println("match name:", c.Name)
			fmt.Println("match pid:", c.PID)
			return c, true
		}
	}
	return Chat{}, false
}

func (s *Store) SetConversation(convo Conversation) {
	chat := newChat(convo)
	s.setChat(chat)
}

// func (s *Store) GetChatFromMessage(m Message) (Chat, bool) {
// 	return s.GetChatByConversationID(m.ConversationID)

// }

func (s *Store) setChat(chat Chat) {
	s.Lock()

	// fmt.Println("setting chat")
	// fmt.Printf("name: %s, pid: %s, convoID: %d\n", chat.Name, chat.PID, chat.ConversationID)
	if chat.PID != "" {
		s.Chats[chat.PID] = chat
	}
	s.Unlock()

	// dm
	if len(chat.Members) == 1 {
		m := chat.Members[0]
		contact := Contact{PID: m, Name: chat.Name}
		s.setContact(m, contact)
		return
	}

	for _, m := range chat.Members {
		_, ok := s.Contacts[m]
		if !ok {
			noname := Contact{PID: m, Name: m}
			s.setContact(m, noname)
		}

	}
}
