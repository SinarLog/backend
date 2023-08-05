package vo

type OpenChatRequest struct {
	RoomId      string `json:"roomId"`
	SenderId    string `json:"sender" binding:"required"`
	RecipientId string `json:"recipient"`
}
