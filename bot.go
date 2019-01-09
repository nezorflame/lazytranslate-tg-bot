package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"cloud.google.com/go/translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

type botClient struct {
	cfg        *viper.Viper
	tgClient   *tgbotapi.BotAPI
	gtClient   *translate.Client
	targetLang language.Tag
}

func (b *botClient) listenUpdates(ctx context.Context) error {
	// Setup the updates channel
	conf := tgbotapi.NewUpdate(0)
	conf.Timeout = b.cfg.GetInt("telegram.timeout")
	updates := b.tgClient.GetUpdatesChan(conf)
	defer b.tgClient.StopReceivingUpdates()

	// Listen to the messages
	for u := range updates {
		// ignore any non-Message updates
		if u.Message == nil {
			continue
		}

		// TODO: remove later
		log.Printf("%s %d", u.Message.From.UserName, u.Message.From.ID)

		// we need only messages directed to the bot
		if !strings.HasPrefix(u.Message.Text, "@"+b.tgClient.Self.UserName) {
			continue
		}

		// only accept messages of the users from the whitelist
		if !intInStringSlice(u.Message.From.ID, b.cfg.GetStringSlice("whitelist")) {
			continue
		}

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "exiting the program")
		default:
			go func(u *tgbotapi.Update) {
				errChan := make(chan error)
				updateCtx, cancel := context.WithTimeout(ctx, b.cfg.GetDuration("ctx_timeout"))
				defer cancel()
				go b.parseUpdate(updateCtx, errChan, u.Message)

				select {
				case err := <-errChan:
					if err != nil {
						log.Printf("Failed to parse update: %v", err)
					}
				case <-updateCtx.Done():
					log.Printf("Failed to parse update: %v", updateCtx.Err())
				}
			}(&u)
		}
	}
	return nil
}

func (b *botClient) parseUpdate(ctx context.Context, errChan chan error, m *tgbotapi.Message) {
	// try to detect target message and language
	var tag string
	targetLang, msgText, ok := b.detectTargets(m)
	if !ok {
		// check if mesage was a reply, return if not
		if m.ReplyToMessage == nil {
			errChan <- nil
			return
		}

		// using previously detected language and new text
		msgText = m.ReplyToMessage.Text

		// check if the reply was a translation itself
		if m.ReplyToMessage.From.ID == b.tgClient.Self.ID {
			msgParts := strings.SplitN(msgText, "\n", 2)
			if len(msgParts) > 1 {
				tag = fmt.Sprintf(strings.Replace(msgParts[0], "]", " -> \"%s\"]", 1), targetLang)
				msgText = msgParts[1]
			}
		}
	}
	log.Printf("[%s] %s", targetLang, msgText)

	// translate the text into the target language
	lang, msgText, err := b.doTranslate(ctx, targetLang, msgText)
	if err != nil {
		errChan <- errors.Wrap(err, "unable to translate the message")
		return
	}
	log.Printf("Translation: %s", msgText)

	// add tag if present
	if tag != "" {
		msgText = fmt.Sprintf("%s\n%s", tag, msgText)
	} else {
		msgText = fmt.Sprintf("[\"%s\" -> \"%s\"]\n%s", lang, targetLang, msgText)
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, msgText)
	msg.ReplyToMessageID = m.MessageID
	if _, err := b.tgClient.Send(msg); err != nil {
		errChan <- errors.Wrap(err, "unable to send the message")
		return
	}

	// all is good
	errChan <- nil
	return
}

func (b *botClient) detectTargets(msg *tgbotapi.Message) (language.Tag, string, bool) {
	// we expect a message like these (language is optional):
	// `@botname ["lang"] message`
	// `original message -> @botname ["lang"]`
	msgWords := strings.SplitN(msg.Text, " ", 3)
	switch len(msgWords) {
	case 1:
		// only bot name is present - use reply message text
		return b.targetLang, "", false
	case 2:
		// try to parse second word as language
		lang, err := language.Parse(strings.Trim(msgWords[1], `"`))
		if err != nil {
			// bot name and message text
			return b.targetLang, msgWords[1], true
		}

		// only bot name and language are present - use reply message text
		return lang, "", false
	default:
		// try to parse second word as language
		lang, err := language.Parse(strings.Trim(msgWords[1], `"`))
		if err != nil {
			// bot name and message text
			return b.targetLang, strings.Join(msgWords[1:], " "), true
		}

		// bot name, language and message text
		return lang, strings.Join(msgWords[2:], " "), true
	}
}

func (b *botClient) doTranslate(ctx context.Context, targetLang language.Tag, text string) (string, string, error) {
	// translate
	translations, err := b.gtClient.Translate(ctx, []string{text}, targetLang, nil)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to translate text")
	} else if len(translations) < 1 {
		return "", "", errors.New("no translations")
	}

	// result is escaped - unescape
	return translations[0].Source.String(), html.UnescapeString(translations[0].Text), nil
}

func intInStringSlice(i int, ss []string) bool {
	si := strconv.Itoa(i)
	for _, s := range ss {
		if s == si {
			return true
		}
	}
	return false
}
