# Celebrations

Facilitates celebrations by sending reminders via custimizable channels like:

- Slack channels,
- Slack direct messages,
- Slack personal reminders.

Celebrations works based on birth date and anniversary dates along with Slack identifiers (see [example/config.yml](example/config.yml)).

## Installation

1. Use `bin/celebrations-...` executable or complile current version to your system architecture.
1. Copy `example/config.yml` to your app directory; modify according to your needs.
1. Install app to desired **Slack** workspace.
1. Required **Slack** permissions:

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

1. To be able to post to private channel, add bot manually (**Channel** -> **Integrations** -> **Add App**).
1. Optional. Use command `./celebrations download-users [--limit x]` to pre-download users from **Slack**. Helpful for populating `config.yml` file.
1. Setup envronment variables for app runtime:
- `SLACK_BOT_TOKEN=xoxb-...` (required for most reminders)
- `SLACK_USER_TOKEN=xoxp-...` (required for setting personal remidners)
1. Schedule running `./celebrations send-reminders` once a day on specified hour e.g. 9:30 am via [Github actions scheduler](example/.github/workflows/main.yml) or other type of cron.

## Development

### Run

```bash
make run
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

[Changelog](Changelog.md)

## License

[MIT](LICENSE)
