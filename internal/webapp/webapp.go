package webapp

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mrinny/LGDisplayEmulator/internal/domain"
)

//go:embed html templates
var staticDir embed.FS
var contentFS, _ = fs.Sub(staticDir, "html")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebApp struct {
	server  *http.Server
	mux     *http.ServeMux
	mu      sync.Mutex
	running bool
	hub     *Hub
}

func New(hub *Hub) *WebApp {
	mux := http.NewServeMux()
	wa := WebApp{
		server: &http.Server{
			Addr:    ":3000",
			Handler: mux,
		},
		mux: mux,
		hub: hub,
	}
	mux.Handle("/", WithLogging(http.FileServer(http.FS(contentFS))))
	mux.Handle("/ws", WithLogging(http.HandlerFunc(wa.servews)))
	return &wa
}

func (wa *WebApp) Start() error {
	go wa.Run()
	return nil
}

func (wa *WebApp) Stop() error {
	wa.mu.Lock()
	if !wa.running {
		return fmt.Errorf("web app was not running")
	}
	wa.mu.Unlock()
	err := wa.server.Shutdown(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (wa *WebApp) Run() {
	wa.mu.Lock()
	wa.running = true
	wa.mu.Unlock()
	defer func() {
		wa.mu.Lock()
		defer wa.mu.Unlock()
		wa.running = false
	}()
	err := wa.server.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}

}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(rw, r) // serve the original request

		duration := time.Since(start)

		// log request details
		slog.Info("webapp",
			"uri", uri,
			"method", method,
			"duration", duration,
		)
	}
	return http.HandlerFunc(logFn)
}

func (wa *WebApp) servews(w http.ResponseWriter, r *http.Request) {
	slog.Info("starting websocket upgrade")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed upgading connection", "error", err)
		return
	}
	cl := &Client{
		conn: conn,
		send: make(chan []byte, 25),
		hub:  wa.hub,
	}

	go cl.writePump()
	go cl.readPump()
}

func renderDisplayTemplate(dp domain.LGDisplay) []byte {
	tmpl, err := template.ParseFS(staticDir, "templates/display.html")
	if err != nil {
		slog.Error(err.Error())
		return []byte{}
	}
	var result bytes.Buffer
	err = tmpl.Execute(&result, dp)
	if err != nil {
		slog.Error(err.Error())
		return []byte{}
	}
	return result.Bytes()
}
