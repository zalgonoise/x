package cookies

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/zalgonoise/x/cookies/internal/repository"
)

const (
	commandAdd  = "dealcookies"
	commandList = "cookies"
	commandGet  = "getcookies"
	commandSwap = "givecookies"
	commandEat  = "eatcookie"

	userLabel  = "User"
	adminLabel = "Cookie Factory"

	minNonAdminMaxCookies = 1
)

var (
	ErrCreateInteractionIsNil = errors.New("create interaction is nil")
	ErrInteractionUserIsNil   = errors.New("interaction user is nil")
	ErrMalformedInteraction   = errors.New("malformed interaction")
)

type ApplicationCommandOpts func(command *discordgo.ApplicationCommand)

type CommandCallback func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)

type Repository interface {
	GetCookies(ctx context.Context, user string) (int, time.Time, error)
	ListCookies(ctx context.Context) (map[string]int, error)
	AddCookie(ctx context.Context, user string, n int) (int, error)
	SwapCookies(ctx context.Context, from, to string, n int) (int, int, error)
	EatCookie(ctx context.Context, user string) (int, error)
}

type Clock interface {
	Now() time.Time
}

type AddCommand struct {
	adminList          []string
	logChannelID       string
	giverRole          string
	nonAdminMaxCookies int
	giftThreshold      time.Duration

	repo   Repository
	clock  Clock
	logger *slog.Logger
}

func (c *AddCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give cookies
	requester, canGiveCookies, err := getUser(i.Interaction, c.giverRole)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set cookies admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	// if user cannot give cookies and is not admin, exit w/ message
	if !canGiveCookies && !isAdmin {
		c.logger.WarnContext(ctx, "user cannot deal cookies")

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNoPermissionsToSendCookies(requester.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "not allowed to deal cookies", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	// get target user and number of cookies from command
	cookies, user, err := getCookiesAndUserID(s, i.ApplicationCommandData().Options)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting receiver's context", slog.String("error", err.Error()))

		return nil, err
	}

	// admins bypass these checks
	if !isAdmin {
		c.logger.DebugContext(ctx, "checking last gift's timestamp")

		// get the timestamp for the last gift
		_, lastGift, err := c.repo.GetCookies(ctx, requester.ID)
		if errors.Is(err, repository.ErrNotFound) {
			c.logger.ErrorContext(ctx, "requester was not found", slog.String("error", err.Error()))

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNotEnoughCookies(requester.ID, 0, cookies)); err != nil {
				c.logger.ErrorContext(ctx, "sending message",
					slog.String("action", c.Name()),
					slog.String("log_channel_id", c.logChannelID),
					slog.String("error", err.Error()))
			}

			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "not enough cookies", Flags: discordgo.MessageFlagsEphemeral},
			}, nil
		}
		if err != nil {
			c.logger.ErrorContext(ctx, "error fetching requester's data", slog.String("error", err.Error()))

			return nil, err
		}

		// calc how long it's been since they last gifted cookies
		durSinceLastGift := c.clock.Now().Sub(lastGift)

		// check if they are spamming too often
		if durSinceLastGift < c.giftThreshold {
			c.logger.ErrorContext(ctx, "requester is spamming cookies too often")

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedSpammingCookies(requester.ID, c.giftThreshold-durSinceLastGift)); err != nil {
				c.logger.ErrorContext(ctx, "sending message",
					slog.String("action", c.Name()),
					slog.String("log_channel_id", c.logChannelID),
					slog.String("error", err.Error()))
			}

			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "you cookie spammer", Flags: discordgo.MessageFlagsEphemeral},
			}, nil
		}

		// if the user is not admin, they will be capped to gifting n cookies
		if cookies > c.nonAdminMaxCookies {
			c.logger.WarnContext(ctx, "limited requester's added cookies to maximum allowed")

			cookies = c.nonAdminMaxCookies
		}
	}

	c.logger.DebugContext(ctx, "adding cookies")

	// add cookies to target user
	n, err := c.repo.AddCookie(ctx, user.ID, cookies)
	if errors.Is(err, repository.ErrNotFound) {
		c.logger.ErrorContext(ctx, "requester does not exist", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNotEnoughCookies(requester.ID, 0, cookies)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "not enough cookies", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to add cookies", slog.String("error", err.Error()))

		return nil, err
	}

	// reply
	label := userLabel
	if isAdmin {
		label = adminLabel
	}

	c.logger.DebugContext(ctx, "added cookies")

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedAddedCookies(label, requester.ID, user.ID, n)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "added cookies!", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *AddCommand) Name() string {
	return commandAdd
}

func (c *AddCommand) Elements() []ApplicationCommandOpts {
	return []ApplicationCommandOpts{
		CommandWithElement("user", "user to assign cookies to", discordgo.ApplicationCommandOptionUser, true),
		CommandWithElement("cookies", "number of cookies to assign", discordgo.ApplicationCommandOptionInteger, true)}
}

type GetCommand struct {
	adminList    []string
	logChannelID string

	repo   Repository
	logger *slog.Logger
}

func (c *GetCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give cookies
	requester, _, err := getUser(i.Interaction, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set cookies admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	user, err := getUserID(s, i.ApplicationCommandData().Options)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting receiver's context", slog.String("error", err.Error()))

		return nil, err
	}

	label := userLabel
	if isAdmin {
		label = adminLabel
	}

	c.logger.DebugContext(ctx, "getting cookies for user")

	n, lastGift, err := c.repo.GetCookies(ctx, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		c.logger.ErrorContext(ctx, "user does not exist", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID,
			EmbedTotalCookies(label, user.ID, 0, time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
		); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "fetched cookies", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to fetch cookies", slog.String("error", err.Error()))

		return nil, err
	}

	c.logger.DebugContext(ctx, "fetched cookies for user")

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedTotalCookies(label, user.ID, n, lastGift)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "fetched cookies", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *GetCommand) Name() string {
	return commandGet
}

func (c *GetCommand) Elements() []ApplicationCommandOpts {
	return []ApplicationCommandOpts{
		CommandWithElement("user", "user to fetch cookies for", discordgo.ApplicationCommandOptionUser, true)}
}

type ListCommand struct {
	adminList    []string
	logChannelID string

	repo   Repository
	logger *slog.Logger
}

func (c *ListCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give cookies
	requester, _, err := getUser(i.Interaction, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set cookies admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	c.logger.DebugContext(ctx, "getting all cookies")

	cookieMap, err := c.repo.ListCookies(ctx)

	if errors.Is(err, repository.ErrNotFound) || len(cookieMap) == 0 {
		c.logger.ErrorContext(ctx, "no users or cookies added yet", slog.String("error", err.Error()))

		// no users found yet
		label := userLabel
		if isAdmin {
			label = adminLabel
		}

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedListCookiesForNoUsers(label)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "no cookies for any users", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to list all cookies", slog.String("error", err.Error()))

		return nil, err
	}

	embeds := make([]*discordgo.MessageEmbed, 0, len(cookieMap))
	for user, cookies := range cookieMap {
		label := strings.ToLower(userLabel)
		if slices.Contains(c.adminList, user) {
			label = adminLabel
		}

		embeds = append(embeds, EmbedListCookiesForUser(label, user, cookies))
	}

	c.logger.DebugContext(ctx, "fetched all cookies")

	if _, err := s.ChannelMessageSendEmbeds(c.logChannelID, embeds); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "listed users with cookies", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *ListCommand) Name() string {
	return commandList
}

func (c *ListCommand) Elements() []ApplicationCommandOpts {
	return nil
}

type SwapCommand struct {
	adminList          []string
	logChannelID       string
	giverRole          string
	nonAdminMaxCookies int
	giftThreshold      time.Duration

	repo   Repository
	clock  Clock
	logger *slog.Logger
}

func (c *SwapCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give cookies
	requester, canGiveCookies, err := getUser(i.Interaction, c.giverRole)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set cookies admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	// if user cannot give cookies and is not admin, exit w/ message
	if !canGiveCookies && !isAdmin {
		c.logger.WarnContext(ctx, "user cannot share cookies")

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNoPermissionsToSendCookies(requester.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "not allowed to send cookies", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	// get target user and number of cookies from command
	cookies, user, err := getCookiesAndUserID(s, i.ApplicationCommandData().Options)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting receiver's context", slog.String("error", err.Error()))

		return nil, err
	}

	// caller tries to input negative cookies (trying to steal)
	if cookies <= 0 {
		c.logger.WarnContext(ctx, "requester tried to steal cookies from receiver")

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedCookieStealer(requester.ID, user.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "STOP! cookie stealer!~", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	// admins bypass these checks
	if !isAdmin {
		c.logger.DebugContext(ctx, "checking last gift's timestamp")

		// get the timestamp for the last gift
		current, lastGift, err := c.repo.GetCookies(ctx, requester.ID)
		if errors.Is(err, repository.ErrNotFound) {
			c.logger.ErrorContext(ctx, "requester was not found", slog.String("error", err.Error()))

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNotEnoughCookies(requester.ID, current, cookies)); err != nil {
				c.logger.ErrorContext(ctx, "sending message",
					slog.String("action", c.Name()),
					slog.String("log_channel_id", c.logChannelID),
					slog.String("error", err.Error()))
			}

			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "not enough cookies", Flags: discordgo.MessageFlagsEphemeral},
			}, nil
		}

		if err != nil {
			c.logger.ErrorContext(ctx, "error fetching requester's data", slog.String("error", err.Error()))

			return nil, err
		}

		if current < cookies {
			c.logger.ErrorContext(ctx, "requester doesn't have enough cookies")

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNotEnoughCookies(requester.ID, current, cookies)); err != nil {
				c.logger.ErrorContext(ctx, "sending message",
					slog.String("action", c.Name()),
					slog.String("log_channel_id", c.logChannelID),
					slog.String("error", err.Error()))
			}

			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "you're trying to send more than what you have", Flags: discordgo.MessageFlagsEphemeral},
			}, nil
		}

		// calc how long it's been since they last gifted cookies
		durSinceLastGift := c.clock.Now().Sub(lastGift)

		// check if they are spamming too often
		if durSinceLastGift < c.giftThreshold {
			c.logger.ErrorContext(ctx, "requester is spamming cookies too often")

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedSpammingCookies(requester.ID, c.giftThreshold-durSinceLastGift)); err != nil {
				c.logger.ErrorContext(ctx, "sending message",
					slog.String("action", c.Name()),
					slog.String("log_channel_id", c.logChannelID),
					slog.String("error", err.Error()))
			}

			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "stop spamming cookies", Flags: discordgo.MessageFlagsEphemeral},
			}, nil
		}

		// if the user is not admin, they will be capped to gifting n cookies
		if cookies > c.nonAdminMaxCookies {
			c.logger.WarnContext(ctx, "limited requester's shared cookies to maximum allowed")

			cookies = c.nonAdminMaxCookies
		}
	}

	c.logger.DebugContext(ctx, "sharing cookies")

	// subtract cookies from requester
	requesterCurrent, targetCurrent, err := c.repo.SwapCookies(ctx, requester.ID, user.ID, cookies)
	if errors.Is(err, repository.ErrNotFound) {
		c.logger.ErrorContext(ctx, "requester does not exist", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNotEnoughCookies(requester.ID, requesterCurrent, cookies)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "not enough cookies", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to swap cookies", slog.String("error", err.Error()))

		return nil, err
	}

	// reply
	label := userLabel
	if isAdmin {
		label = adminLabel
	}

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID,
		EmbedSharedCookies(label, requester.ID, user.ID, cookies, requesterCurrent, targetCurrent)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	c.logger.DebugContext(ctx, "swapped cookies")

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "swapped cookies", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *SwapCommand) Name() string {
	return commandSwap
}

func (c *SwapCommand) Elements() []ApplicationCommandOpts {
	return []ApplicationCommandOpts{
		CommandWithElement("user", "user to share cookies with", discordgo.ApplicationCommandOptionUser, true),
		CommandWithElement("cookies", "number of cookies to share", discordgo.ApplicationCommandOptionInteger, true)}
}

type EatCommand struct {
	adminList    []string
	logChannelID string

	repo   Repository
	logger *slog.Logger
}

func (c *EatCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give cookies
	requester, _, err := getUser(i.Interaction, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set cookies admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	current, _, err := c.repo.GetCookies(ctx, requester.ID)
	if errors.Is(err, repository.ErrNotFound) {
		c.logger.ErrorContext(ctx, "requester does not exist", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedEatingWithoutAnyCookies(requester.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "eating crumbs?", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to fetch requester info", slog.String("error", err.Error()))

		return nil, err
	}

	if current <= 0 {
		c.logger.WarnContext(ctx, "requester doesn't have enough cookies to eat")

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID,
			EmbedEatingWithoutAnyCookies(requester.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "eating crumbs?", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	c.logger.DebugContext(ctx, "eating a cookie")

	// eat a cookie
	n, err := c.repo.EatCookie(ctx, requester.ID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to eat a cookie", slog.String("error", err.Error()))

		return nil, err
	}

	// reply
	label := userLabel
	if isAdmin {
		label = adminLabel
	}

	c.logger.DebugContext(ctx, "ate a cookie")

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID,
		EmbedEatingCookie(label, requester.ID, n)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "ate a cookie!", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *EatCommand) Name() string {
	return commandEat
}

func (c *EatCommand) Elements() []ApplicationCommandOpts {
	return nil
}

func NewAddCommand(
	adminList []string, logChannelID string, giverRole string,
	nonAdminMaxCookies int, thresh time.Duration,
	repo Repository, clock Clock, logger *slog.Logger,
) *AddCommand {
	if nonAdminMaxCookies < minNonAdminMaxCookies {
		nonAdminMaxCookies = minNonAdminMaxCookies
	}

	return &AddCommand{
		adminList:          adminList,
		logChannelID:       logChannelID,
		giverRole:          giverRole,
		nonAdminMaxCookies: nonAdminMaxCookies,
		giftThreshold:      thresh,
		repo:               repo,
		clock:              clock,
		logger:             logger,
	}
}

func NewGetCommand(adminList []string, logChannelID string, repo Repository, logger *slog.Logger) *GetCommand {
	return &GetCommand{adminList: adminList, logChannelID: logChannelID, repo: repo, logger: logger}
}

func NewListCommand(adminList []string, logChannelID string, repo Repository, logger *slog.Logger) *ListCommand {
	return &ListCommand{adminList: adminList, logChannelID: logChannelID, repo: repo, logger: logger}
}

func NewSwapCommand(
	adminList []string, logChannelID string, giverRole string,
	nonAdminMaxCookies int, thresh time.Duration,
	repo Repository, clock Clock, logger *slog.Logger,
) *SwapCommand {
	if nonAdminMaxCookies < minNonAdminMaxCookies {
		nonAdminMaxCookies = minNonAdminMaxCookies
	}

	return &SwapCommand{
		adminList:          adminList,
		logChannelID:       logChannelID,
		giverRole:          giverRole,
		nonAdminMaxCookies: nonAdminMaxCookies,
		giftThreshold:      thresh,
		repo:               repo,
		clock:              clock,
		logger:             logger,
	}
}

func NewEatCommand(adminList []string, logChannelID string, repo Repository, logger *slog.Logger) *EatCommand {
	return &EatCommand{adminList: adminList, logChannelID: logChannelID, repo: repo, logger: logger}
}

func CommandWithElement(name, desc string, typ discordgo.ApplicationCommandOptionType, req bool) ApplicationCommandOpts {
	return func(command *discordgo.ApplicationCommand) {
		command.Options = append(command.Options, &discordgo.ApplicationCommandOption{
			Name:        name,
			Type:        typ,
			Required:    req,
			Description: desc,
		})
	}
}

func RegisterSlashCommand(logger *slog.Logger, command string, callback CommandCallback, opts ...ApplicationCommandOpts) (func(s *discordgo.Session, i *discordgo.InteractionCreate), *discordgo.ApplicationCommand) {
	cmd := &discordgo.ApplicationCommand{
		Name:        command,
		Description: command,
		Options:     []*discordgo.ApplicationCommandOption{},
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name == command {
			ctx := context.Background()

			res, err := callback(ctx, s, i)
			if err != nil {
				logger.ErrorContext(ctx, "failed to execute command",
					slog.String("command", command),
					slog.String("error", err.Error()),
				)

				return
			}

			if err := s.InteractionRespond(i.Interaction, res); err != nil {
				logger.ErrorContext(ctx, "failed to respond to command",
					slog.String("command", command),
					slog.String("error", err.Error()),
					slog.Any("response", res),
				)
			}
		}
	}, cmd
}

func getUser(i *discordgo.Interaction, giverRole string) (*discordgo.User, bool, error) {
	if i == nil {
		return nil, false, ErrCreateInteractionIsNil
	}

	switch {
	case i.User != nil:
		return i.User, false, nil
	case i.Member != nil && i.Member.User != nil:
		canGiveCookies := false
		if giverRole != "" {
			canGiveCookies = slices.Contains(i.Member.Roles, giverRole)
		}

		return i.Member.User, canGiveCookies, nil
	default:
		return nil, false, ErrInteractionUserIsNil
	}
}

func getCookiesAndUserID(s *discordgo.Session, values []*discordgo.ApplicationCommandInteractionDataOption) (int, *discordgo.User, error) {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(values))
	for _, opt := range values {
		optionMap[opt.Name] = opt
	}

	cookies, ok := optionMap["cookies"]
	if !ok {
		return 0, nil, ErrMalformedInteraction
	}

	user, ok := optionMap["user"]
	if !ok {
		return 0, nil, ErrMalformedInteraction
	}

	return int(cookies.IntValue()), user.UserValue(s), nil
}

func getUserID(s *discordgo.Session, values []*discordgo.ApplicationCommandInteractionDataOption) (*discordgo.User, error) {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(values))
	for _, opt := range values {
		optionMap[opt.Name] = opt
	}

	user, ok := optionMap["user"]
	if !ok {
		return nil, ErrMalformedInteraction
	}

	return user.UserValue(s), nil
}
