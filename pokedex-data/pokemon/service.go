package pokemon

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Service struct {
	f io.WriteCloser
	w *csv.Writer
}

func (s Service) Load(ctx context.Context, min, max int) error {
	items := make([][]string, 0, max)

	for i := min; i < max; i++ {
		summary, err := getPokemon(ctx, i)
		if err != nil {
			return err
		}

		items = append(items, []string{strconv.Itoa(summary.ID), summary.Name, summary.Sprite})
	}

	if err := s.w.WriteAll(items); err != nil {
		return err
	}

	s.w.Flush()

	return nil
}

func (s Service) Close() error {
	if err := s.w.Error(); err != nil {
		return err
	}

	return s.f.Close()
}

func NewService(f io.WriteCloser) Service {
	return Service{
		f: f,
		w: csv.NewWriter(f),
	}
}

func getPokemon(ctx context.Context, id int) (Summary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(baseURLFormat, id), http.NoBody)
	if err != nil {
		return Summary{}, err
	}

	res, err := (&http.Client{
		Timeout: defaultTimeout,
	}).Do(req)
	if err != nil {
		return Summary{}, err
	}

	defer res.Body.Close()

	if res.StatusCode > 399 {
		return Summary{}, fmt.Errorf("invalid status code %d", res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return Summary{}, err
	}

	var p Pokemon

	if err = json.Unmarshal(b, &p); err != nil {
		return Summary{}, err
	}

	return Summary{
		ID:     p.Id,
		Name:   strings.Title(p.Name),
		Sprite: fmt.Sprintf(spriteURLFormat, p.Name),
	}, nil
}
