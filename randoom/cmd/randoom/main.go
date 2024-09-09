package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"

	"github.com/google/uuid"
	"github.com/zalgonoise/x/cli"

	"github.com/zalgonoise/x/randoom/config"
	"github.com/zalgonoise/x/randoom/database"
	"github.com/zalgonoise/x/randoom/random"
	"github.com/zalgonoise/x/randoom/repository"
)

var modes = []string{"server", "item", "playlist", "playlist-done"}

func main() {
	runner := cli.NewRunner("randoom",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"server":        cli.Executable(ExecServer),
			"item":          cli.Executable(ExecItem),
			"playlist":      cli.Executable(ExecPlaylist),
			"playlist-done": cli.Executable(ExecPlaylistDone),
		}),
	)

	cli.Run(runner)
}

func setupService(name string, args []string) (*random.Service, *config.Config, error) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)

	path := fs.String("config", "", "path to config file")
	dbURI := fs.String("uri", "", "the SQLite database URI")
	content := fs.String("content", "", "the list's content")
	playlistID := fs.String("playlist-id", "", "the ID of the playlist to mark as completed")

	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}

	cfg, err := config.OpenConfigFile(*path)
	if err != nil && !errors.Is(err, config.ErrInvalidConfig) {
		return nil, nil, err
	}

	if err != nil {
		// use config from flags
		cfg, err = config.ParseContent(*dbURI, *content, *playlistID)
		if err != nil {
			return nil, nil, err
		}
	}

	// use config
	db, err := database.Open(cfg.DatabaseURI)
	if err != nil {
		return nil, nil, err
	}

	if err = database.Migrate(context.Background(), db); err != nil {
		return nil, nil, err
	}

	return random.NewService(repository.NewRepository(db), cfg), cfg, nil
}

func ExecServer(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	service, _, err := setupService("server", args)
	if err != nil {
		return 1, err
	}

	// TODO: gRPC logic for HTTP
	_ = service

	return 0, nil
}

func ExecItem(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	service, _, err := setupService("server", args)
	if err != nil {
		return 1, err
	}

	item, err := service.GetRandom()
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "new random item",
		slog.String("content", item.Content),
		slog.Uint64("count", item.Count),
		slog.Float64("ratio", float64(item.Ratio)),
	)

	return 0, nil
}

func ExecPlaylist(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	service, _, err := setupService("server", args)
	if err != nil {
		return 1, err
	}

	id, items, err := service.NewPlaylist()
	if err != nil {
		return 1, err
	}

	g := make([]slog.Attr, 0, len(items))
	for i := range items {
		g = append(g, slog.Group("item",
			slog.String("content", items[i].Content),
			slog.Uint64("count", items[i].Count),
			slog.Float64("ratio", float64(items[i].Ratio)),
		))
	}

	logger.InfoContext(ctx, "new playlist generated",
		slog.String("playlist_id", id.String()),
		slog.Group("items", g),
	)

	return 0, nil
}

func ExecPlaylistDone(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	service, cfg, err := setupService("server", args)
	if err != nil {
		return 1, err
	}

	id, err := uuid.Parse(cfg.PlaylistID)
	if err != nil {
		return 1, err
	}

	if err := service.ClosePlaylist(id); err != nil {
		return 1, err
	}

	return 0, nil
}
