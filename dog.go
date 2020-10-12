package dog

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

//go:generate bpak --package "$GOPACKAGE" --file bpak.go --root bpak_root

var upgrader = websocket.Upgrader{} // use default options
var homeTemplate *template.Template

func init() {
	b, err := bpakGet("index.template.html")
	if err != nil {
		panic(err)
	}
	if !utf8.Valid(b) {
		panic(fmt.Errorf("index.template.html is not valid utf8"))
	}
	homeTemplate = template.Must(template.New("").Parse(string(b)))
}

type IndexTemplate struct {
	AppBase  string
	AppPath  string
	HostName string
	WsAddr   string
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, IndexTemplate{
		AppBase:  filepath.Base(os.Args[0]),
		AppPath:  filepath.Dir(os.Args[0]),
		HostName: r.Host,
		WsAddr:   "ws://" + r.Host + "/wsconnect",
	})
}

func wsconnect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func ListenAndServe(host string) error {
	http.Handle("/static/", http.FileServer(&bpakFileSystem{}))
	http.HandleFunc("/wsconnect", wsconnect)
	http.HandleFunc("/", home)
	return http.ListenAndServe(host, nil)
}
