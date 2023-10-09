package filters

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	pb "github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
)

var ValidPrinters = map[string]func(context.Context, *slog.Logger, []byte){
	"developers":          PrintDevelopers,
	"publishers":          PrintPublishers,
	"demos":               PrintDemos,
	"price_overview":      PrintPriceOverview,
	"packages":            PrintPackages,
	"platforms":           PrintPlatforms,
	"categories":          PrintCategories,
	"genres":              PrintGenres,
	"screenshots":         PrintScreenshots,
	"movies":              PrintMovies,
	"recommendations":     PrintRecommendations,
	"achievements":        PrintAchievements,
	"release_date":        PrintReleaseDate,
	"support_info":        PrintSupportInfo,
	"background":          PrintBackground,
	"content_descriptors": PrintContentDescriptors,
}

func printObject[T any, M map[string]T](
	ctx context.Context, logger *slog.Logger,
	name string, data M, err error, dataExtractor func(T) slog.Attr,
) {
	if err != nil {
		logger.ErrorContext(ctx, "unable to extract developers object", slog.String("error", err.Error()))

		os.Exit(1)
	}

	logger.InfoContext(ctx, fmt.Sprintf("received %s data", name),
		slog.Int("num_results", len(data)),
	)

	for appID, appData := range data {
		logger.InfoContext(ctx, "listing "+name,
			slog.String("app_id", appID),
			dataExtractor(appData),
		)
	}
}

func PrintDevelopers(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetDevelopers(data)

	printObject[*pb.DevelopersFilter](
		ctx, logger, "developers", content, err,
		func(data *pb.DevelopersFilter) slog.Attr {
			return slog.Any("developers", data.GetData().GetDevelopers())
		},
	)
}

func PrintPublishers(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetPublishers(data)

	printObject[*pb.PublishersFilter](
		ctx, logger, "publishers", content, err,
		func(data *pb.PublishersFilter) slog.Attr {
			return slog.Any("publishers", data.GetData().GetPublishers())
		},
	)
}

func PrintDemos(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetDemos(data)

	printObject[*pb.DemosFilter](
		ctx, logger, "demos", content, err,
		func(data *pb.DemosFilter) slog.Attr {
			demos := data.GetData().GetDemos()
			if len(demos) == 0 {
				return slog.Attr{}
			}

			ids := make([]int64, 0, len(demos))
			for i := range demos {
				ids = append(ids, demos[i].GetAppid())
			}

			return slog.Any("demos", ids)
		},
	)
}

func PrintPriceOverview(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetPriceOverview(data)

	printObject[*pb.PriceOverviewFilter](
		ctx, logger, "price_overview", content, err,
		func(data *pb.PriceOverviewFilter) slog.Attr {
			return slog.Group("price_overview",
				slog.String("currency", data.GetData().GetPriceOverview().GetCurrency()),
				slog.String("initial", data.GetData().GetPriceOverview().GetFinalFormatted()),
				slog.String("current", data.GetData().GetPriceOverview().GetInitialFormatted()),
				slog.Int("discount_percent", int(data.GetData().GetPriceOverview().GetDiscountPercent())),
			)
		},
	)
}

func PrintPackages(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetPackages(data)

	printObject[*pb.PackagesFilter](
		ctx, logger, "packages", content, err,
		func(data *pb.PackagesFilter) slog.Attr {
			groups := data.GetData().GetPackageGroups()
			groupTitles := make([]string, 0, len(groups))

			for i := range groups {
				subs := groups[i].GetSubs()
				subsIDs := &strings.Builder{}
				for idx := range subs {
					subsIDs.WriteString(strconv.Itoa(int(subs[idx].GetPackageid())))

					if idx < len(subs)-1 {
						subsIDs.WriteByte(':')
					}
				}

				groupTitles = append(groupTitles, groups[i].GetTitle()+"::"+subsIDs.String())
			}

			return slog.Group("packages",
				slog.Any("packages", data.GetData().GetPackages()),
				slog.Any("package_groups", groupTitles),
			)
		},
	)
}

func PrintPlatforms(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetPlatforms(data)

	printObject[*pb.PlatformsFilter](
		ctx, logger, "platforms", content, err,
		func(data *pb.PlatformsFilter) slog.Attr {
			return slog.Group("platforms",
				slog.Bool("on_windows", data.GetData().GetPlatforms().GetWindows()),
				slog.Bool("on_mac", data.GetData().GetPlatforms().GetMac()),
				slog.Bool("on_linux", data.GetData().GetPlatforms().GetLinux()),
			)
		},
	)
}

func PrintCategories(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetCategories(data)

	printObject[*pb.CategoriesFilter](
		ctx, logger, "categories", content, err,
		func(data *pb.CategoriesFilter) slog.Attr {
			cat := data.GetData().GetCategories()
			ids := make([]int64, 0, len(cat))
			descs := make([]string, 0, len(cat))

			for i := range cat {
				ids = append(ids, cat[i].GetId())
				descs = append(descs, cat[i].GetDescription())
			}

			return slog.Group("categories",
				slog.Any("ids", ids),
				slog.Any("descriptions", descs),
			)
		},
	)
}

func PrintGenres(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetGenres(data)

	printObject[*pb.GenresFilter](
		ctx, logger, "genres", content, err,
		func(data *pb.GenresFilter) slog.Attr {
			genres := data.GetData().GetGenres()
			ids := make([]int64, 0, len(genres))
			descs := make([]string, 0, len(genres))

			for i := range genres {
				ids = append(ids, genres[i].GetId())
				descs = append(descs, genres[i].GetDescription())
			}

			return slog.Group("genres",
				slog.Any("ids", ids),
				slog.Any("descriptions", descs),
			)
		},
	)
}

func PrintScreenshots(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetScreenshots(data)

	printObject[*pb.ScreenshotsFilter](
		ctx, logger, "screenshots", content, err,
		func(data *pb.ScreenshotsFilter) slog.Attr {
			screenshots := data.GetData().GetScreenshots()
			urls := make([]string, 0, len(screenshots))

			for i := range screenshots {
				urls = append(urls, screenshots[i].GetPathFull())
			}

			return slog.Any("screenshots", urls)
		},
	)
}

func PrintMovies(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetMovies(data)

	printObject[*pb.MoviesFilter](
		ctx, logger, "movies", content, err,
		func(data *pb.MoviesFilter) slog.Attr {
			movies := data.GetData().GetMovies()
			moviesList := make([]slog.Attr, 0, len(movies))

			for i := range movies {
				var webmAttr slog.Attr
				var mp4Attr slog.Attr

				webm := movies[i].GetWebm()
				if wembMax, ok := webm["max"]; ok {
					webmAttr = slog.String("webm", wembMax)
				}

				mp4 := movies[i].GetMp4()
				if mp4Max, ok := mp4["max"]; ok {
					mp4Attr = slog.String("webm", mp4Max)
				}

				moviesList = append(moviesList,
					slog.Group(strconv.Itoa(int(movies[i].GetId())),
						slog.String("name", movies[i].GetName()),
						slog.Bool("is_highlight", movies[i].GetHighlight()),
						webmAttr, mp4Attr,
					),
				)
			}

			return slog.Any("movies", moviesList)
		},
	)
}

func PrintRecommendations(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetRecommendations(data)

	printObject[*pb.RecommendationsFilter](
		ctx, logger, "recommendations", content, err,
		func(data *pb.RecommendationsFilter) slog.Attr {
			return slog.Int("recommendations", int(data.GetData().GetRecommendations().GetTotal()))
		},
	)
}

func PrintAchievements(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetAchievements(data)

	printObject[*pb.AchievementsFilter](
		ctx, logger, "achievements", content, err,
		func(data *pb.AchievementsFilter) slog.Attr {
			highlighted := data.GetData().GetAchievements().GetHighlighted()
			hl := make([]string, 0, len(highlighted))

			for i := range highlighted {
				hl = append(hl, highlighted[i].GetName())
			}

			return slog.Group("achievements",
				slog.Int("total", int(data.GetData().GetAchievements().GetTotal())),
				slog.Any("highlighted", hl),
			)
		},
	)
}

func PrintReleaseDate(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetReleaseDate(data)

	printObject[*pb.ReleaseDateFilter](
		ctx, logger, "release_date", content, err,
		func(data *pb.ReleaseDateFilter) slog.Attr {
			return slog.String("release_date", data.GetData().GetReleaseDate().GetDate())
		},
	)
}

func PrintSupportInfo(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetSupportInfo(data)

	printObject[*pb.SupportInfoFilter](
		ctx, logger, "support_info", content, err,
		func(data *pb.SupportInfoFilter) slog.Attr {
			return slog.Group("support_info",
				slog.String("url", data.GetData().GetSupportInfo().GetUrl()),
				slog.String("email", data.GetData().GetSupportInfo().GetEmail()),
			)
		},
	)
}

func PrintBackground(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetBackground(data)

	printObject[*pb.BackgroundFilter](
		ctx, logger, "background", content, err,
		func(data *pb.BackgroundFilter) slog.Attr {
			return slog.Any("background", data.GetData().GetBackground())
		},
	)
}

func PrintContentDescriptors(ctx context.Context, logger *slog.Logger, data []byte) {
	content, err := GetContentDescriptors(data)

	printObject[*pb.ContentDescriptorsFilter](
		ctx, logger, "content_descriptors", content, err,
		func(data *pb.ContentDescriptorsFilter) slog.Attr {
			return slog.Group("content_descriptors",
				slog.Any("ids", data.GetData().GetContentDescriptors().GetIds()),
				slog.String("notes", data.GetData().GetContentDescriptors().GetNotes()),
			)
		},
	)
}
