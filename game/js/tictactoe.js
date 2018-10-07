const TypeRecovery      = "Recover"       // server
const TypeCreatePlayer  = "CreatePlayer"  // server
const TypePlayerInfo    = "PlayerInfo"    // client
const TypeCreateRoom    = "CreateRoom"    // server
const TypeError         = "Error"         // client
const TypeRoomInfo      = "RoomInfo"      // client
const TypeRoomAdded     = "RoomAdded"     // client
const TypeRooms         = "Rooms"         // client
const TypeJoinRoom      = "JoinRoom"      // server
const TypeLeaveRoom     = "LeaveRoom"     // server
const TypePlayerMove    = "PlayerMove"    // server
const TypeGameInfo      = "GameInfo"      // client
const TypeGameMove      = "GameMove"      // client
const TypeGameOver      = "GameOver"      // client
const TypeGameDraw      = "GameDraw"      // client


var socket = new WebSocket("ws://localhost:8080/upgrade")
var player_name = "";
var room_name = "";
var doc_playername = document.getElementById("playername");
var doc_roomname = document.getElementById("roomname");
var doc_rooms = document.getElementById("Rooms");

var lineColor = "#ddd";

socket.onopen = function () {
    console.log("Created websocket connection");
    var session = {};
    session.Info = TypeRecovery;
    session.Data = {};
    session.Data.Name = getSessionCookie();
    console.log(session.Data.Name);
    socket.send(JSON.stringify(session));
}

function CreatePlayer() {
  var message = {};
  message.Info = TypeCreatePlayer; 
  message.Data = {};
  message.Data.Name = doc_playername.value;
  console.log(message.Data.Name);
 socket.send(JSON.stringify(message));
} 

function CreateRoom() {
  var message = {};
  message.Info = TypeCreateRoom; 
  message.Data = {};
  message.Data.Name = doc_roomname.value;
  socket.send(JSON.stringify(message));
}


function joinRoom(name) {
  var message = {};
  message.Info = TypeJoinRoom; 
  message.Data = {};
  message.Data.Name = name;
  socket.send(JSON.stringify(message));
}


function getSessionCookie() {
    var name = "Player"+ "=";
    var decodedCookie = decodeURIComponent(document.cookie);
    var ca = decodedCookie.split(';');
    for(var i = 0; i <ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) === ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) === 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}

socket.onmessage = function (msg) {
  var json = JSON.parse(msg.data);
  console.log(json);
  switch (json.Info) {
    case TypePlayerInfo:
      player_name = json.Data.Name;
      doc_playername.value = player_name;
      doc_playername.disabled = true;
      var expires = "";
      var date = new Date();
      date.setDate(date.getDate() + 1);
      expires = "expires=" + date.toUTCString();
      document.cookie = "Player"+"=" + json.Data.Name + "; " + expires + "; path=/";
      break;
    case TypeError:
      window.alert(json.Data.Name);
      break;
    case TypeRoomInfo:
      room_name = json.Data.Name;
      doc_roomname.value = room_name;
      doc_roomname.disabled = true;
      break;
    case TypeRooms:
      ShowRooms(json)
      break;
    case TypeGameInfo:
      context.clearRect(0, 0, canvas.width, canvas.height);
      drawLines(10, lineColor);
      console.log(json.Data);
      for (let key in json.Data.Moves) {
        console.log("asdasd");
        var move = json.Data.Moves[key];
        drawPlayingPiece(move.X, move.Y, move.PlayerIndex);
      }
      drawLines(10, lineColor);  
      break;
    case TypeGameMove:
      var move = json.Data;
      drawPlayingPiece(move.X, move.Y, move.PlayerIndex);
      drawLines(10, lineColor);      
      break;
    case TypeGameOver:
      window.alert("Winner is: " + json.Data.Name);
      doc_roomname.value = "";
      doc_roomname.disabled = false;
      break;
    case TypeGameDraw:
      window.alert("Draw! Try harder next time");
      doc_roomname.value = "";
      doc_roomname.disabled = false;
      break;
    break;
  }
}

function ShowRooms(json) {
  doc_rooms.innerHTML = "";
  console.log(json.Info);
  console.log(json.Data);
  for (let key in json.Data) {
    addNewRoom(json.Data[key], doc_rooms);
  }
}

function addNewRoom(room, tbody) {
    newRow = tbody.insertRow(tbody.rows.length);
    let cols = "";
    cols += "<td>" + room.Name + "</td>";
    cols += `<td><button type="button" class="btn btn-primary" onclick="joinRoom('${room.Name}')">Join Room</button></td>`;
    newRow.innerHTML = cols;
}


var canvas = document.getElementById('tictactoe');
var context = canvas.getContext('2d');

var canvasSize = 500;
var sectionSize = canvasSize / 3;
canvas.width = canvasSize;
canvas.height = canvasSize;
context.translate(0.5, 0.5);


function addPlayingPiece (mouse) {
  var xCordinate;
  var yCordinate;

  for (var x = 0;x < 3;x++) {
    for (var y = 0;y < 3;y++) {
      xCordinate = x * sectionSize;
      yCordinate = y * sectionSize;

      if (mouse.x >= xCordinate && mouse.x <= xCordinate + sectionSize &&
          mouse.y >= yCordinate && mouse.y <= yCordinate + sectionSize) {

        var msg = {}
        msg.Info = TypePlayerMove;
        msg.Data = {}
        msg.Data.X = x;
        msg.Data.Y = y;
        socket.send(JSON.stringify(msg));
      }
    }
  }
}

function drawPlayingPiece (x, y, player) {
  var xCordinate = x * sectionSize;
  var yCordinate = y * sectionSize;

  clearPlayingArea(xCordinate, yCordinate);
  if (player === 1) {
    drawX(xCordinate, yCordinate);
  } else {
    drawO(xCordinate, yCordinate);
  }
}


function clearPlayingArea (xCordinate, yCordinate) {
  context.fillStyle = "#fff";
  context.fillRect(
    xCordinate,
    yCordinate,
    sectionSize,
    sectionSize
  );
}

function drawO (xCordinate, yCordinate) {
  var halfSectionSize = (0.5 * sectionSize);
  var centerX = xCordinate + halfSectionSize;
  var centerY = yCordinate + halfSectionSize;
  var radius = (sectionSize - 100) / 2;
  var startAngle = 0 * Math.PI;
  var endAngle = 2 * Math.PI;

  context.lineWidth = 10;
  context.strokeStyle = "#01bBC2";
  context.beginPath();
  context.arc(centerX, centerY, radius, startAngle, endAngle);
  context.stroke();
}

function drawX (xCordinate, yCordinate) {
  context.strokeStyle = "#f1be32";
  context.beginPath();
  var offset = 50;
  context.moveTo(xCordinate + offset, yCordinate + offset);
  context.lineTo(xCordinate + sectionSize - offset, yCordinate + sectionSize - offset);

  context.moveTo(xCordinate + offset, yCordinate + sectionSize - offset);
  context.lineTo(xCordinate + sectionSize - offset, yCordinate + offset);

  context.stroke();
}

function drawLines (lineWidth, strokeStyle) {
  var lineStart = 4;
  var lineLenght = canvasSize - 5;
  context.lineWidth = lineWidth;
  context.lineCap = 'round';
  context.strokeStyle = strokeStyle;
  context.beginPath();

  /*
   * Horizontal lines
   */
  for (var y = 1;y <= 2;y++) {
    context.moveTo(lineStart, y * sectionSize);
    context.lineTo(lineLenght, y * sectionSize);
  }

  /*
   * Vertical lines
   */
  for (var x = 1;x <= 2;x++) {
    context.moveTo(x * sectionSize, lineStart);
    context.lineTo(x * sectionSize, lineLenght);
  }

  context.stroke();
}

drawLines(10, lineColor);

function getCanvasMousePosition (event) {
  var rect = canvas.getBoundingClientRect();

  return {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top
  }
}

canvas.addEventListener('mouseup', function (event) {
  var canvasMousePosition = getCanvasMousePosition(event);
  addPlayingPiece(canvasMousePosition);
  drawLines(10, lineColor);
});
