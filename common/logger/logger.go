package logger

// Initialized in the main function, then passed a copy to each called module.
type Logger interface {
	NewDefault(logFilePath string) Logger
	SetModuleName(name string)
	Error(message string)
	Info(message string)
	Debug(message string)
}
