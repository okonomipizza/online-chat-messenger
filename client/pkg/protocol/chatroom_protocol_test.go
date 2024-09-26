package protocol

import (
	"bytes"
	"testing"
)

func TestCreateRequestProtocol(t *testing.T) {
	req := ChatRoomRequest{
		RoomID:       "123456",
		RoomName:     "TestRoom",
		RoomPassword: "password123",
		UserID:       "user123",
		UserName:     "Alice",
		Operation:    OperationCreateChatRoom,
		State:        StateRequest,
	}

	expectedPayload := []byte(`{"room_id":"123456","room_name":"TestRoom","room_password":"password123","user_id":"user123","user_name":"Alice"}`)

	protocol, err := req.CreateRequestProtocol()
	if err != nil {
		t.Fatalf("CreateRequestProtocol returned an error: %v", err)
	}

	// バイト列の長さを確認 (ヘッダー + ペイロード)
	expectedLength := 1 + 1 + 1 + len(expectedPayload) // PayloadSize(1 byte) + Operation(1 byte) + State(1 byte) + Payload
	if len(protocol) != expectedLength {
		t.Errorf("expected protocol length %d, but got %d", expectedLength, len(protocol))
	}

	// PayloadSizeの確認
	payloadSize := protocol[0]
	if int(payloadSize) != len(expectedPayload) {
		t.Errorf("expected payload size %d, but got %d", len(expectedPayload), payloadSize)
	}

	// Operationの確認
	if protocol[1] != req.Operation {
		t.Errorf("expected operation %d, but got %d", req.Operation, protocol[1])
	}

	// Stateの確認
	if protocol[2] != req.State {
		t.Errorf("expected state %d, but got %d", req.State, protocol[2])
	}

	// Payloadの確認
	payload := protocol[3:]
	if !bytes.Equal(payload, expectedPayload) {
		t.Errorf("expected payload %s, but got %s", expectedPayload, payload)
	}
}
