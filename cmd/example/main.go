package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

type IndexTemplate struct {
	AppBase string
	AppPath string
	WsAddr  string
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, IndexTemplate{
		AppBase: filepath.Base(os.Args[0]),
		AppPath: filepath.Dir(os.Args[0]),
		WsAddr:  "ws://" + r.Host + "/echo",
	})
}

var homeTemplate *template.Template

func main() {
	b, err := ioutil.ReadFile("index.template.html")
	if err != nil {
		panic(err)
	}
	if !utf8.Valid(b) {
		panic(fmt.Errorf("index.template.html is not valid utf8"))
	}
	fmt.Println("before")
	homeTemplate = template.Must(template.New("").Parse(string(b)))
	fmt.Println("after")
	flag.Parse()
	log.SetFlags(0)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
