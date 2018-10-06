package main

import (
  "net/http"
  "github.com/gorilla/websocket"
  "log"
)

var finish = make(chan bool)

func HandleIndex(resp_writer http.ResponseWriter,
                 request *http.Request) {
  log.Println("Handle Index FUNCTION")
  http.ServeFile(resp_writer, request, "game/index.html")
}

func HandleUpgrade(resp_writer http.ResponseWriter,
                   request *http.Request) {
  conn, err := websocket.Upgrade(resp_writer,
                                 request,
                                 resp_writer.Header(),
                                 1024, 1024)
  if (err != nil) {
    log.Println("asdasd")
    http.Error(resp_writer,"Failed to open websocket", http.StatusBadRequest)
    return
  }
  // In new thread
  log.Println("Handle Session FUNCTION")
  go HandleSession(conn)
}

func main() {
  log.Println("Handle Func")
  http.HandleFunc("/", HandleIndex)
  log.Println("Handle Session")
  http.Handle("/game/",
              http.StripPrefix("/game", http.FileServer(http.Dir("game"))))
  log.Println("Handle func")
  http.HandleFunc("/upgrade", HandleUpgrade)
  log.Println("Listen")
  http.ListenAndServe(":8080", nil)
  log.Println("Fin")
  <- finish
}

