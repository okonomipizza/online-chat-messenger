package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
)

// ChatMessageはクライアント・サーバー間でチャットメッセージをやり取りするためのカスタムプロトコル、"Chat Message Protocol"の構造体として定義されている
// プロトコルの長さは最大 4096 byte
// | operation: 1byte | chatroom_id_size: 1byte | user_id_size: 1byte | message_size: 1byte | payload: 4092 byte |
// payload: chatroom_id(uuid) + user_id(uuid) + message
// idには、uuidを採用しており、その長さは36 bytesとなるはず
// したがってmessageが取りうる長さは 0 ~ 4020 byte
type ChatMessage struct {
	Operation  byte
	ChatRoomID string
	UserID     string
	Message    string
}

const (
	ChatProtocolMaxLen     = 4096
	ChatMessageBytesMaxLen = 4020
)

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
