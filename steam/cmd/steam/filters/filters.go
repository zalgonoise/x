package filters

import (
	"bytes"
	"fmt"

	pb "github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	headWrapStart = `{"`
	headWrapEnd   = `":`
	tail          = "}"
)

var ValidFilters = map[string]struct{}{
	"developers":          {},
	"publishers":          {},
	"demos":               {},
	"price_overview":      {},
	"packages":            {},
	"platforms":           {},
	"categories":          {},
	"genres":              {},
	"screenshots":         {},
	"movies":              {},
	"recommendations":     {},
	"achievements":        {},
	"release_date":        {},
	"support_info":        {},
	"background":          {},
	"content_descriptors": {},
}

func GetDevelopers(data []byte) (map[string]*pb.DevelopersFilter, error) {
	return getObject[pb.DevelopersResponse, *pb.DevelopersFilter](
		"developers", data,
		func(response *pb.DevelopersResponse) map[string]*pb.DevelopersFilter {
			return response.GetDevelopers()
		},
	)
}

func GetPublishers(data []byte) (map[string]*pb.PublishersFilter, error) {
	return getObject[pb.PublishersResponse, *pb.PublishersFilter](
		"publishers", data,
		func(response *pb.PublishersResponse) map[string]*pb.PublishersFilter {
			return response.GetPublishers()
		},
	)
}

func GetDemos(data []byte) (map[string]*pb.DemosFilter, error) {
	return getObject[pb.DemosResponse, *pb.DemosFilter](
		"demos", data,
		func(response *pb.DemosResponse) map[string]*pb.DemosFilter {
			return response.GetDemos()
		},
	)
}

func GetPriceOverview(data []byte) (map[string]*pb.PriceOverviewFilter, error) {
	return getObject[pb.PriceOverviewResponse, *pb.PriceOverviewFilter](
		"price_overview", data,
		func(response *pb.PriceOverviewResponse) map[string]*pb.PriceOverviewFilter {
			return response.GetPriceOverview()
		},
	)
}

func GetPackages(data []byte) (map[string]*pb.PackagesFilter, error) {
	return getObject[pb.PackagesResponse, *pb.PackagesFilter](
		"packages", data,
		func(response *pb.PackagesResponse) map[string]*pb.PackagesFilter {
			return response.GetPackages()
		},
	)
}

func GetPlatforms(data []byte) (map[string]*pb.PlatformsFilter, error) {
	return getObject[pb.PlatformsResponse, *pb.PlatformsFilter](
		"platforms", data,
		func(response *pb.PlatformsResponse) map[string]*pb.PlatformsFilter {
			return response.GetPlatforms()
		},
	)
}

func GetCategories(data []byte) (map[string]*pb.CategoriesFilter, error) {
	return getObject[pb.CategoriesResponse, *pb.CategoriesFilter](
		"categories", data,
		func(response *pb.CategoriesResponse) map[string]*pb.CategoriesFilter {
			return response.GetCategories()
		},
	)
}

func GetGenres(data []byte) (map[string]*pb.GenresFilter, error) {
	return getObject[pb.GenresResponse, *pb.GenresFilter](
		"genres", data,
		func(response *pb.GenresResponse) map[string]*pb.GenresFilter {
			return response.GetGenres()
		},
	)
}

func GetScreenshots(data []byte) (map[string]*pb.ScreenshotsFilter, error) {
	return getObject[pb.ScreenshotsResponse, *pb.ScreenshotsFilter](
		"screenshots", data,
		func(response *pb.ScreenshotsResponse) map[string]*pb.ScreenshotsFilter {
			return response.GetScreenshots()
		},
	)
}

func GetMovies(data []byte) (map[string]*pb.MoviesFilter, error) {
	return getObject[pb.MoviesResponse, *pb.MoviesFilter](
		"movies", data,
		func(response *pb.MoviesResponse) map[string]*pb.MoviesFilter {
			return response.GetMovies()
		},
	)
}

func GetRecommendations(data []byte) (map[string]*pb.RecommendationsFilter, error) {
	return getObject[pb.RecommendationsResponse, *pb.RecommendationsFilter](
		"recommendations", data,
		func(response *pb.RecommendationsResponse) map[string]*pb.RecommendationsFilter {
			return response.GetRecommendations()
		},
	)
}

func GetAchievements(data []byte) (map[string]*pb.AchievementsFilter, error) {
	return getObject[pb.AchievementsResponse, *pb.AchievementsFilter](
		"achievements", data,
		func(response *pb.AchievementsResponse) map[string]*pb.AchievementsFilter {
			return response.GetAchievements()
		},
	)
}

func GetReleaseDate(data []byte) (map[string]*pb.ReleaseDateFilter, error) {
	return getObject[pb.ReleaseDateResponse, *pb.ReleaseDateFilter](
		"release_date", data,
		func(response *pb.ReleaseDateResponse) map[string]*pb.ReleaseDateFilter {
			return response.GetReleaseDate()
		},
	)
}

func GetSupportInfo(data []byte) (map[string]*pb.SupportInfoFilter, error) {
	return getObject[pb.SupportInfoResponse, *pb.SupportInfoFilter](
		"support_info", data,
		func(response *pb.SupportInfoResponse) map[string]*pb.SupportInfoFilter {
			return response.GetSupportInfo()
		},
	)
}

func GetBackground(data []byte) (map[string]*pb.BackgroundFilter, error) {
	return getObject[pb.BackgroundResponse, *pb.BackgroundFilter](
		"background", data,
		func(response *pb.BackgroundResponse) map[string]*pb.BackgroundFilter {
			return response.GetBackground()
		},
	)
}

func GetContentDescriptors(data []byte) (map[string]*pb.ContentDescriptorsFilter, error) {
	return getObject[pb.ContentDescriptorsResponse, *pb.ContentDescriptorsFilter](
		"content_descriptors", data,
		func(response *pb.ContentDescriptorsResponse) map[string]*pb.ContentDescriptorsFilter {
			return response.GetContentDescriptors()
		},
	)
}

func getObject[T, K any](
	name string, data []byte,
	responseExtractor func(*T) map[string]K,
) (map[string]K, error) {
	header := wrapHeader(name)

	response := new(T)
	t, ok := any(response).(proto.Message)
	if !ok {
		return nil, fmt.Errorf("input type is not a proto.Message type: %T", t)
	}

	if err := protojson.Unmarshal(addWrapper(header, data), t); err != nil {
		return nil, err
	}

	return responseExtractor(response), nil
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
