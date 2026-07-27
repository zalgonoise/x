package politebot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func EmbedNoPermissionsToAddAngyPoints(requester string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 You wish you could give angy points",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Tries to give angy points",
				Value:  "*ABSOLULY CINEMA*",
				Inline: true,
			},
		},
	}
}

func EmbedSpammingAngyPoints(requester string, wait time.Duration) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 You wish you could spam angy points",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Tries to spam angy points",
				Value:  "*ABSOLULY CINEMA*",
				Inline: true,
			}, {
				Name:   "Needs to wait",
				Value:  wait.String(),
				Inline: true,
			},
		},
	}
}

func EmbedAngyPointsForUnknownUser(requester, target string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 You're giving points to a user who doesn't exist",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Target",
				Value:  fmt.Sprintf("<@%s> (`%s`)", target, target),
				Inline: false,
			}, {
				Name:   "Tries to add angy points to someone who doesn't exist",
				Value:  "*ABSOLULY CINEMA*",
				Inline: true,
			},
		},
	}
}

func EmbedAddedAngyPoints(label, requester, user string, n int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🤬 Added angy points",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", user, user),
				Inline: false,
			}, {
				Name:   "Total angy points",
				Value:  strconv.Itoa(n),
				Inline: true,
			}, {
				Name:   "Operation",
				Value:  fmt.Sprintf("%d angy points by <@%s> (`%s`)", n, requester, requester),
				Inline: true,
			}, {
				Name:   "Rank",
				Value:  GetRank(n).String(),
				Inline: true,
			},
		},
	}
}

func EmbedTotalAngyPoints(label, user string, n int, lastGift time.Time) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🤬 Fetching total amount of angy points",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", user, user),
				Inline: false,
			}, {
				Name:   "Total angy points",
				Value:  strconv.Itoa(n),
				Inline: true,
			}, {
				Name:   "Last time punished",
				Value:  lastGift.Format(time.DateTime),
				Inline: true,
			}, {
				Name:   "Rank",
				Value:  GetRank(n).String(),
				Inline: true,
			},
		},
	}
}

func EmbedListAngyPointsForNoUsers(label string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🤬 Fetching total amount of angy points",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  "no users were assigned angy points, yet",
				Inline: false,
			},
		},
	}
}

func EmbedListCookiesForUser(label, user string, cookies int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🤬 Fetching angy points for %s", label),
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", user, user),
				Inline: false,
			}, {
				Name:   "Total angy points",
				Value:  strconv.Itoa(cookies),
				Inline: true,
			}, {
				Name:   "Rank",
				Value:  GetRank(cookies).String(),
				Inline: true,
			},
		},
	}
}
