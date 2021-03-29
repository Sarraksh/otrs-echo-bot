package Formatter

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/OTRSProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"log"
	"time"
)

const OTRSLayout string = "2006-01-02 15:04:05 MST" // Time layout for parsing detailed info from OTRS.

func EventNewPlainText(ticket OTRSProvider.TicketOTRS) string {
	return fmt.Sprint( // Ticket information formatting
		"NEW ",
		ticket.CustomerID,
		"   ",
		ticket.Type,
		"\n",
		"Ticket ",
		ticket.TicketNumber,
		"\n",
		ticket.Title,
		"\n",
		ticket.URL,
	)
}

func EventReminderPlainText(ticket OTRSProvider.TicketOTRS, logger logger.Logger) string {
	logger.SetModuleName("Message formatter")
	age := ageCalculation(ticket.Created, logger)
	return fmt.Sprint( // Ticket information formatting
		"UP ",
		age,
		" мин.   ",
		ticket.CustomerID,
		"   ",
		ticket.Type,
		"\n",
		"Ticket ",
		ticket.TicketNumber,
		"\n",
		ticket.Title,
		"\n",
		ticket.URL,
	)
}

// Return current ticket age in minutes.
func ageCalculation(absoluteAge string, logger logger.Logger) string {
	created, err := time.Parse(OTRSLayout, fmt.Sprint(absoluteAge, " MSK")) // Add timezone.
	if err != nil {
		log.Println("Can' parse age")
		return "UNKNOWN"
	}

	since := time.Since(created).Minutes()
	logger.Debug(fmt.Sprintf(
		"Age calculation. String - '%+v'. Parsed time - '%+v'. Calculated duration - '%+v'.",
		absoluteAge,
		created,
		since,
	))
	return fmt.Sprint(uint64(since))
}
