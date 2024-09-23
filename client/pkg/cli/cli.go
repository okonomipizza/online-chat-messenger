package cli

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/okonomipizza/chat-client/pkg/protocol"
)

func GetUserInputString(target string, min int, max int) string {
	var input string
	for {
		fmt.Printf("Input %s: ", target)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 64*1024), 100001)
		scanner.Scan()
		input = scanner.Text()

		inputBytes := []byte(input)
		inputLen := len(inputBytes)

		if inputLen < 1 {
			fmt.Println("Input some string")
			continue
		}
		if inputLen > 64 {
			fmt.Println("Username you typed is too long")
			continue
		}
		break
	}
	return input
}

func GetUserChoiceBool(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(question)
		fmt.Print("Enter y or n: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "y" {
			return true
		}
		if input == "n" {
			return false
		}

		fmt.Println("Invalid input. Please enter y or n.")
	}
}

// GenerateRoomRequestはユーザーに新しいチャットルームの作成か、既存のチャットルームへの参加のどちらを行うかを選択させる
// ユーザーの選択に応じて、適切なリクエストプロトコルのバイト列をインタラクティブに生成する
func GenerateRoomRequest() ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	var choice int
	for {
		fmt.Println("Choose an option:")
		fmt.Println("1. Create a new chat room")
		fmt.Println("2. Join an existing chat room")
		fmt.Print("Enter 1 or 2: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		inputInt, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Error converting input to integer:", err)
			fmt.Println("Invalid input. Please enter 1 or 2.")
			continue
		}

		choice = inputInt
		if choice == 1 || choice == 2 {
			break
		} else {
			fmt.Println("Invalid input. Please enter 1 or 2.")
		}
	}

	// New Chat Roomリクエストを作成
	if choice == 1 {
		println("New room creation start")
		request, err := CreateNewRoomRequest()
		if err != nil {
			return nil, err
		}
		return request, nil
	}
	// Chat Room への参加リクエストを作成
	if choice == 2 {
		request, err := CreateJoinRoomRequest()
		if err != nil {
			return nil, errors.New("Room ID you typed does not exists on the server")
		}
		return request, nil
	}

	return nil, errors.New("Could not generate request protocol succesfully")
}

func CreateNewRoomRequest() ([]byte, error) {
	userName := GetUserInputString("user name", 1, 32)
	roomName := GetUserInputString("chat room name", 1, 64)
	request := protocol.ChatRoomRequest{
		RoomName:  roomName,
		UserName:  userName,
		Operation: protocol.OperationCreateChatRoom,
		State:     protocol.StateRequest,
	}
	// passwordの設定は任意
	isPasswordNeeded := GetUserChoiceBool("Do you set password to the room?")
	if isPasswordNeeded {
		password := GetUserInputString("password", 1, 32)
		request.RoomPassword = password
	}

	requestProtocol, err := request.CreateRequestProtocol()
	if err != nil {
		return nil, err
	}

	return requestProtocol, nil
}

// GetRoomNameはサーバーにRoomIDを渡して、対応するChat Roomが存在すればそのChat room の名前を取得する
// また、指定されたチャットルームにログインパスワードが設定されているか否かもbool値で返す
func GetRoomNameByID(roomID string) (string, bool, error) {
	// 問い合わせ用のconnを用意する
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return "", false, errors.New("Server connection refused")
	}
	defer conn.Close()

	request := protocol.ChatRoomRequest{
		RoomID:    roomID,
		Operation: protocol.OperationSerchChatRoomByID,
		State:     protocol.StateRequest,
	}
	requestProtocol, err := request.CreateRequestProtocol()
	if err != nil {
		return "", false, err
	}
	conn.Write(requestProtocol)

	// ack responseを受信
	err = protocol.ReceiveAckResponse(conn)
	if err != nil {
		return "", false, err
	}

	// サーバーの処理結果を受信
	response, err := protocol.ReceiveResponse(conn)
	if err != nil {
		return "", false, err
	}
	roomname := response.RoomName
	if err != nil {
		return "", false, err
	}

	if response.RoomPassword == "" {
		return roomname, false, nil
	} else {
		return roomname, true, nil
	}
}

func CreateJoinRoomRequest() ([]byte, error) {
	roomID := GetUserInputString("room id", 1, 64)
	roomName, isPasswordNeeded, err := GetRoomNameByID(roomID)
	if err != nil {
		return nil, err
	}

	// アクセスしようとしているroomの名前が正しいか
	question := fmt.Sprintf("Join the Room '%s'?", roomName)
	isContinue := GetUserChoiceBool(question)
	if !isContinue {
		os.Exit(0)
	}

	userName := GetUserInputString("user name", 1, 32)
	request := protocol.ChatRoomRequest{
		RoomID:    roomID,
		RoomName:  roomName,
		UserName:  userName,
		Operation: protocol.OperationJoinChatRoom,
		State:     protocol.StateRequest,
	}

	// チャットルームにパスワードが設定されている場合は、パスワードの入力を受け付ける
	if isPasswordNeeded {
		password := GetUserInputString("password", 1, 32)
		request.RoomPassword = password
	}

	requestProtocol, err := request.CreateRequestProtocol()
	if err != nil {
		return nil, err
	}
	return requestProtocol, nil
}
