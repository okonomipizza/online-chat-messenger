package data

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
)

type User struct {
	Id     string
	Name   string
	Addr   *net.UDPAddr
	IsHost bool
}

type ChatRoom struct {
	Id       string
	Name     string
	Password string
	Users    map[string]User
	Messages []Message
}

type Message struct {
	Content string
	User    User
}

type DataStore struct {
	ChatRooms map[string]ChatRoom
	Mu        sync.Mutex
}

func (ds *DataStore) AddChatRooms(id string, room ChatRoom) {
	ds.Mu.Lock()
	ds.ChatRooms[id] = room
	ds.Mu.Unlock()
}

func (ds *DataStore) DeleteChatRooms(id string) {
	ds.Mu.Lock()
	delete(ds.ChatRooms, id)
	ds.Mu.Unlock()
}

// IsUserMemberOfChatRoom はそのユーザーが与えられた指定されたチャットルームに存在するかと、存在する場合はユーザ名を返す
func (ds *DataStore) IsUserMemberOfChatRoom(chatRoomID string, userID string) (bool, User, error) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	chatRoom, exists := ds.ChatRooms[chatRoomID]
	if !exists {
		return false, User{}, errors.New("designated ChatRoom does not exist")
	}

	user, exists := chatRoom.Users[userID]
	if !exists {
		return false, User{}, nil
	}
	return true, user, nil
}

func (ds *DataStore) SaveUserUDPAddr(chatRoomID string, userID string, addr *net.UDPAddr) error {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	chatRoom, exists := ds.ChatRooms[chatRoomID]
	if !exists {
		return errors.New("designated ChatRoom does not exist")
	}
	user, exists := chatRoom.Users[userID]
	if !exists {
		return errors.New("User is not a member of the room")
	}
	user.Addr = addr
	chatRoom.Users[userID] = user
	ds.ChatRooms[chatRoomID] = chatRoom

	return nil
}

func (ds *DataStore) AddUsers(chatRoomID string, user User) error {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	chatRoom, exists := ds.ChatRooms[chatRoomID]
	if exists {
		chatRoom.Users[user.Id] = user
	} else {
		return errors.New("designated ChatRoom does not exist")
	}
	ds.ChatRooms[chatRoomID] = chatRoom

	// chatroomにおける現在のメンバーを一覧にして表示
	currentUsers := ds.ChatRooms[chatRoomID].Users
	userList := []string{}
	for _, currentUser := range currentUsers {
		userList = append(userList, currentUser.Name) // ユーザー名を取得して追加
	}
	fmt.Println("Current users in chat room:", strings.Join(userList, ", "))

	return nil
}

func (ds *DataStore) DeleteUsers(chatRommID string, user_id string) (string, error) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	chatRoom, exists := ds.ChatRooms[chatRommID]
	if !exists {
		return "", errors.New("designated ChatRoom does not exist")
	}
	user, exists := chatRoom.Users[user_id]
	if !exists {
		return "", errors.New("designated user does not exist")
	}
	// userがhostユーザーか確認する
	if user.IsHost {
		delete(ds.ChatRooms, chatRommID)
		fmt.Printf("Chat room 'id: %s, name: %s' removed", chatRommID, chatRoom.Name)
	} else {
		delete(chatRoom.Users, user_id)
		ds.ChatRooms[chatRommID] = chatRoom
		fmt.Printf("'id: %s, name: %s' is logged out from Chat room 'id: %s, name: %s'", user.Id, user.Name, chatRommID, chatRoom.Name)
	}

	return user.Name, nil
}

func (ds *DataStore) GetChatRoomByID(chatRoomID string) (ChatRoom, error) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	chatRoom, exists := ds.ChatRooms[chatRoomID]
	if exists {
		return chatRoom, nil
	}
	return ChatRoom{}, errors.New("designated ChatRoom does not exist")
}

func (ds *DataStore) ConfirmPassword(chatRoomID string, password_input string) (bool, error) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	chatRoom, exists := ds.ChatRooms[chatRoomID]
	if exists {
		savedPassword := chatRoom.Password
		if savedPassword == "" || savedPassword == password_input {
			return true, nil
		} else {
			return false, errors.New("invalid password")
		}
	}
	return false, errors.New("designated Chatroom does not exist")
}
