package logger

// Initialized in the main function, then passed a copy to each called module.
type Logger interface {
	SetModuleName(name string) Logger
	Error(message string)
	Info(message string)
	Debug(message string)
}
