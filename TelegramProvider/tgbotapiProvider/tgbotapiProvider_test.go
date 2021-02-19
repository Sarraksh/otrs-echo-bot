package tgbotapiProvider

import (
	"github.com/Sarraksh/otrs-echo-bot/common/logger/CLIlogger"
	"testing"
)

func TestUpdate(t *testing.T) {
	token := ""
	if token == "" {
		t.Errorf("Rcieved empty token")
	}

	logger := CLILogger.NewDefault()

	bot, err := New(logger, token)
	if err != nil {
		t.Errorf("Failed to create bot - '%v'", err)
	}

	Update(bot)

	t.Logf("Test complete")
}
