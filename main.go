package main

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
	"golang.org/x/text/language"
)

func main() {
	// load config and do preparation
	ctx := context.Background()

	log.Print("Loading config from environment")
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Unable to load config from environment: %v", err)
	}

	targetLang, err := language.Parse(config.defaultLang)
	if err != nil {
		log.Fatalf("Unable to parse the target language: %v", err)
	}

	bot := &botClient{targetLang: targetLang}
	log.Print("Starting the bot")

	// Create a Google Translate client
	gtClient, err := translate.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Google Translate client: %v", err)
	}
	bot.gtClient = gtClient
	defer bot.gtClient.Close()
	log.Print("Connected to the Google Translate API")

	// Setup a custom HTTP client with a SOCKS5 proxy dialer (if enabled)
	httpClient := http.DefaultClient
	if config.proxyAddress != "" {
		log.Print("Proxy is enabled")
		// enabling proxy
		dialer, err := proxy.SOCKS5("tcp", config.proxyAddress, &proxy.Auth{
			User: config.proxyUser, Password: config.proxyPass}, proxy.Direct,
		)
		if err != nil {
			log.Fatalf("Can't connect to the proxy: %v", err)
		}

		httpClient = &http.Client{Transport: &http.Transport{Dial: dialer.Dial}}
		log.Print("Proxy dialer initiated")
	} else {
		log.Print("Proxy is disabled")
	}

	// Init the Telegram bot
	tgClient, err := tgbotapi.NewBotAPIWithClient(config.tgToken, httpClient)
	if err != nil {
		log.Fatalf("Unable to init Telegram bot: %v", err)
	}
	bot.tgClient = tgClient
	log.Print("Bot ready")

	// Listen to the user messages
	log.Fatalf("Failed to listen to the updates: %v", bot.listenUpdates(ctx, config))
}
