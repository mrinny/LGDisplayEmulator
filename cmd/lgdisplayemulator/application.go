package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/mrinny/LGDisplayEmulator/internal/displaymanager"
	"github.com/mrinny/LGDisplayEmulator/internal/eventmessenger"
	"github.com/mrinny/LGDisplayEmulator/internal/lgdisplayapi"
	"github.com/mrinny/LGDisplayEmulator/internal/webapp"
)

func Application(c chan os.Signal) error {
	var err error

	em := eventmessenger.New()

	dm := displaymanager.New(em)

	hub := webapp.NewHub(em, dm)
	go hub.Run()

	api := lgdisplayapi.New()
	err = api.Start()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	wapp := webapp.New(hub)
	err = wapp.Start()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	<-c
	err = wapp.Stop()
	if err != nil {
		slog.Error(err.Error())
	}
	err = api.Stop()
	if err != nil {
		slog.Error(err.Error())
	}
	for i := 10; i > 0; i-- {
		fmt.Printf(" %d", i)
		if !api.Running() {
			fmt.Println("\n api gracefully shutdown")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("failed to gracefully close the application")
}
