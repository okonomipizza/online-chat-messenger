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
	// サーバーとのtcp接続を確立
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(0)
	}
	defer conn.Close()

	// チャットルームの作成 or 参加リクエストを送信
	request, err := cli.GenerateRoomRequest()
	if err != nil {
		fmt.Println("Internal error occored: %s", err)
		os.Exit(0)
	}
	conn.Write(request)

	// サーバーからの応答を確認
	err = protocol.ReceiveAckResponse(conn)
	if err != nil {
		fmt.Printf("Could not receive ACK Response from server: %s", err)
	}
	println("Server is processing your request...")

	// サーバーの処理結果を受信
	readBuf := make([]byte, 1024)
	count, err := conn.Read(readBuf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		os.Exit(0)
	}

	// 受け取ったデータを表示
	fmt.Printf("%d bytes received from server\n", count)
	fmt.Printf("Received data (Raw): %s\n", string(readBuf[:count])) // 文字列として表示

	// レスポンスOKならUDP接続を行う
	if readBuf[2] == protocol.StateInvalid {
		fmt.Println("Something in your request invalid")
		os.Exit(1)
	}

	response, err := protocol.ParseChatRoomResponse(readBuf)
	if err != nil {
		fmt.Printf("Invalid request server said", err)
	}

	chatRoomID := response.RoomID
	userID := response.UserID

	conn, err = net.Dial("udp", "server:9090")
	if err != nil {
		fmt.Println("Error connecting to server by udp: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	// サーバーからのメッセージを受信
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			fmt.Printf("%d bytes was brasdcasted from server\n", n)
			if err != nil {
				fmt.Println("Error receiving data: ", err)
				break
			}

			// サーバーからのメッセージを表示する前に、ユーザー入力行をクリア
			// response, err := protocol.ParseChat

			// ユーザー入力行をクリア
			fmt.Print("\r\033[K") // \033[K で行をクリア

			// サーバーからのメッセージを表示
			fmt.Printf("Server: %s\n", string(buffer[:n]))

			// 入力プロンプトを再度表示
			fmt.Print("Enter message (type 'exit' to quit): ")
		}
	}()

	// サーバーへudp アドレスを知らせるために、空のメッセージを送信
	blankMessage := protocol.ChatMessage{
		ChatRoomID: chatRoomID,
		UserID:     userID,
		Message:    "",
	}
	blankProtocol, err := blankMessage.CreateChatRequest()
	if err != nil {
		fmt.Printf("Error creating chat request : %s\n", err)
		return
	}
	_, err = conn.Write(blankProtocol)
	if err != nil {
		fmt.Printf("Error sending blank message via UDP: %s\n", err)
	}

	// ユーザー入力を監視してサーバーへ送信
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter message (type 'exit' to quit): ")
	for {
		scanner.Scan()
		input := scanner.Text()

		if strings.ToLower(input) == "exit" {
			fmt.Println("Exit from Chat room")
			break
		}

		message := protocol.ChatMessage{
			ChatRoomID: chatRoomID,
			UserID:     userID,
			Message:    input,
		}

		protocol, err := message.CreateChatRequest()
		if err != nil {
			fmt.Printf("Error creating chat request : %s\n", err)
			continue
		}

		_, err = conn.Write(protocol)
		if err != nil {
			fmt.Printf("Error sending message via UDP: %s\n", err)
		}
	}

}
