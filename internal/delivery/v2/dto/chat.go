package dto

type OpenChatResponse struct {
	Room  RoomResponse   `json:"room"`
	Chats []ChatResponse `json:"chats"`
}

type RoomResponse struct {
	Id           string   `json:"id"`
	Participants []string `json:"participants"`
	CreatedAt    string   `json:"createdAt"`
}

type ChatResponse struct {
	Id        string `json:"id,omitempty"`
	RoomId    string `json:"roomId,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Message   string `json:"message,omitempty"`
	Read      bool   `json:"read"`
	SentAt    string `json:"sentAt,omitempty"`
	Timestamp uint32 `json:"timestamp,omitempty"`
}
