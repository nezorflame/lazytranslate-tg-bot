package main

import (
	"log"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	msgEmptyValue      = "empty config value '%s'"
	defaultCtxTimeout  = 30 * time.Second
	defaultTGTimeout   = 60
	defaultRespTimeout = 2 * time.Minute
	defaultLang        = "en"
)

var mandatoryParams = []string{
	"whitelist",
	"telegram.token",
	"google_api.cred_path",
}

func loadConfig(name string) *viper.Viper {
	if name == "" {
		log.Fatal("empty config name")
	}

	cfg := viper.New()

	cfg.SetConfigName(name)
	cfg.SetConfigType("toml")
	cfg.AddConfigPath("$HOME/.config")
	cfg.AddConfigPath("/etc")
	cfg.AddConfigPath(".")

	if err := cfg.ReadInConfig(); err != nil {
		log.Panicf("Unable to read config file: %v", err)
	}
	cfg.WatchConfig()

	cfg.SetDefault("ctx_timeout", defaultCtxTimeout)
	cfg.SetDefault("telegram.timeout", time.Duration(defaultTGTimeout)*time.Second)
	cfg.SetDefault("google_api.timeout", defaultRespTimeout)
	cfg.SetDefault("google_api.default_lang", defaultLang)

	if err := validateConfig(cfg); err != nil {
		log.Panicf("Unable to validate config: %v", err)
	}

	return cfg
}

func validateConfig(cfg *viper.Viper) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	for _, p := range mandatoryParams {
		if cfg.Get(p) == nil {
			return errors.Errorf(msgEmptyValue, p)
		}
	}

	whitelist := cfg.GetStringSlice("whitelist")
	if len(whitelist) < 1 {
		return errors.Errorf("'whitelist' should contain at least one record")
	}

	for _, id := range whitelist {
		if _, err := strconv.Atoi(id); err != nil {
			return errors.Errorf("bad whitelist id '%s'", id)
		}
	}

	return nil
}
