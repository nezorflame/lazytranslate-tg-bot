package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

type botClient struct {
	tgClient   *tgbotapi.BotAPI
	gtClient   *translate.Client
	targetLang language.Tag
}

func (b *botClient) listenUpdates(ctx context.Context, config *appConfig) error {
	// Setup the updates channel
	conf := tgbotapi.NewUpdate(0)
	conf.Timeout = config.updateTimeout
	updates, err := b.tgClient.GetUpdatesChan(conf)
	if err != nil {
		return errors.Wrap(err, "unable to get bot updates")
	}
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
		if !intInSlice(u.Message.From.ID, config.tgWhitelist) {
			continue
		}

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "exiting the program")
		default:
			go func(u *tgbotapi.Update) {
				errChan := make(chan error)
				updateCtx, cancel := context.WithTimeout(ctx, config.ctxTimeout)
				defer cancel()
				go b.parseUpdate(updateCtx, errChan, u)

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

func (b *botClient) parseUpdate(ctx context.Context, errChan chan error, u *tgbotapi.Update) {
	// try to detect target message and language
	targetLang, msgText, ok := b.detectTargets(u.Message)
	if !ok {
		// check if mesage was a reply, return if not
		if u.Message.ReplyToMessage == nil {
			errChan <- nil
			return
		}
		// use previously detected language and the original message text
		msgText = u.Message.ReplyToMessage.Text
	}
	log.Printf("[%s] %s", targetLang, msgText)

	// Translate the text into the target language
	lang, translation, err := b.doTranslate(ctx, targetLang, msgText)
	if err != nil {
		errChan <- errors.Wrap(err, "unable to translate the message")
		return
	}
	log.Printf("Translation: %s", translation)

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, fmt.Sprintf("[%s -> %s]\n%s", lang, targetLang, translation))
	msg.ReplyToMessageID = u.Message.MessageID
	if _, err := b.tgClient.Send(msg); err != nil {
		errChan <- errors.Wrap(err, "unable to send the message")
		return
	}

	// All is good
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
			return b.targetLang, strings.Join(msgWords[1:], " "), false
		}

		// bot name, language and message text
		return lang, strings.Join(msgWords[2:], " "), true
	}
}

func (b *botClient) doTranslate(ctx context.Context, targetLang language.Tag, text string) (string, string, error) {
	translations, err := b.gtClient.Translate(ctx, []string{text}, targetLang, nil)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to translate text")
	} else if len(translations) < 1 {
		return "", "", errors.New("no translations")
	}
	return translations[0].Source.String(), translations[0].Text, nil
}

func intInSlice(s int, ss []int) bool {
	for i := range ss {
		if ss[i] == s {
			return true
		}
	}
	return false
}
