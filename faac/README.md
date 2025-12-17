# faac

_________


**FPC account age checker** is a simple bot for Discord that automatically kicks users 
from a certain server whose account age is younger than _N_ days.

This serves as a basic floodgate for suspicious accounts joining your server (such as phishing, malware, scamming,
sockpuppet / burner-account users and bots).

_________

### Installation

**Creating your bot**

1. **Create Application**: Go to [Discord Developers' Applications](https://discord.com/developers/applications), log in, and click "New Application," then name it.
2. **Add Bot User**: Navigate to the "Bot" tab in your application, click "Add Bot," and reset the token to get your bot's unique password (keep this secret!).
3. **Enable Intents**: In the same "Bot" section, enable privileged **Server Members Intents** for your bot to function.
4. **Generate installation URL**: Go to the "Installation" > "Default Installation Settings" tab, and under Guild Install:
   1. select "bot" and "applications.commands"
   2. "Kick Members", "Send Messages", "Send Messages in Threads"
5. **Install the bot**: In the same "Installation" tab, under "Install Link", copy the Discord Provided Link and paste it into your browser to add the bot to your server.

**Running your bot**

1. Clone this repository in your target host server. You must have Go 1.25.5^ installed.
2. Build the app with the command:

```bash
go build -o out/faac ./cmd/faac
```

3. Prepare your configuration:

|    Flag     |  Type  |                                           Description                                           |
|:-----------:|:------:|:-----------------------------------------------------------------------------------------------:|
|  `-token`   | string |          The access token for the bot, as retrieved in _Creating your bot_, in step 2.          |
|   `chan`    | string |        The target channel ID where the bot should post notifications about its removals.        |
|   `days`    |  int   | The minimum account age in days to serve as threshold for deletion. Default: `30`. Minimum: `7` |
| `allowlist` | string |                A comma-separated list of user IDs that should bypass this rule.                 |

4. Run the app with the `bot` subcommand alongside the appropriate flags. Example:

```bash
./out/faac bot -days 365 -chan 1234567890123456789 -allowlist 123456789012345670,123456789012345671 -token="MTQ0O..."
```

5. When starting, the app will print a log message confirming that it has connected to Discord successfully, and also 
log any errors that it runs into, or when it acts on a suspicious account when kicking it. Example:

```
{"time":"2025-12-15T00:00:00.323017238Z","level":"INFO","source":{"function":"main.ExecBot","file":"/runtime/faac/cmd/faac/main.go","line":95},"msg":"connected to discord","log_channel_id":"1234567890123456789","min_days_age":365}
{"time":"2025-12-15T00:01:00.459575778Z","level":"INFO","source":{"function":"main.ExecBot.MemberAccountAgeFilter.func1","file":"/runtime/faac/member_join.go","line":42},"msg":"suspicious account detected","account_age":1,"user":"totallyNewToDiscordIn2025","user_id":"123456789012345678"}
```
