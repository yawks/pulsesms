package pulsesms

import (
	"fmt"
	"sync"
	"time"
)

type ChatID = string

type Store struct {
	sync.Mutex
	Contacts map[PID]Contact
	Chats    map[ChatID]Chat
}

type Contact struct {
	PID    PID
	Notify string
	Name   string
	Short  string
}

type Chat struct {
	// PID             PID
	ID              ChatID
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

func newChat(conv conversation) Chat {
	id := fmt.Sprint(conv.DeviceId)
	c := Chat{
		ID:              id,
		Name:            conv.Title,
		Members:         conv.members(),
		LastMessageTime: conv.Timestamp,
	}
	return c
}

func newStore() *Store {
	return &Store{
		Contacts: make(map[PID]Contact),
		Chats:    make(map[ChatID]Chat),
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

func (s *Store) setConversation(convo conversation) {
	chat := newChat(convo)
	s.setChat(chat)
}

func (s *Store) setChat(chat Chat) {
	s.Lock()
	if chat.ID != "" && chat.ID != "0" {
		s.Chats[chat.ID] = chat
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
