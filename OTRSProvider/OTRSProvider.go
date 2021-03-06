package OTRSProvider

import (
	"github.com/Sarraksh/otrs-echo-bot/common/config"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

type OTRSProvider interface {
	Initialise(logger logger.Logger, conf config.OTRSConf)
	GetTicketDetails(ticketID string) (TicketOTRS, error)
}

// Wrapper for correct unmarshall JSON. ORTS returns array of tickets.
type TicketsFromJSON struct {
	Ticket []TicketOTRS `json:"Ticket"`
}

// For ticket information received from OTRS and further processed information.
type TicketOTRS struct {
	TicketNumber string `json:"TicketNumber"` // It is returned in the field of the same name from OTRS.
	Type         string `json:"Type"`         // It is returned in the field of the same name from OTRS.
	CustomerID   string `json:"CustomerID"`   // It is returned in the field of the same name from OTRS.
	Priority     string `json:"Priority"`     // It is returned in the field of the same name from OTRS.
	Created      string `json:"Created"`      // It is returned in the field of the same name from OTRS.
	Title        string `json:"Title"`        // It is returned in the field of the same name from OTRS.
	Lock         string `json:"Lock"`         // It is returned in the field of the same name from OTRS.
	StateType    string `json:"StateType"`    // It is returned in the field of the same name from OTRS.
	URL          string // For formatted message.
}
