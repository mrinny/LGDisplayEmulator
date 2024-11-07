package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	err := Application(done)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
