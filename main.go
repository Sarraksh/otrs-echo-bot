package main

import (
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"log"
	"os"
	"path/filepath"
)

func main() {
	programDirectory, err := getProgramDir()
	if err != nil {
		log.Println("unable get program directory")
	}
	var logModule logger.Logger
	logModule = logModule.NewDefault(filepath.Join(programDirectory, "log", "otrs-echo-bot"))

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
