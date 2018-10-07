package main

import (
  //"log"
)

type Game struct {
  Room *Room         `json:"-"`
  Moves [] GameMove
  CurrentPlayer int
  Table [][] int
}

func (game *Game) StartGame() {
  game.Table = [][] int {}
  game.Table = append(game.Table, [] int{-10, -10, -10})
  game.Table = append(game.Table, [] int{-10, -10, -10})
  game.Table = append(game.Table, [] int{-10, -10, -10})
  game.Room.Players[0].MoveQueue = make(chan GameMove)
  game.Room.Players[1].MoveQueue = make(chan GameMove)
  for {
    var move GameMove;
    select {
    case move = <- game.Room.Players[0].MoveQueue:
      move.PlayerIndex = 0
    case move = <- game.Room.Players[1].MoveQueue:
      move.PlayerIndex = 1
    }
    if (game.CurrentPlayer != move.PlayerIndex) {
      continue
    }
    if (game.Table[move.X][move.Y] != -10) {
      continue
    }
    game.Table[move.X][move.Y] = move.PlayerIndex

    game.Moves = append(game.Moves, move)
    message := CreateMessage(TypeGameMove, move)
    game.Room.Players[0].CurrentSession.WriteQueue <- message
    game.Room.Players[1].CurrentSession.WriteQueue <- message
    game.CurrentPlayer = (game.CurrentPlayer + 1) % 2
    win := game.CheckWin()
    draw := game.CheckDraw()
    if (win == -1) {
      if (draw == 1) {
        // Draw
        info := Info {
          Name: "Draw",
        }
        message = CreateMessage(TypeGameDraw, info)
        game.Room.Players[0].CurrentSession.WriteQueue <- message
        game.Room.Players[1].CurrentSession.WriteQueue <- message

        game.Room.Players[0].Room = nil
        game.Room.Players[1].Room = nil
        Server.ServerMutex.Lock()
        delete (Server.Rooms, game.Room.Name)
        for _, player := range Server.Players {
          player.CurrentSession.WriteQueue <- CreateMessage(TypeRooms,
                                                            Server.Rooms)
        }
        Server.ServerMutex.Unlock()
        break;
      } else {
        continue
      }
    }

    winner_name := game.Room.Players[win].Name
    info := Info {
      Name: winner_name,
    }
    message = CreateMessage(TypeGameOver, info)
    game.Room.Players[0].CurrentSession.WriteQueue <- message
    game.Room.Players[1].CurrentSession.WriteQueue <- message

    game.Room.Players[0].Room = nil
    game.Room.Players[1].Room = nil
    Server.ServerMutex.Lock()
    delete (Server.Rooms, game.Room.Name)
    for _, player := range Server.Players {
      player.CurrentSession.WriteQueue <- CreateMessage(TypeRooms,
                                                        Server.Rooms)
    }
    Server.ServerMutex.Unlock()
    break;
  }
}

func (game *Game) CheckDraw() int {
  for x := 0; x < 3; x++ {
    for y := 0; y < 3; y++ {
      if (game.Table[x][y] == -10) {
        return 0
      }
    }
  }
  // No more moves available.
  return 1
}

func (game *Game) CheckWin() int {
  var sum int
  sum = game.Table[0][0] + game.Table[0][1] + game.Table[0][2]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[1][0] + game.Table[1][1] + game.Table[1][2]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[2][0] + game.Table[2][1] + game.Table[2][2]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[0][0] + game.Table[1][0] + game.Table[2][0]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[0][1] + game.Table[1][1] + game.Table[2][1]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[0][2] + game.Table[1][2] + game.Table[2][2]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[0][0] + game.Table[1][1] + game.Table[2][2]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  sum = game.Table[2][0] + game.Table[1][1] + game.Table[0][2]
  if (sum == 3) { return 1 }
  if (sum == 0) { return 0 }
  return -1;
}
