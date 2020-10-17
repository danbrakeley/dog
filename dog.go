package dog

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"
)

//go:generate bpak --package "$GOPACKAGE" --file bpak.go --root bpak_root

type Dog struct {
	homeTemplate *template.Template
	minLevel     Level
	wsRouter     *WsRouter
	server       http.Server
}

type LogLine struct {
	Time   time.Time              `json:"timestamp"`
	Level  string                 `json:"level"`
	Msg    string                 `json:"msg"`
	Fields map[string]interface{} `json:"fields"`
}

func Create(host string) (*Dog, error) {
	index := "index.template.html"
	b, err := bpakGet(index)
	if err != nil {
		return nil, fmt.Errorf("failure loading %s: %w", index, err)
	}
	if !utf8.Valid(b) {
		return nil, fmt.Errorf("contents of %s is not valid utf8", index)
	}

	home, err := template.New("").Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("failure parsing %s: %w", index, err)
	}

	d := &Dog{
		homeTemplate: home,
		minLevel:     Info,
		wsRouter:     NewWsRouter(),
	}

	m := http.NewServeMux()
	m.Handle("/static/", http.FileServer(&bpakFileSystem{}))
	m.HandleFunc("/ws", d.wsRouter.ServeWs)
	m.HandleFunc("/", d.handleHome)
	d.server = http.Server{Addr: host, Handler: m}

	// server thread
	go func() {
		d.wsRouter.Start()
		d.server.RegisterOnShutdown(func() {
			d.wsRouter.BeginShutdown()
		})

		fmt.Printf("--- [server thread] http server listening on %s\n", host)

		err := d.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "[dog] http server thread died with error: %v\n", err)
			return
		}

		fmt.Println("--- [server thread] waiting for websocket clients to finish")
		d.wsRouter.WaitForShutdown()

		fmt.Printf("--- [server thread] http server shutdown\n")
	}()

	return d, nil
}

// Close is part of the Logger interface
func (d *Dog) Close() {
	fmt.Println("===== Dog Close Begin")
	d.server.Shutdown(context.Background())
	fmt.Println("===== Dog Close End")
}

// SetMinLevel is part of the Logger interface
func (d *Dog) SetMinLevel(level Level) Logger {
	d.minLevel = level
	return d
}

// Log is part of the Logger interface
func (d *Dog) Log(level Level, msg string, fields ...Fielder) Logger {
	now := time.Now()

	if level < d.minLevel {
		return d
	}

	rawFields := make(map[string]interface{})
	for _, f := range fields {
		field := f.Field()
		rawFields[field.Name] = field.Raw
	}

	b, err := json.Marshal(LogLine{Time: now, Level: level.String(), Msg: msg, Fields: rawFields})
	if err != nil {
		panic(fmt.Errorf("error marshalling log line to json: %w", err))
	}

	fmt.Printf("## %s\n", string(b))

	d.wsRouter.Broadcast(b)

	return d
}

// Transient is part of the Logger interface
func (d *Dog) Transient(format string, a ...Fielder) Logger {
	d.Log(Transient, format, a...)
	return d
}

// Verbose is part of the Logger interface
func (d *Dog) Verbose(format string, a ...Fielder) Logger {
	d.Log(Verbose, format, a...)
	return d
}

// Info is part of the Logger interface
func (d *Dog) Info(format string, a ...Fielder) Logger {
	d.Log(Info, format, a...)
	return d
}

// Warning is part of the Logger interface
func (d *Dog) Warning(format string, a ...Fielder) Logger {
	d.Log(Warning, format, a...)
	return d
}

// Error is part of the Logger interface
func (d *Dog) Error(format string, a ...Fielder) Logger {
	d.Log(Error, format, a...)
	return d
}

// Fatal is part of the Logger interface
func (d *Dog) Fatal(format string, a ...Fielder) {
	d.Log(Fatal, format, a...)
	d.Close()
	os.Exit(1)
}

type IndexTemplate struct {
	AppBase  string
	AppPath  string
	HostName string
	WsAddr   string
}

func (d *Dog) handleHome(w http.ResponseWriter, r *http.Request) {
	d.homeTemplate.Execute(w, IndexTemplate{
		AppBase:  filepath.Base(os.Args[0]),
		AppPath:  filepath.Dir(os.Args[0]),
		HostName: r.Host,
		WsAddr:   "ws://" + r.Host + "/ws",
	})
}
