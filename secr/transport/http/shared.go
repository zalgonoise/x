package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/shared"
)

func (s *server) sharesCreate() http.HandlerFunc {
	type sharesCreateRequest struct {
		Owner   string         `json:"-"`
		Key     string         `json:"-"`
		Targets []string       `json:"targets,omitempty"`
		Until   *time.Time     `json:"until,omitempty"`
		For     *time.Duration `json:"for,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*sharesCreateRequest, error) {
		req, err := ghttp.ReadBody[sharesCreateRequest](ctx, r)
		if err != nil {
			return nil, err
		}
		caller, ok := authz.GetCaller(r)

		if !ok {
			return nil, errors.New("invalid username")
		}

		req.Owner = caller
		req.Key = getPath(r.URL.Path)[1]
		return req, nil
	}

	var execFn = func(ctx context.Context, q *sharesCreateRequest) *ghttp.Response[shared.Share] {
		ctx, span := spanner.Start(ctx, "http.CreateShare:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[shared.Share](http.StatusBadRequest, "empty request")
		}
		span.Add(
			attr.String("for_user", q.Owner),
			attr.New("targets", q.Targets),
		)

		var (
			newShare *shared.Share
			err      error
		)

		if q.Until != nil {
			newShare, err = s.s.ShareUntil(ctx, q.Owner, q.Key, *q.Until, q.Targets...)
		} else if q.For != nil {
			newShare, err = s.s.ShareFor(ctx, q.Owner, q.Key, *q.For, q.Targets...)
		} else {
			newShare, err = s.s.CreateShare(ctx, q.Owner, q.Key, q.Targets...)
		}

		if err != nil {
			return ghttp.NewResponse[shared.Share](http.StatusInternalServerError, err.Error())
		}
		return ghttp.NewResponse[shared.Share](http.StatusOK, "secret shared successfully").WithData(newShare)
	}

	return ghttp.Do("SharesCreate", parseFn, execFn)
}

func (s *server) sharesGet() http.HandlerFunc {
	type sharesGetRequest struct {
		Owner string
		Key   string
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*sharesGetRequest, error) {
		req := new(sharesGetRequest)
		caller, ok := authz.GetCaller(r)
		if !ok {
			return nil, errors.New("invalid username")
		}

		req.Owner = caller
		req.Key = getPath(r.URL.Path)[1]
		return req, nil
	}

	var execFn = func(ctx context.Context, q *sharesGetRequest) *ghttp.Response[[]*shared.Share] {
		ctx, span := spanner.Start(ctx, "http.GetShare:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[[]*shared.Share](http.StatusBadRequest, "empty request")
		}
		span.Add(attr.String("for_user", q.Owner))

		shares, err := s.s.GetShare(ctx, q.Owner, q.Key)
		if err != nil {
			return ghttp.NewResponse[[]*shared.Share](http.StatusInternalServerError, err.Error())
		}
		return ghttp.NewResponse[[]*shared.Share](http.StatusOK, "shared secret fetched successfully").WithData(&shares)
	}

	return ghttp.Do("SharesGet", parseFn, execFn)
}

func (s *server) sharesList() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*string, error) {
		if caller, ok := authz.GetCaller(r); ok {
			return &caller, nil
		}
		return nil, errors.New("invalid username")
	}

	var execFn = func(ctx context.Context, q *string) *ghttp.Response[[]*shared.Share] {
		ctx, span := spanner.Start(ctx, "http.ListShares:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[[]*shared.Share](http.StatusBadRequest, "empty request")
		}
		span.Add(attr.String("for_user", *q))

		shares, err := s.s.ListShares(ctx, *q)
		if err != nil {
			return ghttp.NewResponse[[]*shared.Share](http.StatusInternalServerError, err.Error())
		}
		return ghttp.NewResponse[[]*shared.Share](http.StatusOK, "shared secrets listed successfully").WithData(&shares)
	}

	return ghttp.Do("SharesList", parseFn, execFn)
}

func (s *server) sharesDelete() http.HandlerFunc {
	type sharesDeleteRequest struct {
		Owner   string   `json:"-"`
		Key     string   `json:"-"`
		Targets []string `json:"targets,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*sharesDeleteRequest, error) {
		req, err := ghttp.ReadBody[sharesDeleteRequest](ctx, r)
		if err != nil {
			req = new(sharesDeleteRequest)
		}
		caller, ok := authz.GetCaller(r)
		if !ok {
			return nil, errors.New("invalid username")
		}

		req.Owner = caller
		req.Key = getPath(r.URL.Path)[1]
		return req, nil
	}

	var execFn = func(ctx context.Context, q *sharesDeleteRequest) *ghttp.Response[shared.Share] {
		ctx, span := spanner.Start(ctx, "http.DeleteShare:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[shared.Share](http.StatusBadRequest, "empty request")
		}
		span.Add(attr.String("for_user", q.Owner))

		var err error

		if len(q.Targets) > 0 {
			err = s.s.DeleteShare(ctx, q.Owner, q.Key, q.Targets...)
		} else {
			err = s.s.PurgeShares(ctx, q.Owner, q.Key)
		}

		if err != nil {
			return ghttp.NewResponse[shared.Share](http.StatusInternalServerError, err.Error())
		}
		return ghttp.NewResponse[shared.Share](http.StatusOK, "secret share removed successfully")
	}

	return ghttp.Do("SharesDelete", parseFn, execFn)
}
