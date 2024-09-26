package protocol

import (
	"encoding/json"
	"testing"

	"github.com/okonomipizza/chat-server/pkg/data"
)

func TestAckResponse(t *testing.T) {
	response, err := AckResponse()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(response) < 3 {
		t.Fatal("response length is too short")
	}

	if response[0] != 0 {
		t.Errorf("expected payload size 0, got %d", response[0])
	}

	if response[1] != 0 {
		t.Errorf("expected operation 0, got %d", response[1])
	}

	if response[2] != 1 {
		t.Errorf("expected state 1, got %d", response[2])
	}
}

func TestInvalidRequestResponse(t *testing.T) {
	message := "Invalid request"
	response, err := InvalidRequestResponse(message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(response) < 3 {
		t.Fatal("response length is too short")
	}

	if response[0] != byte(len(message)) {
		t.Errorf("expected payload size %d, got %d", len(message), response[0])
	}

	if response[1] != 0 {
		t.Errorf("expected operation 0, got %d", response[1])
	}

	if response[2] != StateFail {
		t.Errorf("expected state %d, got %d", StateFail, response[2])
	}
}

func TestInternalServerErrorResponse(t *testing.T) {
	response, err := InternalServerErrorResponse()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	message := "Internal Server Error"
	if len(response) < 3 {
		t.Fatal("response length is too short")
	}

	if response[0] != byte(len(message)) {
		t.Errorf("expected payload size %d, got %d", len(message), response[0])
	}

	if response[1] != 0 {
		t.Errorf("expected operation 0, got %d", response[1])
	}

	if response[2] != 3 {
		t.Errorf("expected state 3, got %d", response[2])
	}
}

func TestCreateChatRoomJoinResponse(t *testing.T) {
	user := data.User{Id: "user-id-123", Name: "Alice"}
	chatRoom := data.ChatRoom{Id: "room-id-123", Name: "General Room"}

	response, err := CreateChatRoomJoinResponse(user, chatRoom)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(response) < 3 {
		t.Fatal("response length is too short")
	}

	var parsedData map[string]interface{}
	if err := json.Unmarshal(response[3:], &parsedData); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsedData["room_id"] != chatRoom.Id {
		t.Errorf("expected room_id %s, got %s", chatRoom.Id, parsedData["room_id"])
	}
	if parsedData["user_id"] != user.Id {
		t.Errorf("expected user_id %s, got %s", user.Id, parsedData["user_id"])
	}
}

func TestCreateNewChatRoomResponse(t *testing.T) {
	user := data.User{Id: "user-id-456", Name: "Bob"}
	chatRoom := data.ChatRoom{Id: "room-id-456", Name: "Tech Room", Password: "secret"}

	response, err := CreateNewChatRoomResponse(user, chatRoom)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(response) < 3 {
		t.Fatal("response length is too short")
	}

	var parsedData map[string]interface{}
	if err := json.Unmarshal(response[3:], &parsedData); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsedData["room_id"] != chatRoom.Id {
		t.Errorf("expected room_id %s, got %s", chatRoom.Id, parsedData["room_id"])
	}
	if parsedData["user_id"] != user.Id {
		t.Errorf("expected user_id %s, got %s", user.Id, parsedData["user_id"])
	}
}

func TestParseChatRoomRequest(t *testing.T) {
	// Prepare a valid request
	originalRequest := ChatRoomRequest{
		RoomID:       "room-id-789",
		RoomName:     "Sports Room",
		RoomPassword: "password",
		UserID:       "user-id-789",
		UserName:     "Charlie",
		Operation:    OperationJoinChatRoom,
		State:        StateRequest,
	}

	// Create the byte representation
	payload, err := json.Marshal(originalRequest)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	buf := []byte{byte(len(payload)), originalRequest.Operation, originalRequest.State}
	buf = append(buf, payload...)

	// Parse the request
	parsedRequest, err := ParseChatRoomRequest(buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate the parsed values
	if parsedRequest.Operation != originalRequest.Operation {
		t.Errorf("expected operation %d, got %d", originalRequest.Operation, parsedRequest.Operation)
	}
	if parsedRequest.RoomID != originalRequest.RoomID {
		t.Errorf("expected room_id %s, got %s", originalRequest.RoomID, parsedRequest.RoomID)
	}
	if parsedRequest.UserID != originalRequest.UserID {
		t.Errorf("expected user_id %s, got %s", originalRequest.UserID, parsedRequest.UserID)
	}
}
