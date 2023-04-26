package internal

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func getServer(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(h)

	return s
}

func getClient(t *testing.T, h http.Handler, url string) *websocket.Conn {
	t.Helper()

	ws, _, err := websocket.DefaultDialer.Dial(strings.Replace(url, "http", "ws", 1), nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s: %v", url, err)
	}

	return ws
}

func sendMsg(t *testing.T, ws *websocket.Conn, msg string) {
	t.Helper()

	msg = strings.TrimSpace(msg)

	if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("%v", err)
	}
}

func getMsg(t *testing.T, ws *websocket.Conn) string {
	t.Helper()

	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	return string(msg)
}

func TestGetRoom(t *testing.T) {
	tests := []struct {
		name string
		want *Room
	}{
		{
			name: "GetRoom",
			want: &Room{
				clients: make(map[*websocket.Conn]*client),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRoom(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoom_ServeHTTP(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "valid string",
			message: "hello world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := GetRoom()
			s := getServer(t, rm)
			defer s.Close()

			client_1 := getClient(t, rm, s.URL)
			client_2 := getClient(t, rm, s.URL)

			sendMsg(t, client_1, tt.message)

			reply_1 := getMsg(t, client_1)
			reply_2 := getMsg(t, client_2)

			if reply_1 != reply_2 {
				t.Errorf("replys mismatched: got %v, want %v", reply_1, reply_2)
			}

			if reply_1 != tt.message {
				t.Errorf("invalid reply: got %v, want %v", reply_1, tt.message)
			}
		})
	}
}
