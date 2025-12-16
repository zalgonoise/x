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

	"github.com/bwmarrin/discordgo"
	"github.com/zalgonoise/x/cli/v2"

	"github.com/zalgonoise/x/faac"
	"github.com/zalgonoise/x/faac/internal/log"
)

const (
	defaultDaysAge = 30
	minDaysAge     = 7
)

var (
	ErrNoToken        = errors.New("no token provided")
	ErrNoLogChannelID = errors.New("no log channel ID provided")
	ErrDaysAgeTooLow  = errors.New("configured account age threshold is too low")
)

func main() {
	logger := log.New("debug", true, true)

	runner := cli.NewRunner("faac",
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
	daysAge := fs.Int("days", defaultDaysAge, "min days age")
	allowedUsers := fs.String("allowlist", "", "comma-separated list of user IDs that should bypass this rule")

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

	if *daysAge < minDaysAge {
		logger.ErrorContext(ctx, ErrDaysAgeTooLow.Error(), slog.Int("days", *daysAge))

		return 1, ErrDaysAgeTooLow
	}

	var allowList []string

	if *allowedUsers != "" {
		allowList = strings.Split(*allowedUsers, ",")

		logger.InfoContext(ctx, "allowlist is set",
			slog.Any("allowed_users", allowList))
	}

	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		logger.ErrorContext(ctx, "error creating Discord session", slog.String("error", err.Error()))

		return 1, err
	}

	// register the handler for when a new user joins
	dg.AddHandler(faac.MemberAccountAgeFilter(logger, *daysAge, *logChannelID, allowList...))

	// intents: we need GuildMembers to detect joins
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	if err := dg.Open(); err != nil {
		logger.ErrorContext(ctx, "error opening connection", slog.String("error", err.Error()))

		return 1, err
	}

	logger.InfoContext(ctx, "connected to discord",
		slog.String("log_channel_id", *logChannelID),
		slog.Int("min_days_age", *daysAge),
	)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.InfoContext(ctx, "shutting down")

	if err := dg.Close(); err != nil {
		logger.ErrorContext(ctx, "error closing discord connection", slog.String("error", err.Error()))

		return 1, err
	}

	logger.InfoContext(ctx, "disconnected from discord")

	return 0, nil
}
