package main

import (
	"context"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/ClientProvider"
	"github.com/Sarraksh/otrs-echo-bot/ClientProvider/basicCilent"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider/SQLite3"
	"github.com/Sarraksh/otrs-echo-bot/OTRSProvider"
	"github.com/Sarraksh/otrs-echo-bot/OTRSProvider/basicOTRS"
	"github.com/Sarraksh/otrs-echo-bot/RESTProvider"
	"github.com/Sarraksh/otrs-echo-bot/RESTProvider/echoREST"
	"github.com/Sarraksh/otrs-echo-bot/TelegramProvider"
	"github.com/Sarraksh/otrs-echo-bot/TelegramProvider/tgbotapiProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/config"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"github.com/Sarraksh/otrs-echo-bot/common/logger/zapLogger"
	"github.com/Sarraksh/otrs-echo-bot/event"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

const version string = `0.1.0.0`

func main() {
	// Get and save program directory.
	programDirectory, err := getProgramDir()
	if err != nil {
		log.Println("unable get program directory")
	}
	var logModule logger.Logger
	logModule = zapLogger.NewDefault(filepath.Join(programDirectory, "log", "otrs-echo-bot"))
	logModule = logModule.SetModuleName("MainRoutine")

	// Print in log start info
	logModule.Info("====================================================")
	logModule.Info("Start OTRS_Echo_bot")
	logModule.Info(fmt.Sprintf("Program version '%v'", version))
	logModule.Info(fmt.Sprintf("Go version '%v'", runtime.Version()))
	logModule.Info("====================================================")

	// TODO Read configuration from file.
	// Read configuration from file.
	var conf config.Config

	// Declare module variables.
	var (
		DBModule       DBProvider.DBProvider
		OTRSModule     OTRSProvider.OTRSProvider
		TelegramModule TelegramProvider.TelegramProvider
		ClientModule   ClientProvider.ClientProvider
		EventProcessor event.Processor
		RESTModule     RESTProvider.RESTProvider
	)

	// Define types for module variables.
	DBModule = new(SQLite3.DB)
	OTRSModule = new(basicOTRS.BasicOTRS)
	TelegramModule = new(tgbotapiProvider.TelegramModule)
	ClientModule = new(basicCilent.BasicClient)
	RESTModule = new(echoREST.EchoREST)

	// Initialise modules.
	err = initialiseModules(
		&conf,
		logModule,
		programDirectory,
		&DBModule,
		&OTRSModule,
		&TelegramModule,
		&ClientModule,
		&EventProcessor,
		&RESTModule,
	)
	if err != nil {
		logModule.Error(fmt.Sprintf("Modules initialisation failed - '%v'. Stop OTRS_Echo_bot", err))
		return
	}

	// Initialise group for services with context.
	logModule.Debug("Initialise group of services.")
	ctxGroup, cancelGroup := context.WithCancel(context.Background())
	defer cancelGroup()
	group, ctxGroup := errgroup.WithContext(ctxGroup)

	// Wait fot sigterm.
	group.Go(func() error {
		logModule.Debug(fmt.Sprintf("Start wait for sigterm."))
		err := Sigterm(ctxGroup, cancelGroup)
		logModule.Debug(fmt.Sprintf("Stop wait for sigterm with error '%v'.", err))
		return err
	})

	// Start telegram update listener.
	group.Go(func() error {
		logModule.Debug(fmt.Sprintf("Start telegram update listener."))
		err := TelegramModule.UpdateListener(ctxGroup, cancelGroup)
		logModule.Debug(fmt.Sprintf("Stop telegram update listener with error '%v'.", err))
		return err
	})

	// Start HTTP listener.
	group.Go(func() error {
		logModule.Debug(fmt.Sprintf("Start HTTP listener."))
		err := RESTModule.Listen(ctxGroup, cancelGroup)
		logModule.Debug(fmt.Sprintf("Stop HTTP listener with error '%v'.", err))
		return err
	})

	// Wait for group routines in group.
	logModule.Debug("Start wait group of services.")
	err = group.Wait()
	logModule.Error("Stop group of services.")
	if err != nil {
		logModule.Error(fmt.Sprintf("Stop group of services with error '%v'.", err))
	}

	logModule.Info("Stop OTRS_Echo_bot")
}

// Get program directory by absolute file path from arguments
// or by working directory and relative file path.
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

func initialiseModules(
	conf *config.Config,
	logModule logger.Logger,
	programDirectory string,
	DBModule *DBProvider.DBProvider,
	OTRSModule *OTRSProvider.OTRSProvider,
	TelegramModule *TelegramProvider.TelegramProvider,
	ClientModule *ClientProvider.ClientProvider,
	EventProcessor *event.Processor,
	RESTModule *RESTProvider.RESTProvider,
) error {

	logModule.Debug("Start module initialisation sequence")

	logModule.Debug("Initialise DB module")
	err := (*DBModule).Initialise(logModule, programDirectory)
	if err != nil {
		logModule.Error(fmt.Sprintf("Initialise DB module  failed - '%v'", err))
		return err
	}

	logModule.Debug("Initialise OTRS module")
	(*OTRSModule).Initialise(logModule, conf.OTRS)

	logModule.Debug("Initialise Telegram module")
	err = (*TelegramModule).Initialise(conf.Telegram.Token, logModule, DBModule)
	if err != nil {
		logModule.Error(fmt.Sprintf("Initialise Telegram module  failed - '%v'", err))
		return err
	}

	logModule.Debug("Initialise Client module")
	(*ClientModule).Initialise(DBModule, logModule)

	logModule.Debug("Initialise Event processor")
	EventProcessor = &event.Processor{
		DB:       DBModule,
		OTRS:     OTRSModule,
		Client:   ClientModule,
		Telegram: TelegramModule,
		Log:      logModule.SetModuleName("EventProcessor"),
	}

	(*RESTModule).Initialise(logModule, DBModule)
	(*RESTModule).PrepareListener(EventProcessor)

	logModule.Debug("Module initialisation sequence complete")
	return nil
}
