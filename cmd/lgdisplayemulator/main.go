package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrinny/LGDisplayEmulator/internal/lgdisplayemulator"
)

func main() {

	var err error

	service := lgdisplayemulator.New()
	err = service.Start()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	fmt.Println("\nshutting down")
	err = service.Stop()
	if err != nil {
		slog.Error(err.Error())
	}
	for i := 10; i > 0; i-- {
		fmt.Printf(" %d", i)
		if !service.Running() {
			fmt.Println("\n gracefully shutdown")
			os.Exit(0)
		}
		time.Sleep(time.Second)
	}
	fmt.Println("\nfailed to cleanly close the connections")
	os.Exit(1)
}
