package protocol

import (
	"bytes"
	"testing"
)

func TestCreateChatRequest(t *testing.T) {
	chat := ChatMessage{
		Operation:  ChatOperationSendMessage,
		ChatRoomID: "12345678-1234-1234-1234-123456789012", // UUID
		UserID:     "87654321-4321-4321-4321-210987654321", // UUID
		Message:    "Hello, world!",
	}

	expected := []byte{
		ChatOperationSendMessage, 36, 36, byte(len([]byte(chat.Message))),
	}

	expected = append(expected, []byte(chat.ChatRoomID)...)
	expected = append(expected, []byte(chat.UserID)...)
	expected = append(expected, []byte(chat.Message)...)

	actual, err := chat.CreateChatRequest(ChatOperationSendMessage)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
