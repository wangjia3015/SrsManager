package room

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

func GetRoomManager() *RoomManager {
	return room_manager
}

func InitRoomManager() {
	room_manager = &RoomManager{}
}

var room_manager *RoomManager

func RoomHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method, req.URL.Path)
	// parse request
	var err error
	switch req.Method {
	case "PUT":
		err = room_manager.CreateRoom()
	case "DELETE":
		err = room_manager.DeleteRoom()
	}
}

type CheckRoomlegal interface {
	IsRoomExists(roomName string) bool
}

type RoomManager struct {
	room_map map[string]string
	mutex    sync.Mutex
}

func (r *RoomManager) IsRoomExists(roomName string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.room_map[roomName]
	return ok
}

func (r *RoomManager) tryAddRoom(roomName, userName string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.room_map[userName]; ok {
		return false
	}
	r.room_map[roomName] = userName
	return true
}

func (r *RoomManager) getRoomName(user, roomName string) string {
	roomName = fmt.Sprintf("%s-%s_jd_add", user, roomName)
	md5 := GetMD5String(roomName)
	return md5
}

func (r *RoomManager) CreateRoom(user, roomName string) (string, error) {
	md5 := r.getRoomName(user, roomName)
	if r.tryAddRoom(md5, roomName) {
		return md5, nil
	}
	return "", errors.New("room already exists")
}

func (r *RoomManager) DeleteRoom(user, roomName string) error {
	// TODO ban publisher
	r.mutex.Lock()
	defer r.mutex.Unlock()
	md5 := r.getRoomName(user, roomName)
	delete(r.room_map, md5)
	return nil
}
