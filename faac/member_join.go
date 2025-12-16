package faac

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

func MemberAccountAgeFilter(logger *slog.Logger, minDaysAge int, logChannelID string, allowedList ...string) func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		ctx := context.Background()

		// 1. Calculate Account Age
		creationTime, err := discordgo.SnowflakeTimestamp(m.User.ID)
		if err != nil {
			logger.ErrorContext(ctx, "getting snowflake timestamp",
				slog.String("error", err.Error()),
				slog.String("user", m.User.Username),
				slog.String("user_id", m.User.ID),
			)

			return
		}

		accountAge := time.Since(creationTime)

		if accountAge == 0 {
			logger.ErrorContext(ctx, "account age cannot be zero",
				slog.String("user", m.User.Username),
				slog.String("user_id", m.User.ID),
			)

			return
		}

		daysOld := int(accountAge.Hours() / 24)

		// 2. Check if user is allowed to bypass this check
		if len(allowedList) > 0 && slices.Contains(allowedList, m.User.ID) {
			logger.WarnContext(ctx, "allow-list user has bypassed the auto-kick rule",
				slog.String("user", m.User.Username),
				slog.String("user_id", m.User.ID))

			if _, err := s.ChannelMessageSendEmbed(logChannelID, &discordgo.MessageEmbed{
				Title: "🛡️ Auto-Kick Bypassed",
				Color: 0xffff00, // Yellow
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Automated user-kick bypassed (allow-list)",
						Value:  fmt.Sprintf("%s (`%s`)", m.User.Mention(), m.User.ID),
						Inline: true,
					},
					{
						Name:   "Account Age",
						Value:  fmt.Sprintf("**%d days** (Threshold: %d)", daysOld, minDaysAge),
						Inline: true,
					},
					{
						Name:   "Creation Date",
						Value:  creationTime.Format("2006-01-02 15:04:05"),
						Inline: false,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Allow-list user has bypassed the server removal rule.",
				},
			}); err != nil {
				logger.ErrorContext(ctx, "notifying bypass of user kick",
					slog.String("error", err.Error()),
					slog.String("user", m.User.Username),
					slog.String("user_id", m.User.ID),
				)
			}

			return
		}

		// 3. Check threshold
		if daysOld < minDaysAge {
			logger.InfoContext(ctx, "suspicious account detected",
				slog.Int("account_age", daysOld),
				slog.String("user", m.User.Username),
				slog.String("user_id", m.User.ID))

			// 4. Kick the User
			// Parameters: Guild ID, User ID
			if err := s.GuildMemberDelete(m.GuildID, m.User.ID); err != nil {
				logger.ErrorContext(ctx, "kicking account from server",
					slog.String("error", err.Error()),
					slog.String("user", m.User.Username),
					slog.String("user_id", m.User.ID),
				)

				// Optional: Alert channel that kick failed (usually due to permissions)
				if _, err := s.ChannelMessageSend(logChannelID,
					fmt.Sprintf("⚠️ **Error:** Tried to kick <@%s> for having an excessively recent account,"+
						"but failed. Check my permissions.", m.User.ID)); err != nil {
					logger.ErrorContext(ctx, "notifying user kick failure",
						slog.String("error", err.Error()),
						slog.String("user", m.User.Username),
						slog.String("user_id", m.User.ID),
					)
				}

				return
			}

			// 5. Log the Kick to the Channel
			if _, err := s.ChannelMessageSendEmbed(logChannelID, &discordgo.MessageEmbed{
				Title: "🛡️ Auto-Kick Triggered",
				Color: 0xff0000, // Red
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User Kicked",
						Value:  fmt.Sprintf("%s (`%s`)", m.User.Mention(), m.User.ID),
						Inline: true,
					},
					{
						Name:   "Account Age",
						Value:  fmt.Sprintf("**%d days** (Threshold: %d)", daysOld, minDaysAge),
						Inline: true,
					},
					{
						Name:   "Creation Date",
						Value:  creationTime.Format("2006-01-02 15:04:05"),
						Inline: false,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "User has been removed from the server.",
				},
			}); err != nil {
				logger.ErrorContext(ctx, "notifying user kick",
					slog.String("error", err.Error()),
					slog.String("user", m.User.Username),
					slog.String("user_id", m.User.ID),
				)
			}
		}
	}
}
