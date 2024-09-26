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

// GetUserInputString は、ユーザーに対して target に対応した文字列の入力を求める
// target は、何を入力して欲しいかを指定するためのもので、入力を受け付ける前に標準出力へプリントされる
// 取得する文字列には、それをバイト列に変換した時の最小サイズと最大サイズを指定する必要がある
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

		if inputLen < min {
			fmt.Println("Input some string")
			continue
		}
		if inputLen > max {
			fmt.Println("Username you typed is too long")
			continue
		}
		break
	}
	return input
}

// GetUserChoiceBool はユーザーに yes or no で答えられる質問を問いかけ、その回答を得る
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

// GetUserActionChoiceで返される値
const (
	CreateNewChatRoom = 1
	JoinChatRoom      = 2
)

// GetUserActionChoice はユーザーに次の２つのどちらかの行動を選択させる
// 1) 新しいチャットルームの作成
// 2) 既存のチャットルームへの参加
func GetUserActionChoice() int {
	reader := bufio.NewReader(os.Stdin)
	var choice int

	// ユーザーへ選択肢を提示して取得
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
	return choice
}

// ユーザーの選択に対応したリクエストプロトコルを生成する
func GenerateRoomRequest(choice int) ([]byte, error) {
	// 選択に応じたリクエストを作成
	// 1) 新しいチャットルームの作成
	if choice == CreateNewChatRoom {
		request, err := CreateNewRoomRequest()
		if err != nil {
			return nil, err
		}
		return request, nil
	}
	// 2) 既存のチャットルームへの参加
	if choice == JoinChatRoom {
		request, err := CreateJoinRoomRequest()
		if err != nil {
			// CreateJoinRoomRequestの中で、ユーザーが指定した Chatroomが存在しない場合がある
			return nil, errors.New("room ID you typed does not exists on the server")
		}
		return request, nil
	}

	// 想定外
	return nil, errors.New("could not generate request protocol succesfully")
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

// GetRoomNameはサーバーにRoomIDを渡して、対応するチャットルームが存在すればその名前を返す
// また、指定されたチャットルームにログインパスワードが設定されているか否かをbool値で返す
func GetRoomNameByID(roomID string) (string, bool, error) {
	// 問い合わせ用の接続を用意する
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Error connecting to server")
		os.Exit(0)
	}
	defer conn.Close()

	// 検索したいチャットルームのIDと操作をリクエストに含める
	request := protocol.ChatRoomRequest{
		RoomID:    roomID,
		Operation: protocol.OperationSerchChatRoomByID,
		State:     protocol.StateRequest,
	}

	// リクエストを作成して送信
	requestProtocol, err := request.CreateRequestProtocol()
	if err != nil {
		return "", false, err
	}

	_, err = conn.Write(requestProtocol)
	if err != nil {
		fmt.Println("Failed to send request for room searching")
	}

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

	// チャットルームが存在しないときはアプリを終了
	if response.State == protocol.StateInvalid {
		fmt.Println("Designated Chatroom does not exist")
		os.Exit(0)
	}

	roomname := response.RoomName
	if response.RoomPassword == "" {
		// パスワードがチャットルームに設定されていない場合
		return roomname, false, nil
	} else {
		// パスワードがチャットルームに設定されている場合
		return roomname, true, nil
	}
}

// CreateJoinRoomRequest はユーザーの入力情報に基づいてチャットルームへの参加リクエストを作成する
// チャットルームが存在しない場合はそこで処理を終了する
func CreateJoinRoomRequest() ([]byte, error) {
	roomID := GetUserInputString("room id", 1, 64)
	roomName, isPasswordNeeded, err := GetRoomNameByID(roomID)
	if err != nil {
		fmt.Println("Some error occured: ", err)
		os.Exit(0)
	}

	// アクセスしようとしているroomの名前が正しいかユーザーに聞いて間違っていればアプリを終了
	question := fmt.Sprintf("Join the Room '%s'?", roomName)
	choiceForContinue := GetUserChoiceBool(question)
	if !choiceForContinue {
		fmt.Println("Room joining was canceled. The application will now exit")
		os.Exit(0)
	}

	// チャットルーム内で使用するハンドルネームを取得
	userName := GetUserInputString("user name", 1, 32)
	request := protocol.ChatRoomRequest{
		RoomID:    roomID,
		RoomName:  roomName,
		UserName:  userName,
		Operation: protocol.OperationJoinChatRoom,
		State:     protocol.StateRequest,
	}

	// チャットルームにパスワードが設定されている場合
	// ユーザーにパスワードの入力を求める
	if isPasswordNeeded {
		passwordInput := GetUserInputString("password", 1, 32)
		request.RoomPassword = passwordInput
	}

	// 取得した入力を元にリクエストを作成
	requestProtocol, err := request.CreateRequestProtocol()
	if err != nil {
		return nil, err
	}
	return requestProtocol, nil
}
