package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/mrinny/LGDisplayEmulator/internal/lgdisplayapi"
)

func Application(c chan os.Signal) error {
	var err error
	api := lgdisplayapi.New()
	err = api.Start()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	<-c
	err = api.Stop()
	if err != nil {
		slog.Error(err.Error())
	}
	for i := 10; i > 0; i-- {
		fmt.Printf(" %d", i)
		if !api.Running() {
			fmt.Println("\n gracefully shutdown")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("failed to gracefully close the application")
}
