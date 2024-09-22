package protocol

import (
	"bytes"
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
