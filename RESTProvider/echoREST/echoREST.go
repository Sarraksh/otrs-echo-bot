package echoREST

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"github.com/Sarraksh/otrs-echo-bot/event"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"strconv"
	"time"
)

type EchoREST struct {
	Instance *echo.Echo
	Log      logger.Logger
	DB       *DBProvider.DBProvider
}

// For marshal response to OTRS.
type ResponseToOTRS struct {
	TicketID string
}

// Initialise echoREST module.
func (eREST *EchoREST) Initialise(logger logger.Logger, db *DBProvider.DBProvider) {
	logger.SetModuleName("echoREST")
	eREST.Log = logger
	eREST.DB = db
}

// Prepare http listener.
func (eREST *EchoREST) PrepareListener(eventProcessor *event.Processor) {
	eREST.Log.Debug(fmt.Sprintf("Start REST instance initialisation"))

	e := echo.New()             // Echo instance
	e.Use(middleware.Logger())  // Middleware
	e.Use(middleware.Recover()) // Middleware

	// Handle requests with new event from OTRS.
	pathNewTicket := "newticket"
	e.POST(fmt.Sprint("/", pathNewTicket), func(c echo.Context) error {
		id := c.FormValue("id")
		eREST.Log.Debug(fmt.Sprintf("Recived new event from '%+v' with ticket id /'%+v'", pathNewTicket, id))

		// Parse text to integer end response with error if fail.
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.Response().Header().Set("ResponseSuccess", "0")                             // Needed by OTRS invoker
			c.Response().Header().Set("ResponseErrorMessage", "Invalid id field content") // Needed by OTRS invoker
			responseBody := ResponseToOTRS{
				TicketID: id,
			}
			data, err := json.Marshal(&responseBody)
			if err != nil {
				eREST.Log.Error(fmt.Sprintf("Can't marshal body for response to OTRS - '%+v'", err))
			}
			eREST.Log.Debug(fmt.Sprintf("Send response with body - '%+v'", string(data)))
			return c.JSONBlob(http.StatusBadRequest, data)
		}

		// Save data into DB.
		// If event already exist write nothing.
		db := *eREST.DB
		eventExist, err := db.OTRSEventIsExistsWithTicketIDAndType(int64(idInt), pathNewTicket)
		if !eventExist {
			err = db.OTRSEventCreateNew(pathNewTicket, pathNewTicket, int64(idInt))
		}

		// Invoke event processor
		eventProcessor.ProcessEvent()

		c.Response().Header().Set("ResponseSuccess", "1")     // Needed by OTRS invoker
		c.Response().Header().Set("ResponseErrorMessage", "") // Needed by OTRS invoker
		responseBody := ResponseToOTRS{
			TicketID: id,
		}
		data, err := json.Marshal(&responseBody)
		if err != nil {
			eREST.Log.Error(fmt.Sprintf("Can't marshal body for response to OTRS - '%+v'", err))
		}
		eREST.Log.Debug(fmt.Sprintf("Send response with body - '%+v'", string(data)))
		return c.JSONBlob(http.StatusOK, data)
	},
	) // Route

	eREST.Log.Debug(fmt.Sprintf("REST instance initialised"))
	eREST.Instance = e
}

//
func (eREST *EchoREST) Listen(ctx context.Context, cancel context.CancelFunc) {
	go listenerWrapper(eREST.Instance, cancel, eREST.Log)
	eREST.Log.Debug("Listener started")
	select {
	case <-ctx.Done():
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		err := eREST.Instance.Shutdown(ctxShutdown)
		if err != nil {
			eREST.Log.Error("Can't gracefully interrupt listener by context done.")

		} else {
			eREST.Log.Debug("Listener interrupted by context done.")
		}
		return
	}
}

// Start listener and cancel into context channel if sever stopped.
func listenerWrapper(instance *echo.Echo, cancel context.CancelFunc, logger logger.Logger) {
	instance.Logger.Fatal(instance.Start(":1323")) // Start server
	logger.Debug("Listener stopped.")
	cancel()
}
