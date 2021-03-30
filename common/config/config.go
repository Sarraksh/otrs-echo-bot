package config

import (
	"encoding/json"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/encryption"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Combine all available options.
type Config struct {
	OTRS     OTRSConf
	Telegram TelegramConf
}

// Options for OTRS module.
type OTRSConf struct {
	Host            string // Host on which OTRS is located.
	TicketURLPrefix string // Prefix for ticket URL for browser (include port if not default).
	API             OTRSAPI
}

// OTRS API configuration.
type OTRSAPI struct {
	Login                   string // Login for API.
	Password                string // Password for API.
	Protocol                string // Protocol over which the API is available. http or https.
	Port                    string // Port over which the API is available.
	InsecureConnection      bool   // If true allow insecure connections to API.
	GetTicketDetailListPath string // Get ticket details.
}

// Used for encryption storage
type SensitiveData struct {
	OTRSLogin     string
	OTRSPassword  string
	TelegramToken string
}

// Options for Telegram module.
type TelegramConf struct {
	Token string // Token from @BotFather.
}

func Initialise(configFileName, encryptedFileName, programDirectory string, logModule logger.Logger) (Config, error) {
	logModule.SetModuleName("Configuration")
	config, err := readFromYaml(configFileName, programDirectory, logModule)
	if err != nil {
		logModule.Error(fmt.Sprintf("Can't read from file '%v' - '%v'", configFileName, err))
		return Config{}, err
	}

	config, err = mergeWitEncryptedData(config, encryptedFileName, programDirectory)
	if err != nil {
		logModule.Error(fmt.Sprintf("While merge with encrypted data - '%v'", err))
		return Config{}, err
	}

	// Find missing mandatory fields, write into log if some fined and close with program error.
	if !isMandatoryFieldsPresent(config, logModule) {
		return Config{}, errors.ErrMandatoryFieldMissing
	}

	return config, nil
}

func readFromYaml(configFileName, programDirectory string, logModule logger.Logger) (Config, error) {
	fullPath := filepath.Join(programDirectory, configFileName)
	configFromFile, err := readConfigFromYAMLFile(fullPath)
	if err != nil {
		logModule.Error(fmt.Sprintf("While read config from file '%v' - '%v'", fullPath, err))
		return Config{}, err
	}
	return configFromFile, nil
}

// Extract configuration file and unmarshall collected data into config variable.
func readConfigFromYAMLFile(cfgFilePath string) (Config, error) {
	var fileConfig Config
	file, err := os.Open(cfgFilePath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(data, fileConfig)
	if err != nil {
		return Config{}, err
	}
	return fileConfig, err
}

func mergeWitEncryptedData(conf Config, encryptedFile, programDirectory string) (Config, error) {
	encryptionFileFullPath := filepath.Join(encryptedFile, programDirectory)
	sensitiveData, err := readEncryptedDataFromFile(encryptionFileFullPath)
	if err != nil {
		return Config{}, err
	}

	// Merge data from  config and previously encrypted data.
	if conf.OTRS.API.Login == "" {
		conf.OTRS.API.Login = sensitiveData.OTRSLogin
	} else {
		sensitiveData.OTRSLogin = conf.OTRS.API.Login
	}
	if conf.OTRS.API.Password == "" {
		conf.OTRS.API.Password = sensitiveData.OTRSPassword
	} else {
		sensitiveData.OTRSPassword = conf.OTRS.API.Password
	}
	if conf.Telegram.Token == "" {
		conf.Telegram.Token = sensitiveData.TelegramToken
	} else {
		sensitiveData.TelegramToken = conf.Telegram.Token
	}

	// Check all sensitive data existence.
	err = checkSensitiveDataProvided(sensitiveData)
	if err != nil {
		return Config{}, err
	}

	err = writeEncryptedDataIntoFile(encryptionFileFullPath, sensitiveData)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}

func readEncryptedDataFromFile(encryptionFileFullPath string) (SensitiveData, error) {
	file, errF := os.Open(encryptionFileFullPath)
	if errF != nil {
		return SensitiveData{}, errF
	}
	defer file.Close()
	dataEncrypted, erF := ioutil.ReadAll(file)
	if erF != nil {
		return SensitiveData{}, erF
	}
	dataJSON, err := encryption.Decrypt(string(dataEncrypted))
	if err != nil {
		return SensitiveData{}, err
	}
	var data SensitiveData
	err = json.Unmarshal(dataJSON, &data)
	if err != nil {
		return SensitiveData{}, err
	}

	return data, nil
}

func writeEncryptedDataIntoFile(encryptionFileFullPath string, sensitiveData SensitiveData) error {
	dataJSON, err := json.Marshal(sensitiveData)
	if err != nil {
		return err
	}
	dataEncrypted := encryption.Encrypt(dataJSON)

	file, errF := os.OpenFile(encryptionFileFullPath, os.O_CREATE|os.O_WRONLY, 0660)
	if errF != nil {
		return errF
	}
	defer file.Close()
	dataLen, errF := file.WriteString(dataEncrypted)
	if errF != nil {
		return errF
	}
	errF = file.Truncate(int64(dataLen))
	if errF != nil {
		return errF
	}
	return nil
}

func checkSensitiveDataProvided(sensData SensitiveData) error {
	switch {
	case sensData.OTRSLogin == "":
		return errors.ErrOTRSLoginNotProvided
	case sensData.OTRSPassword == "":
		return errors.ErrOTRSPasswordNotProvided
	case sensData.TelegramToken == "":
		return errors.ErrTelegramTokenNotProvided
	}
	return nil
}

func isMandatoryFieldsPresent(config Config, logModule logger.Logger) bool {
	var allFieldsPresent bool = true
	if config.OTRS.Host == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.Host' is mandatory but not present")
	}
	if config.OTRS.TicketURLPrefix == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.TicketURLPrefix' is mandatory but not present")
	}
	if config.OTRS.API.Login == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.API.Login' is mandatory but not present")
	}
	if config.OTRS.API.Password == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.API.Password' is mandatory but not present")
	}
	if config.OTRS.API.Protocol == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.API.Protocol' is mandatory but not present")
	}
	if config.OTRS.API.Port == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.API.Port' is mandatory but not present")
	}
	if config.OTRS.API.GetTicketDetailListPath == "" {
		allFieldsPresent = false
		logModule.Error("Option 'OTRS.API.GetTicketDetailListPath' is mandatory but not present")
	}
	if config.Telegram.Token == "" {
		allFieldsPresent = false
		logModule.Error("Option 'Telegram.Token' is mandatory but not present")
	}

	return allFieldsPresent
}