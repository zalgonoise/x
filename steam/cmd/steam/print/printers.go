package print

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/zalgonoise/x/steam"
	pb "github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
)

var ValidPrinters = map[string]func(context.Context, *slog.Logger, []byte){
	"developers":          Developers,
	"publishers":          Publishers,
	"demos":               Demos,
	"price_overview":      PriceOverview,
	"packages":            Packages,
	"platforms":           Platforms,
	"categories":          Categories,
	"genres":              Genres,
	"screenshots":         Screenshots,
	"movies":              Movies,
	"recommendations":     Recommendations,
	"achievements":        Achievements,
	"release_date":        ReleaseDate,
	"support_info":        SupportInfo,
	"background":          Background,
	"content_descriptors": ContentDescriptors,
}

func printObject[T any](
	ctx context.Context, logger *slog.Logger,
	name string, data []byte, dataExtractor func(*T) slog.Attr,
) {
	obj, err := steam.Get[T](data, name)
	if err != nil {
		logger.ErrorContext(ctx, "unable to extract object",
			slog.String("type", name),
			slog.String("error", err.Error()),
		)

		os.Exit(1)
	}

	logger.InfoContext(ctx, fmt.Sprintf("received %s data", name),
		slog.Int("num_results", len(obj)),
	)

	for appID, appData := range obj {
		logger.InfoContext(ctx, "listing "+name,
			slog.String("app_id", appID),
			dataExtractor(appData),
		)
	}
}

func Developers(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.DevelopersData](
		ctx, logger, "developers", data,
		func(data *pb.DevelopersData) slog.Attr {
			return slog.Any("developers", data.GetDevelopers())
		},
	)
}

func Publishers(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.PublishersData](
		ctx, logger, "publishers", data,
		func(data *pb.PublishersData) slog.Attr {
			return slog.Any("publishers", data.GetPublishers())
		},
	)
}

func Demos(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.DemosData](
		ctx, logger, "demos", data,
		func(data *pb.DemosData) slog.Attr {
			demos := data.GetDemos()
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

func PriceOverview(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.PriceOverview](
		ctx, logger, "price_overview", data,
		func(data *pb.PriceOverview) slog.Attr {
			return slog.Group("price_overview",
				slog.String("currency", data.GetCurrency()),
				slog.String("initial", data.GetFinalFormatted()),
				slog.String("current", data.GetInitialFormatted()),
				slog.Int("discount_percent", int(data.GetDiscountPercent())),
			)
		},
	)
}

func Packages(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.PackagesData](
		ctx, logger, "packages", data,
		func(data *pb.PackagesData) slog.Attr {
			groups := data.GetPackageGroups()
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
				slog.Any("packages", data.GetPackages()),
				slog.Any("package_groups", groupTitles),
			)
		},
	)
}

func Platforms(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.Platforms](
		ctx, logger, "platforms", data,
		func(data *pb.Platforms) slog.Attr {
			return slog.Group("platforms",
				slog.Bool("on_windows", data.GetWindows()),
				slog.Bool("on_mac", data.GetMac()),
				slog.Bool("on_linux", data.GetLinux()),
			)
		},
	)
}

func Categories(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.CategoriesData](
		ctx, logger, "categories", data,
		func(data *pb.CategoriesData) slog.Attr {
			cat := data.GetCategories()
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

func Genres(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.GenresData](
		ctx, logger, "genres", data,
		func(data *pb.GenresData) slog.Attr {
			genres := data.GetGenres()
			ids := make([]string, 0, len(genres))
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

func Screenshots(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.ScreenshotsData](
		ctx, logger, "screenshots", data,
		func(data *pb.ScreenshotsData) slog.Attr {
			screenshots := data.GetScreenshots()
			urls := make([]string, 0, len(screenshots))

			for i := range screenshots {
				urls = append(urls, screenshots[i].GetPathFull())
			}

			return slog.Any("screenshots", urls)
		},
	)
}

func Movies(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.MoviesData](
		ctx, logger, "movies", data,
		func(data *pb.MoviesData) slog.Attr {
			movies := data.GetMovies()
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

func Recommendations(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.Recommendations](
		ctx, logger, "recommendations", data,
		func(data *pb.Recommendations) slog.Attr {
			return slog.Int("recommendations", int(data.GetTotal()))
		},
	)
}

func Achievements(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.Achievements](
		ctx, logger, "achievements", data,
		func(data *pb.Achievements) slog.Attr {
			highlighted := data.GetHighlighted()
			hl := make([]string, 0, len(highlighted))

			for i := range highlighted {
				hl = append(hl, highlighted[i].GetName())
			}

			return slog.Group("achievements",
				slog.Int("total", int(data.GetTotal())),
				slog.Any("highlighted", hl),
			)
		},
	)
}

func ReleaseDate(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.ReleaseDate](
		ctx, logger, "release_date", data,
		func(data *pb.ReleaseDate) slog.Attr {
			return slog.String("release_date", data.GetDate())
		},
	)
}

func SupportInfo(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.SupportInfo](
		ctx, logger, "support_info", data,
		func(data *pb.SupportInfo) slog.Attr {
			return slog.Group("support_info",
				slog.String("url", data.GetUrl()),
				slog.String("email", data.GetEmail()),
			)
		},
	)
}

func Background(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.BackgroundData](
		ctx, logger, "background", data,
		func(data *pb.BackgroundData) slog.Attr {
			return slog.Any("background", data.GetBackground())
		},
	)
}

func ContentDescriptors(ctx context.Context, logger *slog.Logger, data []byte) {
	printObject[pb.ContentDescriptorsData](
		ctx, logger, "content_descriptors", data,
		func(data *pb.ContentDescriptorsData) slog.Attr {
			return slog.Group("content_descriptors",
				slog.Any("ids", data.GetContentDescriptors().GetIds()),
				slog.String("notes", data.GetContentDescriptors().GetNotes()),
			)
		},
	)
}
