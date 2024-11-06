package lgdisplayemulator

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
)

func New() *LGDisplayEmulator {
	return &LGDisplayEmulator{}
}

type LGDisplayEmulator struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	active bool
	conn   net.Listener
	host   string
	port   int
}

func (l *LGDisplayEmulator) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.active {
		return fmt.Errorf("service already started")
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	go l.service()
	return nil
}

func (l *LGDisplayEmulator) Stop() error {
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

func (l *LGDisplayEmulator) Running() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.active
}

func (l *LGDisplayEmulator) service() {
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

func (l *LGDisplayEmulator) handleClient(clientConn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(clientConn)

}
