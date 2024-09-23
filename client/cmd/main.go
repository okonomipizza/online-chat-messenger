package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/okonomipizza/chat-client/pkg/cli"
	"github.com/okonomipizza/chat-client/pkg/protocol"
)

func main() {
	// サーバーとの間にtcp接続を確立
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// チャットルームの作成 or チャットルームへの参加をサーバーにリクエスト
	request, err := cli.GenerateRoomRequest()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conn.Write(request)

	// サーバーからの Ack response を確認
	err = protocol.ReceiveAckResponse(conn)
	if err != nil {
		fmt.Printf("Failed to receive Ack response from the server: %s", err)
		fmt.Print("Server may be unavailable")
		os.Exit(1)
	}
	println("The server is processing your request...")

	// サーバーの処理結果を受信
	readBuf := make([]byte, 1024)
	_, err = conn.Read(readBuf)
	if err != nil {
		fmt.Println("Failed to receive response from the server:", err)
		os.Exit(1)
	}

	response, err := protocol.ParseChatRoomResponse(readBuf)
	if err != nil {
		fmt.Printf("Failed to read response from the server", err)
		os.Exit(1)
	}

	if response.State == protocol.StateInvalid {
		fmt.Println("Your request refused from the server")
		os.Exit(1)
	}
	if response.State == protocol.StateFail {
		fmt.Println("Server error occured")
		os.Exit(1)
	}

	chatRoomID := response.RoomID
	userID := response.UserID

	fmt.Printf("Chat room ID:<%s> \n", chatRoomID)
	fmt.Println("You are Logged in to the room")

	conn, err = net.Dial("udp", "server:9090")
	if err != nil {
		fmt.Println("Error connecting to server by udp: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	// サーバーから配信されるメッセージを受信
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error receiving data: ", err)
				break
			}

			// ユーザー入力行をクリア
			fmt.Print("\r\033[K") // \033[K で行をクリア

			// サーバーからのメッセージを表示
			fmt.Println(string(buffer[:n]))
		}
	}()

	// サーバーへudp アドレスを知らせるために、空のメッセージを送信
	// udp アドレスはサーバー側で受信したメッセージをすべての参加メンバーへ配信するために使用する
	blankMessage := protocol.ChatMessage{
		ChatRoomID: chatRoomID,
		UserID:     userID,
		Message:    "",
	}
	udpAddrSendRequestProtocol, err := blankMessage.CreateChatRequest(protocol.ChatOperationSendUDPAddr)
	if err != nil {
		fmt.Printf("Error creating chat request : %s\n", err)
		os.Exit(1)
	}
	_, err = conn.Write(udpAddrSendRequestProtocol)
	if err != nil {
		fmt.Printf("Error sending blank message via UDP: %s\n", err)
	}

	// メッセージの入力を受け付けてサーバーへ送信
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter message (type 'exit' to quit):\n")
	for {
		scanner.Scan()
		input := scanner.Text()

		message := protocol.ChatMessage{
			ChatRoomID: chatRoomID,
			UserID:     userID,
			Message:    input,
		}

		// "exit"は、ユーザーがチャットルームから退出する意志をサーバーへ伝えたいときに実行される
		// cliクライアントは、サーバーへoperation exitを含めたリクエストを送信する
		if strings.ToLower(input) == "exit" {
			protocol, err := message.CreateChatRequest(protocol.ChatOperationExit)
			if err != nil {
				fmt.Printf("Error creating chat request : %s\n", err)
				continue
			}
			_, err = conn.Write(protocol)
			if err != nil {
				fmt.Printf("Failed to send exit message to server\nError: %s\n", err)
			}
			fmt.Println("Exit from Chat room")
			os.Exit(0)
		}

		// 入力された文字列の長さをチェック
		if len([]byte(input)) > protocol.ChatMessageBytesMaxLen {
			fmt.Println("Sorry! This Message is too long to send!")
			continue
		}

		// 入力された文字列が適切な長さであれば、サーバーへのリクエストメッセージを作成して送信
		request, err := message.CreateChatRequest(protocol.ChatOperationSendMessage)
		if err != nil {
			fmt.Printf("Error creating chat request : %s\n", err)
			continue
		}

		_, err = conn.Write(request)
		if err != nil {
			fmt.Printf("Failed to send message to server\nError: %s\n", err)
		}
	}
}
