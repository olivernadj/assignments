package main

import (
	"flag"
	"fmt"
	"time"
	"runtime"
	"os/exec"
	"io/ioutil"
	"net"
	"net/http"
	"math/rand"
	"encoding/json"
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

var (
	httpListen  = flag.String("http", "127.0.0.1:8080", "host:port to listen on")
	openBrowser = flag.Bool("openbrowser", true, "open browser automatically")
	httpAddr string
)

func main() {
	host, port, err := net.SplitHostPort(*httpListen)
	if err != nil {
		fmt.Println(err)
	}
	if host == "" {
		host = "localhost"
	}
	if host != "127.0.0.1" && host != "localhost" {
		fmt.Println(localhostWarning)
	}
	httpAddr = host + ":" + port

	http.HandleFunc("/ws", wsHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	go func() {
		url := "http://" + httpAddr
		if waitServer(url) && *openBrowser && startBrowser(url) {
			fmt.Printf("A browser window should open. If not, please visit %v\n", url)
		} else {
			fmt.Printf("Please open your web browser and visit %v\n", url)
		}
	}()
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
	for i := 1; true; i++ {
		rand.Seed(int64(i))
		index := rand.Intn(6)
		file, err := ioutil.ReadFile(fmt.Sprintf("./static/services.%d.json", index))
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

func waitServer(url string) bool {
	tries := 20
	for tries > 0 {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
		tries--
	}
	return false
}

func startBrowser(url string) bool {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

const localhostWarning = `
WARNING!  WARNING!  WARNING!

I appear to be listening on an address that is not localhost.
Anyone with access to this address and port will have access
to this machine as the user running front-end assignment test.

If you don't understand this message, hit Control-C to terminate this process.

WARNING!  WARNING!  WARNING!
`