package store

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	pb "github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	headWrapStart = `{"`
	headWrapEnd   = `":`
	tail          = "}"
)

var validFilters = map[string]func(context.Context, *slog.Logger, []byte) error{
	"developers":          getDevelopers,
	"publishers":          getPublishers,
	"demos":               getDemos,
	"price_overview":      getPriceOverview,
	"packages":            getPackages,
	"platforms":           getPlatforms,
	"categories":          getCategories,
	"genres":              getGenres,
	"screenshots":         getScreenshots,
	"movies":              getMovies,
	"recommendations":     getRecommendations,
	"achievements":        getAchievements,
	"release_date":        getReleaseDate,
	"support_info":        getSupportInfo,
	"background":          getBackground,
	"content_descriptors": getContentDescriptors,
}

func getDevelopers(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.DevelopersResponse, *pb.DevelopersFilter](
		ctx, logger, "developers", data,
		func(response *pb.DevelopersResponse) map[string]*pb.DevelopersFilter {
			return response.GetDevelopers()
		},
		func(data *pb.DevelopersFilter) slog.Attr {
			return slog.Any("developers", data.GetData().GetDevelopers())
		},
	)
}

func getPublishers(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.PublishersResponse, *pb.PublishersFilter](
		ctx, logger, "publishers", data,
		func(response *pb.PublishersResponse) map[string]*pb.PublishersFilter {
			return response.GetPublishers()
		},
		func(data *pb.PublishersFilter) slog.Attr {
			return slog.Any("publishers", data.GetData().GetPublishers())
		},
	)
}

func getDemos(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.DemosResponse, *pb.DemosFilter](
		ctx, logger, "demos", data,
		func(response *pb.DemosResponse) map[string]*pb.DemosFilter {
			return response.GetDemos()
		},
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

func getPriceOverview(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.PriceOverviewResponse, *pb.PriceOverviewFilter](
		ctx, logger, "price_overview", data,
		func(response *pb.PriceOverviewResponse) map[string]*pb.PriceOverviewFilter {
			return response.GetPriceOverview()
		},
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

func getPackages(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.PackagesResponse, *pb.PackagesFilter](
		ctx, logger, "packages", data,
		func(response *pb.PackagesResponse) map[string]*pb.PackagesFilter {
			return response.GetPackages()
		},
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

func getPlatforms(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.PlatformsResponse, *pb.PlatformsFilter](
		ctx, logger, "platforms", data,
		func(response *pb.PlatformsResponse) map[string]*pb.PlatformsFilter {
			return response.GetPlatforms()
		},
		func(data *pb.PlatformsFilter) slog.Attr {
			return slog.Group("platforms",
				slog.Bool("on_windows", data.GetData().GetPlatforms().GetWindows()),
				slog.Bool("on_mac", data.GetData().GetPlatforms().GetMac()),
				slog.Bool("on_linux", data.GetData().GetPlatforms().GetLinux()),
			)
		},
	)
}

func getCategories(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.CategoriesResponse, *pb.CategoriesFilter](
		ctx, logger, "categories", data,
		func(response *pb.CategoriesResponse) map[string]*pb.CategoriesFilter {
			return response.GetCategories()
		},
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

func getGenres(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.GenresResponse, *pb.GenresFilter](
		ctx, logger, "genres", data,
		func(response *pb.GenresResponse) map[string]*pb.GenresFilter {
			return response.GetGenres()
		},
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

func getScreenshots(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.ScreenshotsResponse, *pb.ScreenshotsFilter](
		ctx, logger, "screenshots", data,
		func(response *pb.ScreenshotsResponse) map[string]*pb.ScreenshotsFilter {
			return response.GetScreenshots()
		},
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

func getMovies(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.MoviesResponse, *pb.MoviesFilter](
		ctx, logger, "movies", data,
		func(response *pb.MoviesResponse) map[string]*pb.MoviesFilter {
			return response.GetMovies()
		},
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

func getRecommendations(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.RecommendationsResponse, *pb.RecommendationsFilter](
		ctx, logger, "recommendations", data,
		func(response *pb.RecommendationsResponse) map[string]*pb.RecommendationsFilter {
			return response.GetRecommendations()
		},
		func(data *pb.RecommendationsFilter) slog.Attr {
			return slog.Int("recommendations", int(data.GetData().GetRecommendations().GetTotal()))
		},
	)
}

func getAchievements(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.AchievementsResponse, *pb.AchievementsFilter](
		ctx, logger, "achievements", data,
		func(response *pb.AchievementsResponse) map[string]*pb.AchievementsFilter {
			return response.GetAchievements()
		},
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

func getReleaseDate(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.ReleaseDateResponse, *pb.ReleaseDateFilter](
		ctx, logger, "release_date", data,
		func(response *pb.ReleaseDateResponse) map[string]*pb.ReleaseDateFilter {
			return response.GetReleaseDate()
		},
		func(data *pb.ReleaseDateFilter) slog.Attr {
			return slog.String("release_date", data.GetData().GetReleaseDate().GetDate())
		},
	)
}

func getSupportInfo(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.SupportInfoResponse, *pb.SupportInfoFilter](
		ctx, logger, "support_info", data,
		func(response *pb.SupportInfoResponse) map[string]*pb.SupportInfoFilter {
			return response.GetSupportInfo()
		},
		func(data *pb.SupportInfoFilter) slog.Attr {
			return slog.Group("support_info",
				slog.String("url", data.GetData().GetSupportInfo().GetUrl()),
				slog.String("email", data.GetData().GetSupportInfo().GetEmail()),
			)
		},
	)
}

func getBackground(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.BackgroundResponse, *pb.BackgroundFilter](
		ctx, logger, "background", data,
		func(response *pb.BackgroundResponse) map[string]*pb.BackgroundFilter {
			return response.GetBackground()
		},
		func(data *pb.BackgroundFilter) slog.Attr {
			return slog.Any("background", data.GetData().GetBackground())
		},
	)
}

func getContentDescriptors(ctx context.Context, logger *slog.Logger, data []byte) error {
	return getObject[pb.ContentDescriptorsResponse, *pb.ContentDescriptorsFilter](
		ctx, logger, "content_descriptors", data,
		func(response *pb.ContentDescriptorsResponse) map[string]*pb.ContentDescriptorsFilter {
			return response.GetContentDescriptors()
		},
		func(data *pb.ContentDescriptorsFilter) slog.Attr {
			return slog.Group("content_descriptors",
				slog.Any("ids", data.GetData().GetContentDescriptors().GetIds()),
				slog.String("notes", data.GetData().GetContentDescriptors().GetNotes()),
			)
		},
	)
}

func getObject[T, K any](
	ctx context.Context, logger *slog.Logger,
	name string, data []byte,
	responseExtractor func(*T) map[string]K,
	dataExtractor func(K) slog.Attr,
) error {
	header := wrapHeader(name)

	response := new(T)
	t, ok := any(response).(proto.Message)
	if !ok {
		return fmt.Errorf("input type is not a proto.Message type: %T", t)
	}

	if err := protojson.Unmarshal(addWrapper(header, data), t); err != nil {
		return err
	}

	res := responseExtractor(response)
	logger.InfoContext(ctx, fmt.Sprintf("received %s data", name),
		slog.Int("num_results", len(res)),
	)

	for appID, appData := range res {
		logger.InfoContext(ctx, "listing "+name,
			slog.String("app_id", appID),
			dataExtractor(appData),
		)
	}

	return nil
}

func addWrapper(header string, data []byte) []byte {
	if len(data) < len(header) {
		return data
	}

	if bytes.Equal(data[:len(header)], []byte(header)) {
		return data
	}

	buf := make([]byte, len(data)+len(header)+len(tail))
	n := copy(buf, header)
	n += copy(buf[n:], data)
	copy(buf[n:], tail)

	return buf
}

func wrapHeader(header string) string {
	buf := make([]byte, len(header)+2+2)

	n := copy(buf, headWrapStart)
	n += copy(buf[n:], header)
	copy(buf[n:], headWrapEnd)

	return string(buf)
}
