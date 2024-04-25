# Celebrations

Facilitates celebrations by sending reminders via custimizable channels like:

- Slack direct messages,
- Slack personal reminders,
- Slack channels.

Celebrations works based on birth date and anniversary dates along with Slack identifiers (see [example/config.yml](example/config.yml)).

## Installation

1. Copy `example/config.yml` to your app directory; modify according to your needs.
1. Install app to desired **Slack** workspace.
1. Required **Slack** permissions:

   - bot token scopes:
     - `chat:write`
     - `chat:write.customize`
     - `chat:write.public`
     - `users.profile:read`
     - `users:read`
     - `channels:manage`
     - `groups:write`
     - `im:write`
     - `mpim:write`
   - user token scopes:
     - `reminders:write`

1. Optional.
   Use command `./celebrations download-users [--limit x]` to pre-download users from **Slack**. Can be helpful when populating `config.yml` file with people.
1. Schedule running `./celebrations send-reminders` exacutable once a day on specified hour e.g. 9:30 am via [Github actions scheduler](example/.github/workflows/main.yml).

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

## License

[MIT](LICENSE)
