package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
)

// UDP接続に利用するプロトコル
type ChatMessage struct {
	Operation  byte
	ChatRoomID string
	UserID     string
	Message    string
}

const (
	ChatOperationSendMessage byte = iota
	ChatOperationSendUDPAddr
	ChatOperationExit
)

func (chat ChatMessage) CreateChatRequest(operation byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	chatRoomIDBytes := []byte(chat.ChatRoomID)
	userIDBytes := []byte(chat.UserID)
	messageBytes := []byte(chat.Message)

	// operation
	if err := buf.WriteByte(operation); err != nil {
		return nil, err
	}

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

func ParseChatRequest(message []byte) (ChatMessage, error) {
	if len(message) < 2 {
		return ChatMessage{}, errors.New("Invalid message length")
	}

	operation := message[0]
	chatRoomIDSize := int(message[1])
	userIDSize := int(message[2])
	payloadSize := int(message[3])
	payload := message[4:]

	chatRoomID := string(payload[:chatRoomIDSize])
	payload = payload[chatRoomIDSize:]

	userID := string(payload[:userIDSize])
	payload = payload[userIDSize:]

	chatMessage := string(payload[:payloadSize])

	fmt.Printf("chat room id: %s\n", chatRoomID)
	fmt.Printf("user id: %s\n", userID)
	fmt.Printf("message: %s\n", chatMessage)

	return ChatMessage{
		Operation:  operation,
		ChatRoomID: chatRoomID,
		UserID:     userID,
		Message:    chatMessage,
	}, nil
}
