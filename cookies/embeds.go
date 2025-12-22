package cookies

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func EmbedNoPermissionsToSendCookies(requester string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 You wish you could give cookies",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Tries to give cookies",
				Value:  "*total cinema*",
				Inline: true,
			},
		},
	}
}

func EmbedSpammingCookies(requester string, wait time.Duration) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 You wish you could spam cookies",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Tries to give cookies",
				Value:  "*total cinema*",
				Inline: true,
			}, {
				Name:   "Needs to wait",
				Value:  wait.String(),
				Inline: true,
			},
		},
	}
}

func EmbedAddedCookies(label, requester, user string, n int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🍪 Added cookies",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", user, user),
				Inline: false,
			}, {
				Name:   "Total cookies",
				Value:  strconv.Itoa(n),
				Inline: true,
			}, {
				Name:   "Operation",
				Value:  fmt.Sprintf("%d cookies by <@%s> (`%s`)", n, requester, requester),
				Inline: true,
			}, {
				Name:   "Rank",
				Value:  GetRank(n).String(),
				Inline: true,
			},
		},
	}
}

func EmbedTotalCookies(label, user string, n int, lastGift time.Time) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🍪 Fetching total amount of cookies",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", user, user),
				Inline: false,
			}, {
				Name:   "Total cookies",
				Value:  strconv.Itoa(n),
				Inline: true,
			}, {
				Name:   "Last time gifted",
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

func EmbedListCookiesForNoUsers(label string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🍪 Fetching total amount of cookies",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  "no users were assigned cookies, yet",
				Inline: false,
			},
		},
	}
}

func EmbedListCookiesForUser(label, user string, cookies int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🍪 Fetching cookies for %s", label),
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", user, user),
				Inline: false,
			}, {
				Name:   "Total cookies",
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

func EmbedCookieStealer(requester, user string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 f***ing cookie stealer",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "You filthy thief >:c",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   fmt.Sprintf("Tries to STEAL cookies from <@%s> (`%s`)", user, user),
				Value:  "*total cinema*",
				Inline: true,
			},
		},
	}
}

func EmbedNotEnoughCookies(requester string, current, cookies int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 Your broke ass doesn't even have enough cookies",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Tries to give cookies",
				Value:  "*total cinema*",
				Inline: true,
			}, {
				Name:   "Currently has",
				Value:  strconv.Itoa(current),
				Inline: true,
			}, {
				Name:   "Wants to give",
				Value:  strconv.Itoa(cookies),
				Inline: true,
			},
		},
	}
}

func EmbedSharedCookies(label, requester, user string, cookies, requesterCurrent, targetCurrent int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🍪 Shared cookies",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Operation",
				Value:  fmt.Sprintf("shared %d cookies with <@%s> (`%s`)", cookies, user, user),
				Inline: true,
			}, {
				Name:   "Total cookies (sender)",
				Value:  strconv.Itoa(requesterCurrent),
				Inline: true,
			}, {
				Name:   "Total cookies (receiver)",
				Value:  strconv.Itoa(targetCurrent),
				Inline: true,
			},
		},
	}
}

func EmbedEatingWithoutAnyCookies(requester string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔒 you cannot eat other's crumbs",
		Color: 0xff0000, // red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Tries to eat cookies",
				Value:  "*total cinema*",
				Inline: true,
			}, {
				Name:   "Total cookies",
				Value:  "0",
				Inline: true,
			},
		},
	}
}

func EmbedEatingCookie(label, requester string, n int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🍪 Ate a cookie",
		Color: 0xaa7733, // https://www.color-hex.com/color-palette/9176
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   label,
				Value:  fmt.Sprintf("<@%s> (`%s`)", requester, requester),
				Inline: false,
			}, {
				Name:   "Total cookies",
				Value:  strconv.Itoa(n),
				Inline: true,
			}, {
				Name:   "Operation",
				Value:  fmt.Sprintf("Ate a cookie without leaving any crumbs"),
				Inline: true,
			}, {
				Name:   "Rank",
				Value:  GetRank(n).String(),
				Inline: true,
			},
		},
	}
}
