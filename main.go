package main

import (
	"context"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"github.com/Sarraksh/otrs-echo-bot/common/logger/zapLogger"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	programDirectory, err := getProgramDir()
	if err != nil {
		log.Println("unable get program directory")
	}
	var logModule logger.Logger
	logModule = zapLogger.NewDefault(filepath.Join(programDirectory, "log", "otrs-echo-bot"))

	logModule = logModule.SetModuleName("MainRoutine")
	logModule.Info("Stop program")
}

// Get program directory from arguments or arguments and working directory.
func getProgramDir() (string, error) {
	argumentsFilePath := os.Args[0]
	if filepath.IsAbs(argumentsFilePath) {
		return filepath.Dir(argumentsFilePath), nil
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	relativeDir := filepath.Dir(argumentsFilePath)
	return filepath.Join(workingDirectory, relativeDir), nil
}

// Handle system terminal signals.
func Sigterm(ctx context.Context, cancel context.CancelFunc) error {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case sig := <-signalChannel:
		log.Printf("Received signal: %v", sig)
		cancel()
	case <-ctx.Done():
		log.Printf("Closing signal goroutine")
		return ctx.Err()
	}
	return nil
}
