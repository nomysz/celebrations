[![Version](https://img.shields.io/badge/version-v0.5.0-green.svg)](https://github.com/nomysz/celebrations/releases)
[![GoDoc](https://godoc.org/github.com/nomysz/celebrations?status.svg)](http://godoc.org/github.com/nomysz/celebrations)
[![Go Report Card](https://goreportcard.com/badge/github.com/nomysz/celebrations)](https://goreportcard.com/report/github.com/nomysz/celebrations)

# Celebrations

Facilitates celebrations by sending reminders via custimizable channels like:

- Slack channels,
- Slack direct messages,
- Slack personal reminders.

Celebrations works based on birth date and anniversary dates along with Slack identifiers (see [example/config.yml](example/config.yml)).

## How it works?

* When executed with `./celebrations send-reminders` command *every 1st day of the month* **Monthly report** will be posted to configured leads channel:
<img src="./example/screenshots/monthly-report.png" alt="Monthly report" style="width: 50% !important; box-shadow: 0px 0px 4px 0px #eee;">

* On birthday additional post may be configured to be sent to leads channel:
<img src="./example/screenshots/channel-reminder.png" alt="Channel reminder" style="width: 50% !important; box-shadow: 0px 0px 4px 0px #eee;">

* For direct leads **Direct message** will be sent couple days earlier:
<img src="./example/screenshots/dm-pre-reminder.png" alt="DM reminder" style="width: 50% !important; box-shadow: 0px 0px 4px 0px #eee;">

* Direct leads can also receive extra personal **Reminder** set to given hour:
<img src="./example/screenshots/peronal-reminder.png" alt="Personal reminder" style="width: 50% !important; box-shadow: 0px 0px 4px 0px #eee;">

* When enabled anniversaries, bot will publish anniversary celebrations as well:
<img src="./example/screenshots/anniversary.png" alt="Anniversary" style="width: 50% !important; box-shadow: 0px 0px 4px 0px #eee;">


## Installation

1. Use `bin/celebrations-...` executable or complile (see [Development](#development)) current version to your system architecture.
2. Copy `example/config.yml` to your app directory; modify according to your needs.
3. Install app to desired **Slack** workspace.
4. Required **Slack** permissions:

   - bot token scopes:

     - `chat:write` (posting to channels)
     - `chat:write.public`

     - `channels:manage` (sending DMs)
     - `groups:write`
     - `im:write`
     - `mpim:write`

     - `chat:write.customize` (customizing app visibility)

     - `users:read` (downloading users)
     - `users.profile:read`

   - user token scopes:
     - `reminders:write` (adding reminders)

5. To be able to post to private channel, add bot manually (**Channel** -> **Integrations** -> **Add App**).
6. Optional. Use command `./celebrations download-users [--limit x]` to pre-download users from **Slack**. Helpful for populating `config.yml` file.
7. Setup envronment variables for app runtime:
  - `SLACK_BOT_TOKEN=xoxb-...` (required for most reminders)
  - `SLACK_USER_TOKEN=xoxp-...` (required for setting personal remidners)
8. Schedule running `./celebrations send-reminders` once a day on specified hour e.g. 9:30 am via [Github actions scheduler](example/.github/workflows/main.yml) or other type of cron.

## Development

### Run

```bash
make run # or directly via binary
```

### Build

```bash
make all
```

### Test

```bash
make test
```

## Changelog

### 0.5.0

- Refactor code for readability

### 0.4.0

- Add dates sorting to **Monthly Report** channel reminder
- Move secrets from `config.yml` to ENV VARs

### 0.3.0

- Humanize **Monthly Report** channel reminder

### 0.2.0

- Add **Monthly Report** channel reminder

### 0.1.0

- Init

## License

[MIT](LICENSE)
