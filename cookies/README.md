# cookies

_________

cookies is a Discord bot that allows creating and exchanging cookies in a community, as a mini-game with a character 
growth system. 

Initially, this app keeps track of cookies for users using simple commands, described below.

_________

### Installation

**Creating your bot**

1. **Create Application**: Go to [Discord Developers' Applications](https://discord.com/developers/applications), log in, and click "New Application," then name it.
2. **Add Bot User**: Navigate to the "Bot" tab in your application, click "Add Bot," and reset the token to get your bot's unique password (keep this secret!).
3. **Enable Intents**: In the same "Bot" section, enable privileged **Server Members Intents** for your bot to function.
4. **Generate installation URL**: Go to the "Installation" > "Default Installation Settings" tab, and under Guild Install:
    1. select "bot" and "applications.commands"
    2. "Send Messages", "Send Messages in Threads"
5. **Install the bot**: In the same "Installation" tab, under "Install Link", copy the Discord Provided Link and paste it into your browser to add the bot to your server.

**Running your bot**

1. Clone this repository in your target host server. You must have Go 1.25.5^ installed.
2. Build the app with the command:

```bash
go build -o out/cookies ./cmd/cookies
```

3. Prepare your configuration:

|      Flag      | Required |     Type      |                                                 Description                                                  |
|:--------------:|:--------:|:-------------:|:------------------------------------------------------------------------------------------------------------:|
|    `-token`    |   yes    |    string     |                The access token for the bot, as retrieved in _Creating your bot_, in step 2.                 |
|    `-chan`     |   yes    |    string     |              The target channel ID where the bot should post notifications about its removals.               |
|   `-db-path`   |    no    |    string     |                           The path to the SQLite database file to use (or create).                           |
|  `-in-memory`  |    no    |     bool      | Use an in-memory implementation instead of one with a database (data is lost each time the bot is restarted) |
|  `-adminlist`  |    no    |    string     |           A comma-separated list of Discord user IDs that will be labeled as Admin within the app.           |
|   `-server`    |   yes    |    string     |                                The Discord server ID where the app is running                                |
|     `-app`     |   yes    |    string     |                            The Discord app ID (for this application's instance).                             |
|    `-role`     |    no    |    string     |                           Discord role ID to allow users to give or share cookies                            |
| `-max-cookies` |    no    |      int      |                       The maximum number of cookies a regular user can give or share.                        |
|   `-thresh`    |    no    | time.Duration |                                 Cooldown between adding or sharing cookies.                                  |

4. Run the app with the `bot` subcommand alongside the appropriate flags. Example:

```bash
./out/cookies bot \
  -token="MTQ1..." \
  -db-path=$(pwd)/data/db.sqlite \
  -server 1234567890123456789 -app 1234567890123456780 \
  -chan 1234567890123456781  -adminlist 123456789012345679 -role 1234567890123456782 \
  -thresh 1m -max-cookies 10
```

5. When starting, the app will print a log message confirming that it has connected to Discord successfully, and any 
events that occur during its runtime

```
{"time":"2025-12-22T00:00:00.323017238Z","level":"INFO","source":{"function":"main.ExecBot","file":"/runtime/cookies/cmd/cookies/main.go","line":95},"msg":"admin-list is set","admin_users":["123456789012345679"]}
{"time":"2025-12-22T00:00:00.450779066Z","level":"INFO","source":{"function":"github.com/zalgonoise/x/cookies/internal/repository/sqlite.OpenSQLite","file":"/runtime/cookies/internal/repository/sqlite/database.go","line":60},"msg":"opened target DB","uri":"/runtime/cookies/data/db.sqlite"}
{"time":"2025-12-22T00:00:00.45147947Z","level":"INFO","source":{"function":"github.com/zalgonoise/x/cookies/internal/repository/sqlite.OpenSQLite","file":"/runtime/cookies/internal/repository/sqlite/database.go","line":66},"msg":"prepared pragmas"}
{"time":"2025-12-22T00:00:00.451621271Z","level":"INFO","source":{"function":"github.com/zalgonoise/x/cookies/internal/repository/sqlite.MigrateSQLite","file":"/runtime/cookies/internal/repository/sqlite/migrations.go","line":25},"msg":"operation completed","time_elapsed":106281}
{"time":"2025-12-22T00:00:00.062845557Z","level":"INFO","source":{"function":"main.ExecBot","file":"/runtime/cookies/cmd/cookies/main.go","line":176},"msg":"connected to discord","server_id":"1234567890123456789","app_id":"1234567890123456780","log_channel_id":"1234567890123456781","admin_list":["123456789012345679"],"role":"1234567890123456782","threshold":"1m0s","non_admin_max_cookies":10}
```
