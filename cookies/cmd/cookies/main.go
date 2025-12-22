package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zalgonoise/x/cli/v2"

	"github.com/zalgonoise/x/cookies"
	"github.com/zalgonoise/x/cookies/internal/log"
	"github.com/zalgonoise/x/cookies/internal/repository/memory"
	"github.com/zalgonoise/x/cookies/internal/repository/sqlite"
)

const minDuration = time.Minute

var (
	ErrNoToken        = errors.New("no token provided")
	ErrNoLogChannelID = errors.New("no log channel ID provided")
	ErrNoDatabasePath = errors.New("no database path provided")
)

type Command interface {
	Name() string
	Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)
	Elements() []cookies.ApplicationCommandOpts
}

func main() {
	logger := log.New("debug", true, true)

	runner := cli.NewRunner("cookies",
		cli.WithExecutors(map[string]cli.Executor{
			"bot": cli.Executable(ExecBot),
		}),
	)

	code, err := runner.Run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func ExecBot(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("bot", flag.ContinueOnError)

	token := fs.String("token", "", "bot token")
	logChannelID := fs.String("chan", "", "log channel id")
	dbPath := fs.String("db-path", "", "SQLite database path")
	inMemory := fs.Bool("in-memory", false, "use in-memory storage")
	adminUsers := fs.String("adminlist", "", "comma-separated list of admin user IDs that should be able to add or remove any number of cookies")
	serverID := fs.String("server", "", "discord server ID")
	appID := fs.String("app", "", "discord app ID")
	role := fs.String("role", "", "discord role ID to allow users to give cookies")
	nonAdminMaxCookies := fs.Int("max-cookies", 1, "maximum number of cookies regular users can give")
	thresh := fs.Duration("thresh", time.Hour*24, "duration between adding and sharing cookies")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *token == "" {
		logger.ErrorContext(ctx, "token is required")

		return 1, ErrNoToken
	}

	if *logChannelID == "" {
		logger.ErrorContext(ctx, "token is required")

		return 1, ErrNoLogChannelID
	}

	if *dbPath == "" && !*inMemory {
		logger.ErrorContext(ctx, "database path is required")

		return 1, ErrNoDatabasePath
	}

	var adminList []string

	if *adminUsers != "" {
		adminList = strings.Split(*adminUsers, ",")

		logger.InfoContext(ctx, "admin-list is set",
			slog.Any("admin_users", adminList))
	}

	if *thresh < minDuration {
		*thresh = minDuration
	}

	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		logger.ErrorContext(ctx, "error creating Discord session", slog.String("error", err.Error()))

		return 1, err
	}

	defer func() {
		if err := dg.Close(); err != nil {
			logger.ErrorContext(ctx, "closing discord connection", slog.String("error", err.Error()))

			return
		}

		logger.InfoContext(ctx, "disconnected from discord")
	}()

	var repo cookies.Repository
	switch {
	case *inMemory:
		repo = memory.NewInMemory(realClock{})
	default:
		db, err := sqlite.OpenSQLite(*dbPath, sqlite.ReadWritePragmas(), logger)
		if err != nil {
			logger.ErrorContext(ctx, "opening SQLite database", slog.String("error", err.Error()))

			return 1, err
		}

		defer func() {
			if err := db.Close(); err != nil {
				logger.ErrorContext(ctx, "closing SQLite database", slog.String("error", err.Error()))
			}
		}()

		if err := sqlite.MigrateSQLite(ctx, db, logger); err != nil {
			logger.ErrorContext(ctx, "migrating SQLite database", slog.String("error", err.Error()))

			return 1, err
		}

		repo = sqlite.NewSQLite(db, realClock{})
	}

	commands := []Command{
		cookies.NewAddCommand(adminList, *logChannelID, *role, *nonAdminMaxCookies, *thresh, repo, realClock{}, logger),
		cookies.NewGetCommand(adminList, *logChannelID, repo, logger),
		cookies.NewListCommand(adminList, *logChannelID, repo, logger),
		cookies.NewSwapCommand(adminList, *logChannelID, *role, *nonAdminMaxCookies, *thresh, repo, realClock{}, logger),
		cookies.NewEatCommand(adminList, *logChannelID, repo, logger),
	}

	cmds := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, command := range commands {
		handler, cmd := cookies.RegisterSlashCommand(logger, command.Name(), command.Callback, command.Elements()...)
		dg.AddHandler(handler)

		cmds = append(cmds, cmd)
	}

	if _, err := dg.ApplicationCommandBulkOverwrite(*appID, *serverID, cmds); err != nil {
		logger.ErrorContext(ctx, "error creating commands",
			slog.String("error", err.Error()))

		return 1, err
	}

	if err := dg.Open(); err != nil {
		logger.ErrorContext(ctx, "error opening connection", slog.String("error", err.Error()))

		return 1, err
	}

	logger.InfoContext(ctx, "connected to discord",
		slog.String("server_id", *serverID),
		slog.String("app_id", *appID),
		slog.String("log_channel_id", *logChannelID),
		slog.Any("admin_list", adminList),
		slog.String("role", *role),
		slog.String("threshold", thresh.String()),
		slog.Int("non_admin_max_cookies", *nonAdminMaxCookies),
	)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.InfoContext(ctx, "shutting down")

	return 0, nil
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
