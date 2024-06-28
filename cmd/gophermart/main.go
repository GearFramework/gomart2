package main

import (
	"github.com/GearFramework/gomart2/internal/gm"
	"github.com/GearFramework/gomart2/internal/gm/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

// Запуск сервиса авторизации
func run() error {
	app := gm.NewGomartApp(config.GetConfig())
	if err := app.Init(); err != nil {
		return err
	}
	gracefulStop(app.Stop)
	if err := app.Run(); err != nil {
		return err
	}
	return nil
}

func gracefulStop(stopCallback func()) {
	gracefulStopChan := make(chan os.Signal, 1)
	signal.Notify(
		gracefulStopChan,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	go func() {
		sig := <-gracefulStopChan
		stopCallback()
		log.Printf("Caught sig: %+v\n", sig)
		log.Println("Application graceful stop!")
		os.Exit(0)
	}()
}
