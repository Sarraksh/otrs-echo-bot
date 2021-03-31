package basicOTRS

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/OTRSProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/config"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"io/ioutil"
	"net/http"
)

const ModuleName string = "OTRS Provider"

type BasicOTRS struct {
	URLFormat       string // String for fmt.Sprintf. Represent full URL to OTRS API with %s flag for ticketID.
	TicketURLPrefix string
	HTTPClient      *http.Client
	Log             logger.Logger
}

func (bo *BasicOTRS) Initialise(logger logger.Logger, conf config.OTRSConf) {
	bo.Log = logger.SetModuleName(ModuleName)

	// Generate and save URLFormat
	bo.URLFormat = urlFormat(
		conf.API.Protocol,
		conf.Host,
		conf.API.GetTicketDetailListPath,
		conf.API.Login,
		conf.API.Password,
	)
	maskedURLString := urlFormat(
		conf.API.Protocol,
		conf.Host,
		conf.API.GetTicketDetailListPath,
		`*********`,
		`*********`,
	)
	bo.Log.Debug(fmt.Sprintf("Set URLFormat - '%v'", maskedURLString))

	// Avoid insecure connection error if OTRS API available by http.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.API.InsecureConnection},
	}
	bo.HTTPClient = &http.Client{Transport: tr}
	bo.Log.Debug(fmt.Sprintf("Transport initialised. Set InsecureConnection as '%v'", conf.API.InsecureConnection))

	// Set TicketURLPrefix.
	bo.TicketURLPrefix = conf.TicketURLPrefix

	bo.Log.Debug("Initialisation complete")
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

	ticketDetails := ticketsFromJSON.Ticket[0]
	ticketDetails.URL = fmt.Sprint(bo.TicketURLPrefix, ticketID)
	return ticketDetails, nil
}

func urlFormat(protocol, URL, endpoint, login, password string) string {
	return fmt.Sprint(
		protocol,
		`://`,
		URL,
		endpoint,
		`%s`,
		`?UserLogin=`,
		login,
		`&Password=`,
		password,
	)
}
