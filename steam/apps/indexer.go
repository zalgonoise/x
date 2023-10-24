package apps

import (
	"errors"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/fts"
	"github.com/zalgonoise/x/ptr"
	"github.com/zalgonoise/x/steam"
)

const inMemory = ":memory:"

func NewIndexer(uri string, logger *slog.Logger) (fts.Indexer[int64, string], error) {
	var (
		attr []fts.Attribute[int64, string]
		err  error
	)

	// perform an initial load if in-memory, or if (database) file does not exist
	switch uri {
	case "", inMemory:
		if attr, err = loadApps(); err != nil {
			return nil, err
		}
	default:
		_, err = os.Stat(uri)

		if err == nil {
			break
		}

		if errors.Is(err, os.ErrNotExist) {
			if attr, err = loadApps(); err != nil {
				return nil, err
			}

			break
		}

		return nil, err
	}

	indexer, err := fts.New(
		attr,
		fts.WithURI(uri),
		fts.WithLogger(logger),
	)

	return indexer, err
}

func loadApps() ([]fts.Attribute[int64, string], error) {
	appsList, err := steam.LoadAppsList()
	if err != nil {
		return nil, err
	}

	return ptr.Cast[[]fts.Attribute[int64, string]](appsList), nil
}
