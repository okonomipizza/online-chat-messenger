package chat

import (
	"github.com/google/uuid"
	"github.com/okonomipizza/chat-server/pkg/data"
	"github.com/okonomipizza/chat-server/pkg/protocol"
)

func CreateNewChatRoom(request protocol.ChatRoomRequest, dataStore *data.DataStore) (data.User, data.ChatRoom) {
	// リクエストに含まれていた情報からサーバー側でユーザーインスタンスを作成する
	// チャットルームの作成者がそのルームのホストユーザーとなる
	user := data.User{
		Id:     uuid.NewString(),
		Name:   request.UserName,
		IsHost: true,
	}

	// リクエストからチャットルームインスタンスを作成する
	chatRoom := data.ChatRoom{
		Id:       uuid.NewString(),
		Name:     request.RoomName,
		Password: request.RoomPassword,
		Users:    make(map[string]data.User),
		Messages: []data.Message{},
	}

	// 作成したチャットルームにリクエストユーザーを追加
	chatRoom.Users[user.Id] = user

	// アプリケーション全体へ反映
	dataStore.AddChatRooms(chatRoom.Id, chatRoom)

	return user, chatRoom
}
