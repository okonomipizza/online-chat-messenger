package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

// ChatRoomRequestはユーザーの入力から作成される
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

func (req ChatRoomRequest) payload() ([]byte, error) {
	data := map[string]interface{}{
		"room_id":       req.RoomID,
		"room_name":     req.RoomName,
		"room_password": req.RoomPassword,
		"user_id":       req.UserID,
		"user_name":     req.UserName,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON変換エラー", err)
		return nil, errors.New("failed to generate json data")
	}

	return jsonData, nil
}

func (req ChatRoomRequest) CreateRequestProtocol() ([]byte, error) {
	payload, err := req.payload()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	// PayloadSize
	if err := buf.WriteByte(byte(len(payload))); err != nil {
		return nil, err
	}

	// operation
	if err := buf.WriteByte(req.Operation); err != nil {
		return nil, err
	}

	// state
	if err := buf.WriteByte(req.State); err != nil {
		return nil, err
	}

	// payload
	if _, err := buf.Write(payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ReceiveAckResponse(conn net.Conn) error {
	// サーバーからのack responseは ChatRoomProtocolのヘッダのみの3 byteで送信される (payload, operation, state)
	readBuf := make([]byte, 3)
	count, err := conn.Read(readBuf)
	if count != 3 || err != nil {
		fmt.Println("Error reading from connection:", err)
		return errors.New("failed to load server response")
	}

	if readBuf[2] == StateAckResponse {
		return nil
	}

	return errors.New("internal server error")
}

func ReceiveResponse(conn net.Conn) (ChatRoomRequest, error) {
	readBuf := make([]byte, 1024)
	_, err := conn.Read(readBuf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return ChatRoomRequest{}, err
	}

	state := readBuf[2]

	if state == StateSuccess || state == StateInvalid {
		response, err := ParseChatRoomResponse(readBuf)
		if err != nil {
			return ChatRoomRequest{}, err
		}
		return response, nil
	} else {
		return ChatRoomRequest{}, errors.New("some errors occured")
	}
}

// ParseChatRoomProtocolはtcp接続により受信したbyte列を解析して構造体ChatRoomProtocolに変換する
func ParseChatRoomResponse(buf []byte) (ChatRoomRequest, error) {
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

	response := ChatRoomRequest{
		Operation: operation,
		State:     state,
	}

	// state が成功の時のみpayloadを読み込む
	if state == StateSuccess {
		err := json.Unmarshal(payload, &response)
		if err != nil {
			return ChatRoomRequest{}, errors.New("invalid payload for request")
		}
	}

	return response, nil
}
