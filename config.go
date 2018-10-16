package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	msgBadEnv            = "empty %s_%s env"
	envPrefix            = "LAZYTRANSLATE"
	envTelegramToken     = "TG_TOKEN"
	envTelegramWhitelist = "TG_WHITELIST"
	envProxyAddress      = "PROXY_ADDR"
	envProxyUsername     = "PROXY_USER"
	envProxyPassword     = "PROXY_PASS"
	envDefaultLang       = "DEFAULT_LANG"
	envUpdateTimeout     = "UPDATE_TIMEOUT"   // in seconds
	envResponseTimeout   = "RESPONSE_TIMEOUT" // in seconds
)

type appConfig struct {
	tgToken       string
	tgWhitelist   []int
	proxyAddress  string
	proxyUser     string
	proxyPass     string
	defaultLang   string
	updateTimeout int
	ctxTimeout    time.Duration
}

func loadConfig() (*appConfig, error) {
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	// load dotenv file if present
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	config := &appConfig{}

	// mandatory first
	if config.tgToken = viper.GetString(envTelegramToken); config.tgToken == "" {
		return nil, errors.Errorf(msgBadEnv, envPrefix, envTelegramToken)
	}

	if whitelist := viper.GetString(envTelegramWhitelist); whitelist != "" {
		// check and fill ID whitelist
		var err error
		listSlice := strings.Split(whitelist, ",")
		config.tgWhitelist = make([]int, len(whitelist), len(whitelist))
		for i := range listSlice {
			config.tgWhitelist[i], err = strconv.Atoi(listSlice[i])
			if err != nil {
				return nil, errors.Errorf("bad whitelist value '%s'", listSlice[i])
			}
		}
	} else {
		return nil, errors.Errorf(msgBadEnv, envPrefix, envTelegramWhitelist)
	}

	// can be empty, then proxy is not enabled
	config.proxyAddress = viper.GetString(envProxyAddress)

	// can't be empty if proxy URl isn't
	if config.proxyUser = viper.GetString(envProxyUsername); config.proxyAddress != "" && config.proxyUser == "" {
		return nil, errors.Errorf(msgBadEnv, envPrefix, envProxyUsername)
	}

	// can't be empty if proxy URl isn't
	if config.proxyPass = viper.GetString(envProxyPassword); config.proxyAddress != "" && config.proxyPass == "" {
		return nil, errors.Errorf(msgBadEnv, envPrefix, envProxyPassword)
	}

	// set to "en" if empty
	if config.defaultLang = viper.GetString(envDefaultLang); config.defaultLang == "" {
		config.defaultLang = "en"
		log.Printf(msgBadEnv+" - setting default value to 'en'", envPrefix, envDefaultLang)
	}

	// set to 60 if empty
	if config.updateTimeout = viper.GetInt(envUpdateTimeout); config.updateTimeout == 0 {
		config.updateTimeout = 60
		log.Printf(msgBadEnv+" - setting default value to 60", envPrefix, envUpdateTimeout)
	}

	// set to 2m if empty
	if config.ctxTimeout = viper.GetDuration(envResponseTimeout); config.ctxTimeout == 0 {
		config.ctxTimeout = 2 * time.Minute
		log.Printf(msgBadEnv+" - setting default value to 2m", envPrefix, envResponseTimeout)
	}

	return config, nil
}
