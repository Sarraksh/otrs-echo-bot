package basicOTRS

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/OTRSProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"io/ioutil"
	"net/http"
)

type BasicOTRS struct {
	URLFormat  string
	HTTPClient *http.Client
	Log        logger.Logger
}

func New(logger logger.Logger) BasicOTRS {
	logger.SetModuleName("OTRSProvider")
	return BasicOTRS{Log: logger}
}

// Initialise transport and set behavior for insecure connections.
func (bo *BasicOTRS) SetTransport(InsecureConnection bool) {
	// Avoid insecure connection error.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: InsecureConnection},
	}
	bo.HTTPClient = &http.Client{Transport: tr}
	bo.Log.Debug(fmt.Sprintf("Transport initialised. Set InsecureConnection as '%v'", InsecureConnection))
}

// Set protocol, URL and credentials for OTRS instance and store as format fo fmt.Sprintf .
func (bo *BasicOTRS) SetURLFormat(protocol, URL, login, password string) {
	bo.URLFormat = urlFormat(protocol, URL, login, password)
	maskedURLString := urlFormat(protocol, URL, `*********`, `*********`)
	bo.Log.Debug(fmt.Sprintf("Set URLFormat - '%v'", maskedURLString))
}

func (bo *BasicOTRS) GetTicketDetails(ticketID string) (OTRSProvider.TicketOTRS, error) {
	bo.Log.Debug(fmt.Sprintf("Start GetTicketDetails sequence for '%v'", ticketID))
	defer bo.Log.Debug(fmt.Sprintf("Stop  GetTicketDetails sequence for '%v'", ticketID))

	requestURL := fmt.Sprintf(bo.URLFormat, ticketID)
	response, err := bo.HTTPClient.Get(requestURL)
	if err != nil {
		bo.Log.Error(fmt.Sprintf("GET request '%+v'", err))
		return OTRSProvider.TicketOTRS{}, err // TODO - make less sensitive for errors
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		bo.Log.Error(fmt.Sprintf("ReadAll from responce - '%+v'", err))
		return OTRSProvider.TicketOTRS{}, err // TODO - make less sensitive for errors
	}
	var ticketsFromJSON OTRSProvider.TicketsFromJSON
	bo.Log.Debug(fmt.Sprintf("Respose body - '%+v'", string(body)))
	err = json.Unmarshal(body, &ticketsFromJSON)
	if err != nil {
		bo.Log.Error(fmt.Sprintf("Unmarshal error - '%+v'", err))
		return OTRSProvider.TicketOTRS{}, err // TODO - make less sensitive for errors
	}

	return ticketsFromJSON.Ticket[0], nil
}

func urlFormat(protocol, URL, login, password string) string {
	return fmt.Sprint(
		protocol,
		`://`,
		URL,
		`/otrs/nph-genericinterface.pl/Webservice/GenericTicketConnectorREST/TicketGet/`,
		`%s`,
		`?UserLogin=`,
		login,
		`&Password=`,
		password,
	)
}