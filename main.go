package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"google.golang.org/api/option"

	"cloud.google.com/go/translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
	"golang.org/x/text/language"
)

func main() {
	// load config and do preparation
	ctx := context.Background()

	configName := flag.String("config", "config", "Config file name")
	flag.Parse()

	log.Print("Loading config")
	cfg := loadConfig(*configName)

	targetLang, err := language.Parse(cfg.GetString("google_api.default_lang"))
	if err != nil {
		log.Fatalf("Unable to parse the target language: %v", err)
	}

	bot := &botClient{cfg: cfg, targetLang: targetLang}
	log.Print("Starting the bot")

	// Create a Google Translate client
	gtClient, err := translate.NewClient(ctx, option.WithCredentialsFile(cfg.GetString("google_api.cred_path")))
	if err != nil {
		log.Fatalf("Failed to create Google Translate client: %v", err)
	}
	bot.gtClient = gtClient
	defer bot.gtClient.Close()
	log.Print("Connected to the Google Translate API")

	// Setup a custom HTTP client with a SOCKS5 proxy dialer (if enabled)
	httpClient := http.DefaultClient
	if proxyAddress := cfg.GetString("proxy.address"); proxyAddress != "" {
		log.Print("Proxy is enabled")
		// enabling proxy
		dialer, sErr := proxy.SOCKS5("tcp",
			proxyAddress,
			&proxy.Auth{
				User:     cfg.GetString("proxy.user"),
				Password: cfg.GetString("proxy.pass"),
			},
			proxy.Direct,
		)
		if sErr != nil {
			log.Fatalf("Can't connect to the proxy: %v", sErr)
		}

		httpClient = &http.Client{Transport: &http.Transport{Dial: dialer.Dial}}
		log.Print("Proxy dialer initiated")
	} else {
		log.Print("Proxy is disabled")
	}

	// Init the Telegram bot
	tgClient, err := tgbotapi.NewBotAPIWithClient(cfg.GetString("telegram.token"), httpClient)
	if err != nil {
		log.Fatalf("Unable to init Telegram bot: %v", err)
	}
	bot.tgClient = tgClient
	log.Print("Bot ready")

	// Listen to the user messages
	log.Fatalf("Failed to listen to the updates: %v", bot.listenUpdates(ctx))
}
