package pulsesms

import "time"

type Store struct {
	Contacts map[PID]Contact
	Chats    map[ConversationID]Chat
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

func newStore() *Store {
	return &Store{
		Contacts: make(map[PID]Contact),
		Chats:    make(map[ConversationID]Chat),
	}
}

func (s *Store) setContact(phone PhoneNumber, contact Contact) {
	s.Contacts[phone] = contact
}

func (s *Store) GetContactByPhone(phone PhoneNumber) (Contact, bool) {
	c, ok := s.Contacts[phone]
	return c, ok
}

func (s *Store) GetContactByName(name string) (Contact, bool) {
	for _, c := range s.Contacts {
		if c.Name == name {
			return c, true
		}
	}
	return Contact{}, false
}

func (s *Store) setChat(chat Chat) {
	s.Chats[chat.ConversationID] = chat

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

func (s *Store) SetConversation(chat Chat) {
	s.Chats[chat.ConversationID] = chat
}
