package apps

import (
	"log/slog"

	"github.com/zalgonoise/x/fts"
	"github.com/zalgonoise/x/ptr"
	"github.com/zalgonoise/x/steam"
)

func NewIndexer(uri string, logger *slog.Logger) (fts.Indexer[int64, string], error) {
	appsList, err := steam.LoadAppsList()
	if err != nil {
		return nil, err
	}

	index, err := fts.New(
		ptr.Cast[[]fts.Attribute[int64, string]](appsList),
		fts.WithURI(uri),
		fts.WithLogger(logger),
	)

	if err != nil {
		return nil, err
	}

	return index, nil
}
