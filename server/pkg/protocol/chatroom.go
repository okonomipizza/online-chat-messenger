package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/okonomipizza/chat-server/pkg/data"
)

// ChatRoomProtocol: アプリケーション層で動作するカスタムプロトコル
// Operation: 1 byte
// State: 1 byte
// Payload: username (32 byte) + roomid (16 byte) + roomname (64 byte) + password(32 byte)

const payloadMaxLen = 144

// ChatRoomProtocolはユーザーの入力から作成される
// operation = 0: chat roomの作成をリクエストする時に使用
// operation = 1: chat roomへの参加をリクエストする時に使用
// state = 0: 成功レスポンス
// state = 1: 失敗レスポンス
type ChatRoomRequest struct {
	RoomID       string `json:"room_id"`
	RoomName     string `json:"room_name"`
	RoomPassword string `json:"room_password"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	Operation    byte
	State        byte
}

const (
	OperationCreateChatRoom byte = iota
	OperationSerchChatRoomByID
	OperationJoinChatRoom
	OperationLeaveChatRoom
)

const (
	StateRequest byte = iota
	StateAckResponse
	StateSuccess
	StateFail
	StateInvalid
)

// AckResponseはサーバーがリクエストを受信したら受信した事実のみを返すためのもの
func AckResponse() ([]byte, error) {
	buf := new(bytes.Buffer)
	// payload size
	if err := buf.WriteByte(byte(0)); err != nil {
		return nil, err
	}
	// operation
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// state 1はリクエストの受信を示す
	if err := buf.WriteByte(1); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ResponseToInvalidRequestはリクエストが無効な際にその旨をクライアントへ伝えるためのもの
func InvalidRequestResponse(message string) ([]byte, error) {
	buf := new(bytes.Buffer)

	payload := []byte(message)

	// payload size
	if err := buf.WriteByte(byte(len(payload))); err != nil {
		return nil, err
	}
	// operation
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// state 4は無効なリクエストが送信されたことを示す
	if err := buf.WriteByte(4); err != nil {
		return nil, err
	}
	// payload
	if _, err := buf.Write(payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func InternalServerErrorResponse() ([]byte, error) {
	buf := new(bytes.Buffer)

	message := "Internal Server Error"
	payload := []byte(message)

	// payload size
	if err := buf.WriteByte(byte(len(payload))); err != nil {
		return nil, err
	}
	// operation
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// state 3はサーバー側でエラーが発生したことを示す
	if err := buf.WriteByte(3); err != nil {
		return nil, err
	}
	// payload
	if _, err := buf.Write(payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func CreateExistingChatroomResponse(chatroom data.ChatRoom) ([]byte, error) {
	buf := new(bytes.Buffer)

	data := map[string]interface{}{
		"room_id":       chatroom.Id,
		"room_name":     chatroom.Name,
		"room_password": chatroom.Password,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON変換エラー", err)
		return nil, errors.New("Failed to generate json data")
	}
	fmt.Println("JSON created:", string(jsonData))

	// payloadSize
	if err := buf.WriteByte(byte(len(jsonData))); err != nil {
		return nil, err
	}

	// Operation (1 byte)
	if err := buf.WriteByte(OperationSerchChatRoomByID); err != nil {
		return nil, err
	}

	// State (1 byte) //完了: 2
	if err := buf.WriteByte(StateSuccess); err != nil {
		return nil, err
	}

	// payload
	if _, err := buf.Write(jsonData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func CreateChatRoomJoinResponse(user data.User, chatRoom data.ChatRoom) ([]byte, error) {
	buf := new(bytes.Buffer)

	data := map[string]interface{}{
		"room_id":   chatRoom.Id,
		"room_name": chatRoom.Name,
		"user_id":   user.Id,
		"user_name": user.Name,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON変換エラー", err)
		return nil, errors.New("Failed to generate json data")
	}
	fmt.Println("JSON created:", string(jsonData))

	// payloadSize
	if err := buf.WriteByte(byte(len(jsonData))); err != nil {
		return nil, err
	}

	// Operation (1 byte)
	if err := buf.WriteByte(OperationJoinChatRoom); err != nil {
		return nil, err
	}

	// State (1 byte) //完了: 2
	if err := buf.WriteByte(StateSuccess); err != nil {
		return nil, err
	}

	// payload
	if _, err := buf.Write(jsonData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

func CreateNewChatRoomResponse(user data.User, chatroom data.ChatRoom) ([]byte, error) {
	buf := new(bytes.Buffer)

	data := map[string]interface{}{
		"room_id":       chatroom.Id,
		"room_name":     chatroom.Name,
		"room_password": chatroom.Password,
		"user_id":       user.Id,
		"user_name":     user.Name,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON変換エラー", err)
		return nil, errors.New("Failed to generate json data")
	}
	fmt.Println("JSON created:", string(jsonData))

	// payloadSize
	if err := buf.WriteByte(byte(len(jsonData))); err != nil {
		return nil, err
	}

	// Operation (1 byte) チャットルームの作成(0)に対する応答なので
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}

	// State (1 byte) //完了: 2
	if err := buf.WriteByte(2); err != nil {
		return nil, err
	}

	// payload
	if _, err := buf.Write(jsonData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

// ParseChatRoomProtocolはtcp接続により受信したbyte列を解析して構造体ChatRoomProtocolに変換する
func ParseChatRoomRequest(buf []byte) (ChatRoomRequest, error) {
	if len(buf) < 3 {
		return ChatRoomRequest{}, errors.New("buffer size is too small")
	}

	// ヘッダの情報を取得
	payloadSize := buf[0]
	operation := buf[1]
	state := buf[2]
	payload := buf[3 : payloadSize+3]

	if byte(len(payload)) != payloadSize {
		return ChatRoomRequest{}, errors.New("recieved packet is not complete")
	}

	request := ChatRoomRequest{
		Operation: operation,
		State:     state,
	}

	err := json.Unmarshal(payload, &request)
	if err != nil {
		return ChatRoomRequest{}, errors.New("Invalid payload for request")
	}

	return request, nil
}
