package chat

import "time"

type User struct {
	ID           string `json:"_id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email,omitempty"`
	ProfileImage string `json:"profileImage,omitempty"`
	AccountType  string `json:"accountType,omitempty"`
}
type Job struct {
	ID       string `json:"_id"`
	JobTitle string `json:"jobTitle"`
	Status   string `json:"status"`
}
type Message struct {
	ID             string     `json:"_id"`
	ConversationID string     `json:"conversationId"`
	Sender         any        `json:"senderId"`
	Receiver       any        `json:"receiverId"`
	MessageType    string     `json:"messageType"`
	Content        string     `json:"content,omitempty"`
	FileURL        string     `json:"fileUrl,omitempty"`
	FileName       string     `json:"fileName,omitempty"`
	FileSize       *int64     `json:"fileSize,omitempty"`
	IsRead         bool       `json:"isRead"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
	IsDelivered    bool       `json:"isDelivered"`
	DeliveredAt    time.Time  `json:"deliveredAt"`
	IsDeleted      bool       `json:"isDeleted"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
type Conversation struct {
	ID             string         `json:"_id"`
	Participants   []User         `json:"participants"`
	Job            Job            `json:"jobId"`
	LastMessage    *Message       `json:"lastMessage,omitempty"`
	LastMessageAt  time.Time      `json:"lastMessageAt"`
	IsArchived     bool           `json:"isArchived"`
	IsActive       bool           `json:"isActive"`
	IsUrgent       bool           `json:"isUrgent"`
	MarkedUrgentBy *string        `json:"markedUrgentBy,omitempty"`
	UnreadCount    map[string]int `json:"unreadCount"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}
type Page[T any] struct {
	Docs          []T   `json:"docs"`
	TotalDocs     int64 `json:"totalDocs"`
	Limit         int   `json:"limit"`
	TotalPages    int   `json:"totalPages"`
	Page          int   `json:"page"`
	PagingCounter int   `json:"pagingCounter"`
	HasPrevPage   bool  `json:"hasPrevPage"`
	HasNextPage   bool  `json:"hasNextPage"`
	PrevPage      *int  `json:"prevPage"`
	NextPage      *int  `json:"nextPage"`
}
type ListInput struct {
	Page, Limit int
	Archived    *bool
	Active      *bool
	Query       string
}
type SendInput struct {
	Content, MessageType, FileURL, FileName string
	FileSize                                *int64
}
type Event struct {
	Type           string   `json:"type"`
	ConversationID string   `json:"conversationId"`
	Data           any      `json:"data,omitempty"`
	UserIDs        []string `json:"-"`
}
