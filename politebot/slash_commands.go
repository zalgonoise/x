package politebot

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/zalgonoise/x/politebot/internal/repository"
)

const (
	commandAdd  = "isangy"
	commandList = "angylist"
	commandGet  = "angy"

	userLabel  = "Polite"
	adminLabel = "Necrosympathy"

	minNonAdminMaxPoints = 1
)

var (
	ErrCreateInteractionIsNil = errors.New("create interaction is nil")
	ErrInteractionUserIsNil   = errors.New("interaction user is nil")
	ErrMalformedInteraction   = errors.New("malformed interaction")
)

type ApplicationCommandOpts func(command *discordgo.ApplicationCommand)

type CommandCallback func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)

type Repository interface {
	GetAngyPoints(ctx context.Context, user string) (int, time.Time, error)
	ListAngyPoints(ctx context.Context) (map[string]int, error)
	AddAngyPoints(ctx context.Context, user string, n int) (int, error)
	RegisterUser(ctx context.Context, user string) error
}

type Clock interface {
	Now() time.Time
}

type AddCommand struct {
	adminList           []string
	logChannelID        string
	giverRole           string
	nonAdminMaxPoints   int
	punishmentThreshold time.Duration

	repo   Repository
	clock  Clock
	logger *slog.Logger
}

func (c *AddCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give angy points
	requester, canGiveAngyPoints, err := getUser(i.Interaction, c.giverRole)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set angy points admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	// if user cannot give angy points and is not admin, exit w/ message
	if !canGiveAngyPoints && !isAdmin {
		c.logger.WarnContext(ctx, "user cannot give angy points")

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNoPermissionsToAddAngyPoints(requester.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "not allowed to give angy points", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	// get target user and number of angy points from command
	angyPoints, user, err := getAngyPointsAndUserID(s, i.ApplicationCommandData().Options)
	if err != nil {
		c.logger.ErrorContext(ctx, "getting receiver's context", slog.String("error", err.Error()))

		return nil, err
	}

	// admins bypass these checks
	if !isAdmin {
		c.logger.DebugContext(ctx, "checking last punishment's timestamp")

		// get the timestamp for the last punishment
		_, lastPunishment, err := c.repo.GetAngyPoints(ctx, requester.ID)
		if errors.Is(err, repository.ErrNotFound) {
			c.logger.ErrorContext(ctx, "requester was not found", slog.String("error", err.Error()))

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedNoPermissionsToAddAngyPoints(requester.ID)); err != nil {
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

		// calc how long it's been since they last gave angy points
		durSinceLastPunishment := c.clock.Now().Sub(lastPunishment)

		// check if they are spamming too often
		if durSinceLastPunishment < c.punishmentThreshold {
			c.logger.ErrorContext(ctx, "requester is spamming angy points too often")

			if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedSpammingAngyPoints(requester.ID, c.punishmentThreshold-durSinceLastPunishment)); err != nil {
				c.logger.ErrorContext(ctx, "sending message",
					slog.String("action", c.Name()),
					slog.String("log_channel_id", c.logChannelID),
					slog.String("error", err.Error()))
			}

			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "spamming makes *you* angy", Flags: discordgo.MessageFlagsEphemeral},
			}, nil
		}

		// if the user is not admin, they will be capped to gifting n cookies
		if angyPoints > c.nonAdminMaxPoints {
			c.logger.WarnContext(ctx, "limited requester's added angy points to maximum allowed")

			angyPoints = c.nonAdminMaxPoints
		}
	}

	c.logger.DebugContext(ctx, "adding angy points")

	// add angy points to target user
	n, err := c.repo.AddAngyPoints(ctx, user.ID, angyPoints)
	if errors.Is(err, repository.ErrNotFound) {
		c.logger.ErrorContext(ctx, "requester does not exist", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedAngyPointsForUnknownUser(requester.ID, user.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "not enough angy points", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to add angy points", slog.String("error", err.Error()))

		return nil, err
	}

	// reply
	label := userLabel
	if isAdmin {
		label = adminLabel
	}

	c.logger.DebugContext(ctx, "added angy points")

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedAddedAngyPoints(label, requester.ID, user.ID, n)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "added angy points!", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *AddCommand) Name() string {
	return commandAdd
}

func (c *AddCommand) Elements() []ApplicationCommandOpts {
	return []ApplicationCommandOpts{
		CommandWithElement("user", "user to assign angy points to", discordgo.ApplicationCommandOptionUser, true),
		CommandWithElement("angy_points", "number of angy points to assign", discordgo.ApplicationCommandOptionInteger, true)}
}

type GetCommand struct {
	adminList    []string
	logChannelID string

	repo   Repository
	logger *slog.Logger
}

func (c *GetCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give angy points
	requester, _, err := getUser(i.Interaction, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set angy points admin
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

	c.logger.DebugContext(ctx, "getting angy points for user")

	n, lastPunishment, err := c.repo.GetAngyPoints(ctx, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		c.logger.ErrorContext(ctx, "user does not exist", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID,
			EmbedTotalAngyPoints(label, user.ID, 0, time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
		); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "fetched angy points", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to fetch angy points", slog.String("error", err.Error()))

		return nil, err
	}

	c.logger.DebugContext(ctx, "fetched angy points for user")

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedTotalAngyPoints(label, user.ID, n, lastPunishment)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "fetched angy points", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *GetCommand) Name() string {
	return commandGet
}

func (c *GetCommand) Elements() []ApplicationCommandOpts {
	return []ApplicationCommandOpts{
		CommandWithElement("user", "user to fetch angy points for", discordgo.ApplicationCommandOptionUser, true)}
}

type ListCommand struct {
	adminList    []string
	logChannelID string

	repo   Repository
	logger *slog.Logger
}

func (c *ListCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give angy points
	requester, _, err := getUser(i.Interaction, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set angy points admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	c.logger.DebugContext(ctx, "getting all angy points")

	cookieMap, err := c.repo.ListAngyPoints(ctx)

	if errors.Is(err, repository.ErrNotFound) || len(cookieMap) == 0 {
		c.logger.ErrorContext(ctx, "no users or angy points added yet", slog.String("error", err.Error()))

		// no users found yet
		label := userLabel
		if isAdmin {
			label = adminLabel
		}

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedListAngyPointsForNoUsers(label)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "no angy points for any users", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "failed to list all angy points", slog.String("error", err.Error()))

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

	c.logger.DebugContext(ctx, "fetched all angy points")

	if _, err := s.ChannelMessageSendEmbeds(c.logChannelID, embeds); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "listed users with angy points", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *ListCommand) Name() string {
	return commandList
}

func (c *ListCommand) Elements() []ApplicationCommandOpts {
	return nil
}

type RegisterCommand struct {
	adminList    []string
	logChannelID string

	repo   Repository
	logger *slog.Logger
}

func (c *RegisterCommand) Callback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	// get requester's user and if they have a role enabling them to give angy points
	requester, _, err := getUser(i.Interaction, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "getting requester context", slog.String("error", err.Error()))

		return nil, err
	}

	// check if this user ID is a set angy points admin
	isAdmin := slices.Contains(c.adminList, requester.ID)
	c.logger.DebugContext(ctx, "admin check", slog.Bool("is_admin", isAdmin))

	c.logger.DebugContext(ctx, "registering user", slog.String("user_id", requester.ID))

	if err := c.repo.RegisterUser(ctx, requester.ID); err != nil {
		c.logger.ErrorContext(ctx, "failed to register user", slog.String("error", err.Error()))

		if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedFailedToRegisterUser(requester.ID)); err != nil {
			c.logger.ErrorContext(ctx, "sending message",
				slog.String("action", c.Name()),
				slog.String("log_channel_id", c.logChannelID),
				slog.String("error", err.Error()))
		}

		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "failed to register user", Flags: discordgo.MessageFlagsEphemeral},
		}, nil
	}

	if _, err := s.ChannelMessageSendEmbed(c.logChannelID, EmbedRegisteredUser(requester.ID)); err != nil {
		c.logger.ErrorContext(ctx, "sending message",
			slog.String("action", c.Name()),
			slog.String("log_channel_id", c.logChannelID),
			slog.String("error", err.Error()))
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "registered user", Flags: discordgo.MessageFlagsEphemeral},
	}, nil
}

func (c *RegisterCommand) Name() string {
	return commandList
}
func (c *RegisterCommand) Elements() []ApplicationCommandOpts {
	return nil
}

func NewAddCommand(
	adminList []string, logChannelID string, giverRole string,
	nonAdminMaxPoints int, thresh time.Duration,
	repo Repository, clock Clock, logger *slog.Logger,
) *AddCommand {
	if nonAdminMaxPoints < minNonAdminMaxPoints {
		nonAdminMaxPoints = minNonAdminMaxPoints
	}

	return &AddCommand{
		adminList:           adminList,
		logChannelID:        logChannelID,
		giverRole:           giverRole,
		nonAdminMaxPoints:   nonAdminMaxPoints,
		punishmentThreshold: thresh,
		repo:                repo,
		clock:               clock,
		logger:              logger,
	}
}

func NewGetCommand(adminList []string, logChannelID string, repo Repository, logger *slog.Logger) *GetCommand {
	return &GetCommand{adminList: adminList, logChannelID: logChannelID, repo: repo, logger: logger}
}

func NewListCommand(adminList []string, logChannelID string, repo Repository, logger *slog.Logger) *ListCommand {
	return &ListCommand{adminList: adminList, logChannelID: logChannelID, repo: repo, logger: logger}
}

func NewRegisterCommand(adminList []string, logChannelID string, repo Repository, logger *slog.Logger) *RegisterCommand {
	return &RegisterCommand{adminList: adminList, logChannelID: logChannelID, repo: repo, logger: logger}
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
		canGiveAngyPoints := false
		if giverRole != "" {
			canGiveAngyPoints = slices.Contains(i.Member.Roles, giverRole)
		}

		return i.Member.User, canGiveAngyPoints, nil
	default:
		return nil, false, ErrInteractionUserIsNil
	}
}

func getAngyPointsAndUserID(s *discordgo.Session, values []*discordgo.ApplicationCommandInteractionDataOption) (int, *discordgo.User, error) {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(values))
	for _, opt := range values {
		optionMap[opt.Name] = opt
	}

	cookies, ok := optionMap["angy_points"]
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
