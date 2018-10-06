package main

import (
  "sync"
  "encoding/json"
  "errors"
  "github.com/gorilla/websocket"
  "github.com/mitchellh/mapstructure"
  "log"
)

const (
  TypeRecovery      = "Recover"       // server
  TypeCreatePlayer  = "CreatePlayer"  // server
  TypePlayerInfo    = "PlayerInfo"    // client
  TypeCreateRoom    = "CreateRoom"    // client
  TypeError         = "Error"         // client
  TypeRoomInfo      = "RoomInfo"      // client
  TypeRooms         = "Rooms"         // client
  TypeJoinRoom      = "JoinRoom"      // server
  TypeLeaveRoom     = "LeaveRoom"     // server
  TypeKickPlayer    = "KickPlayer"    // client
  TypePlayerMove    = "PlayerMove"    // server
  TypeGameInfo      = "GameInfo"      // client
  TypeGameMove      = "GameMove"      // client
  TypeGameOver      = "GameOver"      // client
)

type Message struct {
  Info string
  Data map[string]interface{}
}

type GameMove struct {
  X int
  Y int
  PlayerIndex int
}

type Player struct {
  Name string
  Room *Room
  CurrentSession *Session
  MoveQueue chan GameMove
}

type Info struct {
  Name string
}

type Room struct {
  Players [] *Player
  Name string
  JoinMutex *sync.Mutex `json:"-"`
}

type Rooms struct {
  Rooms [] *Room
}

type Session struct {
  Socket *websocket.Conn
  Player *Player
  WriteQueue chan Message
}

type ServerSingleton struct {
  Players map[string] *Player
  Rooms map[string] *Room
  Sessions [] *Session
  ServerMutex *sync.Mutex
}

var Server = ServerSingleton {
  Players: map[string] *Player{},
  Rooms: map[string] *Room{},
  Sessions: make([]*Session, 0),
  ServerMutex: &sync.Mutex{},
}

func HandleSession(connection *websocket.Conn) {
  session := &Session {
    Socket: connection,
    WriteQueue: make(chan Message),
  }
  Server.ServerMutex.Lock();
  Server.Sessions = append(Server.Sessions, session)
  Server.ServerMutex.Unlock()

  go session.Sender()
  for {
    message_type, message, err := session.Socket.ReadMessage()
    if (err != nil) {
      // err
    }
    if (message_type == websocket.TextMessage) {
      message_obj := Message{}
      err = json.Unmarshal(message, &message_obj)
      if (err != nil) {
        session.InformSession(err)
        continue
      }
      switch message_obj.Info {
      case TypeRecovery:
        session.Recovery(message_obj)
      case TypeCreatePlayer:
        session.CreatePlayer(message_obj)
      case TypeJoinRoom:
        session.JoinRoom(message_obj)
      case TypePlayerMove:
        session.DoMove(message_obj)
      case TypeCreateRoom:
        session.CreateRoom(message_obj)
      }
    } else if (message_type == websocket.BinaryMessage) {
      // Do not handle.
    } else if (message_type == websocket.CloseMessage || message_type == -1) {
      if (session.Player != nil) {
        // Invalidate current session.
        session.Player.CurrentSession = nil
      }
      Server.ServerMutex.Lock()
      for i := range Server.Sessions {
        if Server.Sessions[i] == session {
          Server.Sessions = append(Server.Sessions[:i],
                                   Server.Sessions[i + 1:]...)
        }
      }
      Server.ServerMutex.Unlock()
      break
    } else {
      // Should not happen.
    }
  }
}

func (session *Session) Sender() {
  for {
    message := <- session.WriteQueue
    session.Socket.WriteJSON(message)
  }
}

func (session *Session) InformSession(err error) {
  info := Info {
    Name: err.Error(),
  }
  message := CreateMessage(TypeError, info)
  // Send message
  session.WriteQueue <- message
}

func (session *Session) CreatePlayer(message Message) {
  if (session.Player != nil) {
     session.InformSession(errors.New("Player already created."))
     return
  }
  info := Info{}
  err := mapstructure.Decode(message.Data, &info)
  if (err != nil) {
    session.InformSession(err)
    return
  }
  Server.ServerMutex.Lock()
  if _, ok := Server.Players[info.Name]; ok {
    session.InformSession(errors.New("Name already taken"))
    return
  }
  player := &Player {
    Name: info.Name,
    CurrentSession: session,
    Room: nil,
    MoveQueue: nil,
  }
  session.Player = player
  Server.Players[player.Name] = player
  Server.ServerMutex.Unlock()

  // Send message
  reply := CreateMessage(TypePlayerInfo,info)
  session.WriteQueue <- reply
}

func CreateMessage(info string, obj interface{}) Message {
  var data = map[string]interface{}{}
  js, _ := json.Marshal(obj)
  json.Unmarshal(js, &data)
  message := Message {
    Info: info,
    Data: data,
  }
  return message
}

func (session *Session) JoinRoom(message Message) {
  if (session.Player == nil) {
     session.InformSession(errors.New("Player does not exist. Cannot join."))
     return
  }
  if (session.Player.Room != nil) {
     session.InformSession(errors.New("Player already in some room."))
     return
  }
  info := Info{}
  err := mapstructure.Decode(message.Data, &info)
  if (err != nil) {
    session.InformSession(err)
    return
  }
  log.Println("Join room")
  log.Println(info.Name)
  Server.ServerMutex.Lock()
  if _, ok := Server.Rooms[info.Name]; !ok {
    session.InformSession(errors.New("Room does not exist."))
    return
  }
  if (len(Server.Rooms[info.Name].Players) > 1) {
    session.InformSession(errors.New("Room full."))
    return
  }
  log.Println("JOIN")

  Server.Rooms[info.Name].Players = append(Server.Rooms[info.Name].Players,
                                           session.Player)
  log.Println("JOIN")

  for _, player := range Server.Rooms[info.Name].Players {
  log.Println("JOIN2")
      player.CurrentSession.WriteQueue <- CreateMessage(TypeRoomInfo,
                                                        Server.Rooms[info.Name])
  }
  log.Println("JOIN")
  Server.ServerMutex.Unlock()
  log.Println("JOIN")
  if (len(Server.Rooms[info.Name].Players) == 2) {
  log.Println("JOIN")
    game := &Game {
      Moves: make([] GameMove, 0),
      CurrentPlayer: 0,
      Room: Server.Rooms[info.Name],
    }
    go game.StartGame()
  }
}

func (session *Session) DoMove(message Message) {
  if (session.Player == nil) {
     return
     //session.InformSession(errors.New("Player does not exist. Cannot join."))
     //return
  }
  if (session.Player.MoveQueue == nil) {
    session.InformSession(errors.New("Don't poke."))
    return
  }

  move := GameMove{}
  err := mapstructure.Decode(message.Data, &move)
  if (err != nil) {
    session.InformSession(err)
    return
  }

  session.Player.MoveQueue <- move
}

func (session *Session) CreateRoom(message Message) {
  if (session.Player == nil) {
     session.InformSession(errors.New("Player does not exist. Cannot join."))
     return
  }
  if (session.Player.Room != nil) {
     session.InformSession(errors.New("Player already in some room."))
     return
  }

  info := Info{}
  err := mapstructure.Decode(message.Data, &info)
  if (err != nil) {
    session.InformSession(err)
    return
  }
  Server.ServerMutex.Lock()
  if _, ok := Server.Rooms[info.Name]; ok {
    session.InformSession(errors.New("Room does exist."))
    return
  }
  room := &Room {
    Name: info.Name,
  }
  Server.Rooms[room.Name] = room

  Server.ServerMutex.Unlock()

  log.Println("JOIN")
  session.JoinRoom(message)
  log.Println("JOINed")

  Server.ServerMutex.Lock()
  for _, player := range Server.Players {
    player.CurrentSession.WriteQueue <- CreateMessage(TypeRooms,
                                                      Server.Rooms)
  }
  Server.ServerMutex.Unlock()
}

func (session *Session) Recovery(message Message) {

}
