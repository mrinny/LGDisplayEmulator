package lgdisplayapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/mrinny/LGDisplayEmulator/internal/displaymanager"
)

func New() *LGDisplayAPI {
	return &LGDisplayAPI{}
}

type LGDisplayAPI struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	active bool
	conn   net.Listener
	host   string
	port   int
	dm     *displaymanager.DisplayManager
}

func (l *LGDisplayAPI) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.active {
		return fmt.Errorf("service already started")
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	go l.service()
	return nil
}

func (l *LGDisplayAPI) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active {
		return fmt.Errorf("service is not started")
	}
	err := l.conn.Close()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	l.cancel()
	return nil
}

func (l *LGDisplayAPI) Running() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.active
}

func (l *LGDisplayAPI) service() {
	var err error
	l.conn, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", l.host, l.port))
	if err != nil {
		slog.Error("failed to start tcp server")
		return
	}
	defer func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.active = false
	}()
	l.mu.Lock()
	l.active = true
	l.mu.Unlock()
	for {
		select {
		case <-l.ctx.Done():
			return
		default:
		}
		c, err := l.conn.Accept()
		if err != nil {
			slog.Warn(err.Error())
			continue
		}
		if c != nil {
			go l.handleClient(c)
		}
	}

}

func (l *LGDisplayAPI) handleClient(clientConn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(clientConn)
	commandReader := newCommandReader(clientConn)
	for {
		cmd, err := commandReader.Next()
		if err != nil {
			slog.Error("failed to read command", "error", err)
			return
		}
		switch cmd.Cmd {
		case LGPower:
			switch cmd.Value {
			case PowerOff:
			case PowerOn:
			case PowerRestart:
			default:
				slog.Warn("invallid powerstate")
			}
		default:
			slog.Warn("unsupported command")
		}
	}
}
