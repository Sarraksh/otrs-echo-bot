package CLILogger

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"log"
)

type CLILogger struct {
	Module string // Name of module that uses logger.
}

func NewDefault() CLILogger {
	return CLILogger{}
}

func (cl CLILogger) SetModuleName(name string) logger.Logger {
	cl.Module = name
	return cl
}

func (cl CLILogger) Error(message string) {
	log.Println(formatString("ERROR", cl.Module, message))
}

func (cl CLILogger) Info(message string) {
	log.Println(formatString("INFO", cl.Module, message))
}

func (cl CLILogger) Debug(message string) {
	log.Println(formatString("DEBUG", cl.Module, message))
}

// Add formatted module name to error string.
func formatString(level, module, message string) string {
	return fmt.Sprintf("[%5s] [%25s] - '%s'", level, module, message)
}
