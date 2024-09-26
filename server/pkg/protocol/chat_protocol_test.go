package protocol

import (
	"testing"
)

func TestCreateChatRequest(t *testing.T) {
	chatMessage := ChatMessage{
		Operation:  ChatOperationSendMessage,
		ChatRoomID: "123e4567-e89b-12d3-a456-426614174000",
		UserID:     "123e4567-e89b-12d3-a456-426614174001",
		Message:    "Hello, World!",
	}

	data, err := chatMessage.CreateChatRequest(chatMessage.Operation)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) > ChatProtocolMaxLen {
		t.Fatalf("expected length to be <= %d, got %d", ChatProtocolMaxLen, len(data))
	}

	if data[0] != chatMessage.Operation {
		t.Errorf("expected operation %d, got %d", chatMessage.Operation, data[0])
	}

	if data[1] != byte(len(chatMessage.ChatRoomID)) {
		t.Errorf("expected chatRoomID size %d, got %d", len(chatMessage.ChatRoomID), data[1])
	}
	if data[2] != byte(len(chatMessage.UserID)) {
		t.Errorf("expected userID size %d, got %d", len(chatMessage.UserID), data[2])
	}
	if data[3] != byte(len(chatMessage.Message)) {
		t.Errorf("expected message size %d, got %d", len(chatMessage.Message), data[3])
	}
}

func TestParseChatRequest(t *testing.T) {
	chatMessage := ChatMessage{
		Operation:  ChatOperationSendMessage,
		ChatRoomID: "123e4567-e89b-12d3-a456-426614174000",
		UserID:     "123e4567-e89b-12d3-a456-426614174001",
		Message:    "Hello, World!",
	}

	data, err := chatMessage.CreateChatRequest(chatMessage.Operation)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parsedMessage, err := ParseChatRequest(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if parsedMessage.Operation != chatMessage.Operation {
		t.Errorf("expected operation %d, got %d", chatMessage.Operation, parsedMessage.Operation)
	}
	if parsedMessage.ChatRoomID != chatMessage.ChatRoomID {
		t.Errorf("expected chatRoomID %s, got %s", chatMessage.ChatRoomID, parsedMessage.ChatRoomID)
	}
	if parsedMessage.UserID != chatMessage.UserID {
		t.Errorf("expected userID %s, got %s", chatMessage.UserID, parsedMessage.UserID)
	}
	if parsedMessage.Message != chatMessage.Message {
		t.Errorf("expected message %s, got %s", chatMessage.Message, parsedMessage.Message)
	}
}
