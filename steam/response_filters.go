package steam

import pb "github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"

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

func GetData(data []byte) (map[string]*pb.Data, error) {
	return Get[pb.Data](data, "")
}

func GetDevelopers(data []byte) (map[string]*pb.DevelopersData, error) {
	return Get[pb.DevelopersData](data, "developers")
}

func GetPublishers(data []byte) (map[string]*pb.PublishersData, error) {
	return Get[pb.PublishersData](data, "publishers")
}

func GetDemos(data []byte) (map[string]*pb.DemosData, error) {
	return Get[pb.DemosData](data, "demos")
}

func GetPriceOverview(data []byte) (map[string]*pb.PriceOverview, error) {
	return Get[pb.PriceOverview](data, "price_overview")
}

func GetPackages(data []byte) (map[string]*pb.PackagesData, error) {
	return Get[pb.PackagesData](data, "packages")
}

func GetPlatforms(data []byte) (map[string]*pb.Platforms, error) {
	return Get[pb.Platforms](data, "platforms")
}

func GetCategories(data []byte) (map[string]*pb.CategoriesData, error) {
	return Get[pb.CategoriesData](data, "categories")
}

func GetGenres(data []byte) (map[string]*pb.GenresData, error) {
	return Get[pb.GenresData](data, "genres")
}

func GetScreenshots(data []byte) (map[string]*pb.ScreenshotsData, error) {
	return Get[pb.ScreenshotsData](data, "screenshots")
}

func GetMovies(data []byte) (map[string]*pb.MoviesData, error) {
	return Get[pb.MoviesData](data, "movies")
}

func GetRecommendations(data []byte) (map[string]*pb.Recommendations, error) {
	return Get[pb.Recommendations](data, "recommendations")
}

func GetAchievements(data []byte) (map[string]*pb.Achievements, error) {
	return Get[pb.Achievements](data, "achievements")
}

func GetReleaseDate(data []byte) (map[string]*pb.ReleaseDate, error) {
	return Get[pb.ReleaseDate](data, "release_date")
}

func GetSupportInfo(data []byte) (map[string]*pb.SupportInfo, error) {
	return Get[pb.SupportInfo](data, "support_info")
}

func GetBackground(data []byte) (map[string]*pb.BackgroundData, error) {
	return Get[pb.BackgroundData](data, "background")
}

func GetContentDescriptors(data []byte) (map[string]*pb.ContentDescriptorsData, error) {
	return Get[pb.ContentDescriptorsData](data, "content_descriptors")
}
