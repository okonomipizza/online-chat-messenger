package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/okonomipizza/chat-server/pkg/chat"
	"github.com/okonomipizza/chat-server/pkg/data"
	"github.com/okonomipizza/chat-server/pkg/protocol"
)

// client から新しい chatRoom の作成か、既存の chatRomm への接続を求められるのでそれに対応する
func handleChatRoomRequest(conn net.Conn, dataStore *data.DataStore) {
	defer conn.Close()

	// メッセージを受信したことをクライアントへ知らせる
	ackResponse, err := protocol.AckResponse()
	_, err = conn.Write(ackResponse)
	if err != nil {
		fmt.Println("Failed to send ack response to client")
		return
	}

	// クライアントからのデータを読み取るためのバッファ
	bufferSize := 1024
	readBuf := make([]byte, bufferSize)
	count, _ := conn.Read(readBuf)
	fmt.Printf("%d bytes data received through tcp connection\n", count)

	request, err := protocol.ParseChatRoomRequest(readBuf)
	if err != nil {
		// TODO 無効なリクエストの時
		response, err := protocol.InvalidRequestResponse(string(err.Error()))
		println(err)
		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Failed to send respose to invalid request")
		}
	}
	fmt.Printf("request: %+v\n", request)

	// 新しいチャットルームの作成がリクエストされた場合
	if request.Operation == protocol.OperationCreateChatRoom {
		err = SendNewRoomResponse(conn, request, dataStore)
		if err != nil {
			response, _ := protocol.InternalServerErrorResponse()
			_, err = conn.Write(response)
			if err != nil {
				fmt.Println("Failed to send response to client")
				return
			}
			return
		}
		println("New Chat room Created")
		return

		// ChatRoomのIDによるChatRoom検索がリクエストされた場合
	} else if request.Operation == protocol.OperationSerchChatRoomByID {
		// リクエストに含まれるidに該当するチャットルームがあるか検索

		chatroom, exists := dataStore.ChatRooms[request.RoomID]
		if exists {
			response, _ := protocol.CreateExistingChatroomResponse(chatroom)
			_, err = conn.Write(response)
			if err != nil {
				fmt.Println("Failed to send response to client")
				return
			}
			return
		} else {
			response, _ := protocol.InvalidRequestResponse("No room exist")
			_, err = conn.Write(response)
			if err != nil {
				fmt.Println("Failed to send response to client")
				return
			}
			return
		}

		// チャットルームへの参加がリクエストされた場合
	} else if request.Operation == protocol.OperationJoinChatRoom {
		println("Catch request to join chat room")
		// Join to chat requested
		// チャットルームがあるかを確認
		chatRoom, err := dataStore.GetChatRoomByID(request.RoomID)
		if err != nil {
			println(err)
			return
		}

		// チャットルームのパスワードを確認
		if chatRoom.Password != "" && chatRoom.Password != request.RoomPassword {
			// リクエストされたパスワードが間違っていた時
			println("Invalid password requested")
			// 応答
			return
		}

		// 受信されたパスワードが正しければ、
		// リクエストに含まれる情報からユーザーインスタンスを作成し、所定のチャットルームへ登録する
		user := data.User{
			Id:     uuid.NewString(),
			Name:   request.UserName,
			IsHost: false,
		}

		err = dataStore.AddUsers(request.RoomID, user)
		if err != nil {
			println(err)
			return
		}

		// リクエストが許可されたことを応答する
		println("Request creating ...")
		response, _ := protocol.CreateChatRoomJoinResponse(user, chatRoom)
		println("Created request")

		_, err = conn.Write(response)
		if err != nil {
			fmt.Printf("failed to send response to join request %s\n", err)
			return
		}
		return

	}

}

// SendNewRoomResponseはクライアントの要望に沿った新しいチャットルームの作成を試みる。
// 成功した時は、作成されたチャットルームに関するデータをjson形式で表してpayloadに含める
// 失敗した時は、失敗した旨を送信 (state=1)
func SendNewRoomResponse(conn net.Conn, request protocol.ChatRoomRequest, dataStore *data.DataStore) error {
	user, chatRoom := chat.CreateNewChatRoom(request, dataStore)

	// レスポンスを作成
	response, err := protocol.CreateNewChatRoomResponse(user, chatRoom)
	if err != nil {
		return err
	}

	// レスポンスを送信
	_, err = conn.Write(response)
	if err != nil {
		return err
	}

	return nil
}

func hostingUDPServer(port string, datastore *data.DataStore) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	// UDP サーバーを開始
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		return
	}
	defer udpConn.Close()

	fmt.Println("UDP server listening on port", port)

	for {
		// クライアントからのメッセージを受信するバッファ
		buffer := make([]byte, 4096)
		n, addr, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving UDP packet:", err)
			continue
		}

		go handleChatMessages(udpConn, addr, buffer, n, datastore)

	}
}

func handleChatMessages(udpConn *net.UDPConn, addr *net.UDPAddr, data []byte, length int, datastore *data.DataStore) {
	fmt.Printf("Received %d bytes from %s: %s\n", length, addr.String(), string(data[:length]))
	req, err := protocol.ParseChatRequest(data)
	if err != nil {
		fmt.Print("Received data invalid to read")
		return
	}

	// exitがリクエストされたとき
	if req.Operation == protocol.ChatOperationExit {
		// ユーザーをチャットルームから外す
		// ユーザーがチャットルームのホストならチャットルームごとdatastoreから削除
		logoutUserName, err := datastore.DeleteUsers(req.ChatRoomID, req.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}
		// ユーザーが退出した場合は、それをサーバーから全員へ配信
		message := fmt.Sprintf("%d is logged out", logoutUserName)
		err = bradcastToClients(req.ChatRoomID, udpConn, message, datastore)
		return
	}

	chatroom, err := datastore.GetChatRoomByID(req.ChatRoomID)

	// udp addressが送られてきた時
	if req.Operation == protocol.ChatOperationSendUDPAddr {
		isUserMember, err := datastore.IsUserMemberOfChatRoom(chatroom.Id, req.UserID)
		if err != nil {
			fmt.Println("Failed to confirm user token")
			return
		}
		if isUserMember {
			fmt.Printf("saving udp addr\n")
			err = datastore.SaveUserUDPAddr(chatroom.Id, req.UserID, addr)
			if err != nil {
				fmt.Printf("Failed to save udp address of the user")
				return
			}
		}
	}

	// メッセージの配信リクエストが送られてきた時
	if req.Operation == protocol.ChatOperationSendMessage {
		// client全員へメッセージをブロードキャスト
		err = bradcastToClients(chatroom.Id, udpConn, req.Message, datastore)
		if err != nil {
			fmt.Printf("Failed to bradcast: %s\n", err)
			return
		}
	}

	return

}

func bradcastToClients(chatRoomID string, udpConn *net.UDPConn, message string, datastore *data.DataStore) error {
	datastore.Mu.Lock()
	defer datastore.Mu.Unlock()

	chatRoom, exists := datastore.ChatRooms[chatRoomID]
	if !exists {
		return errors.New("No chat room")
	}

	for _, user := range chatRoom.Users {
		// ユーザーのアドレスにメッセージを送信
		_, err := udpConn.WriteToUDP([]byte(message), user.Addr)
		if err != nil {
			fmt.Printf("Error sending message to user %s: %v\n", user.Name, err)
			continue

		}
		fmt.Printf("Message sent to user %s (%s)\n", user.Name, user.Addr.String())
	}
	return nil
}

func main() {
	dataStore := &data.DataStore{
		ChatRooms: make(map[string]data.ChatRoom),
	}

	// UDP サーバーを起動
	go hostingUDPServer("9090", dataStore)

	// TCPサーバーをポート8080でリッスン開始
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080...")

	for {
		// クライアントからの接続を待機
		tcpConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// 接続を新しいゴルーチンで処理
		go handleChatRoomRequest(tcpConn, dataStore)
	}
}
