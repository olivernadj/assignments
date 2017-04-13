package main

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/gorilla/websocket"
)

type msg struct {
	Services []struct {
		Service string `json:"service"`
		CacheIsOk bool `json:"cache_is_ok"`
		DbIsOk bool `json:"db_is_ok"`
		Status string `json:"status"`
		StructCacheIsOk bool `json:"struct_cache_is_ok"`
		Version string `json:"version"`
	} `json:"services"`
	ServiceMap []struct {
		Client string `json:"client"`
		Server string `json:"server"`
		RequestsPerMinutes int `json:"requests-per-minutes"`
	} `json:"service-map"`
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	panic(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	go echo(conn)
}

func echo(conn *websocket.Conn) {
	for {
		file, err := ioutil.ReadFile("./static/services.json")
    if err != nil {
      fmt.Printf("File error: %v\n", err)
    }
    var m msg
    json.Unmarshal(file, &m)
		if err := conn.WriteJSON(m); err != nil {
			fmt.Println(err)
		}
		time.Sleep(5000 * time.Millisecond)
	}
}
