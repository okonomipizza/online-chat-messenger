package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
)

// UDP接続に利用するプロトコル
type ChatMessage struct {
	ChatRoomID string `json:"room_id"`
	UserID     string `json:"user_id"`
	Message    string `json:"message"`
}

func (chat ChatMessage) CreateChatRequest() ([]byte, error) {
	buf := new(bytes.Buffer)

	chatRoomIDBytes := []byte(chat.ChatRoomID)
	userIDBytes := []byte(chat.UserID)
	messageBytes := []byte(chat.Message)

	// chatroom size
	if err := buf.WriteByte(byte(len(chatRoomIDBytes))); err != nil {
		return nil, err
	}

	// user id size
	if err := buf.WriteByte(byte(len(userIDBytes))); err != nil {
		return nil, err
	}

	// message size
	if err := buf.WriteByte(byte(len(messageBytes))); err != nil {
		return nil, err
	}

	payload := slices.Concat(chatRoomIDBytes, userIDBytes, messageBytes)
	if _, err := buf.Write(payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ParseChatRequest(message []byte) (ChatMessage, bool, error) {
	if len(message) < 2 {
		return ChatMessage{}, false, errors.New("Invalid message length")
	}

	chatRoomIDSize := int(message[0])
	userIDSize := int(message[1])
	payloadSize := int(message[2])
	payload := message[3:]

	chatRoomID := string(payload[:chatRoomIDSize])
	payload = payload[chatRoomIDSize:]

	userID := string(payload[:userIDSize])
	payload = payload[userIDSize:]
	ChatMessageBytes := payload[:payloadSize]
	println("message payload length: ", len(ChatMessageBytes))
	chatMessage := string(payload[:payloadSize])

	if len(ChatMessageBytes) == 0 {
		return ChatMessage{
			ChatRoomID: chatRoomID,
			UserID:     userID,
		}, true, nil
	}

	fmt.Printf("chat room id: %s\n", chatRoomID)
	fmt.Printf("user id: %s\n", userID)
	fmt.Printf("message: %s\n", chatMessage)

	return ChatMessage{
		ChatRoomID: chatRoomID,
		UserID:     userID,
		Message:    chatMessage,
	}, false, nil
}
