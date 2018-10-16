# lazytranslate-tg-bot [![Build Status](https://travis-ci.org/nezorflame/lazytranslate-tg-bot.svg?branch=master)](https://travis-ci.org/nezorflame/lazytranslate-tg-bot) [![Go Report Card](https://goreportcard.com/badge/github.com/nezorflame/lazytranslate-tg-bot)](https://goreportcard.com/report/github.com/nezorflame/lazytranslate-tg-bot)

Telegram bot for the lazy translation through the Google Translate API.

## Usage

Either

1. Create `.env` file... OR
2. Export env variables...

...according to the `example.env`:

| Env variable | Type | Mandatory | Default value | Description |
| - | - | - | - | - |
| GOOGLE_APPLICATION_CREDENTIALS | `string` | Yes | - | Google Translate API credentials, see more [here](https://cloud.google.com/translate/docs/) |
| LAZYTRANSLATE_TG_WHITELIST | `string` (comma-separated `[]int`) | Yes | - | List of Telegram user IDs |
| LAZYTRANSLATE_TG_TOKEN | `string` | Yes | - | Your Telegram bot token |
| LAZYTRANSLATE_PROXY_ADDR | `string` | Yes | - | Proxy address URL |
| LAZYTRANSLATE_PROXY_USER | `string` | Yes | - | Proxy username |
| LAZYTRANSLATE_PROXY_PASS | `string` | Yes | - | Proxy password |
| LAZYTRANSLATE_DEFAULT_LANG | `string` | No | 'en' | Default language for translations |
| LAZYTRANSLATE_UPDATE_TIMEOUT | `int` | No | 60 | Timeout for Telegram updates (in seconds) |
| LAZYTRANSLATE_RESPONSE_TIMEOUT | `time.Duration` | No | '1m' | Context timeout (for bot operations) |

## Installation

Bot requires Go 1.11+ as it's using Go modules as a dependency management tool:

```bash
export GO111MODULE=on
git clone https://github.com/nezorflame/lazytranslate-tg-bot.git
cd lazytranslate-tg-bot
# create env variables
# or edit .env file:
# cp example.env .env && nano .env
go build
./lazytranslate-tg-bot
```
