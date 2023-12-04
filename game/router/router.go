package router

import (
	"gofish/game/service"
	"net/http"
)

func init() {
	http.HandleFunc("/socket.io/", service.ServeWs)
}
