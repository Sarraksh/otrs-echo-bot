package CLILogger

import (
	"fmt"
	"log"
)

type CLILogger struct {
	Module string // Name of module that uses logger.
}

func NewDefault(logFilePath string) CLILogger {
	return CLILogger{}
}

func (cl *CLILogger) SetModuleName(name string) {
	cl.Module = name
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
