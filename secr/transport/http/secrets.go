package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/sqlite"
)

func (s *server) secretsGet() http.HandlerFunc {
	type secretsGetRequest struct {
		Username string `json:"-"`
		Key      string `json:"-"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*secretsGetRequest, error) {
		splitPath := getPath(r.URL.Path)
		key := splitPath[1]

		if u, ok := authz.GetCaller(r); ok {
			return &secretsGetRequest{
				Username: u,
				Key:      key,
			}, nil
		}
		return nil, authz.ErrInvalidUser
	}

	var execFn = func(ctx context.Context, q *secretsGetRequest) *ghttp.Response[secret.Secret] {
		ctx, span := spanner.Start(ctx, "http.GetSecret:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid request")
		}
		span.Add(attr.String("for_user", q.Username))

		dbsecr, err := s.s.GetSecret(ctx, q.Username, q.Key)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundSecret, err) {
				return ghttp.NewResponse[secret.Secret](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[secret.Secret](http.StatusOK, "secret fetched successfully").WithData(dbsecr)
	}

	return ghttp.Do("SecretsGet", parseFn, execFn)
}

func (s *server) secretsList() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*string, error) {
		if u, ok := authz.GetCaller(r); ok {
			return &u, nil
		}
		return nil, authz.ErrInvalidUser
	}

	var execFn = func(ctx context.Context, q *string) *ghttp.Response[[]*secret.Secret] {
		ctx, span := spanner.Start(ctx, "http.ListSecrets:exec")
		defer span.End()

		if q == nil || *q == "" {
			return ghttp.NewResponse[[]*secret.Secret](http.StatusBadRequest, "invalid username")
		}
		span.Add(attr.String("for_user", *q))

		dbsecr, err := s.s.ListSecrets(ctx, *q)
		if err != nil {
			return ghttp.NewResponse[[]*secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[[]*secret.Secret](http.StatusOK, "secrets listed successfully").WithData(&dbsecr)
	}

	return ghttp.Do("SecretsList", parseFn, execFn)
}

func (s *server) secretsCreate() http.HandlerFunc {
	type secretsCreateRequest struct {
		Username string `json:"-"`
		Key      string `json:"key,omitempty"`
		Value    string `json:"value,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*secretsCreateRequest, error) {
		secr, err := ghttp.ReadBody[secretsCreateRequest](ctx, r)
		if err != nil {
			return nil, err
		}
		if u, ok := authz.GetCaller(r); ok {
			secr.Username = u
			return secr, nil
		}
		return nil, authz.ErrInvalidUser
	}

	var execFn = func(ctx context.Context, q *secretsCreateRequest) *ghttp.Response[secret.Secret] {
		ctx, span := spanner.Start(ctx, "http.CreateSecret:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid request")
		}
		span.Add(attr.String("for_user", q.Username))

		err := s.s.CreateSecret(ctx, q.Username, q.Key, []byte(q.Value))
		if err != nil {
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		secr, err := s.s.GetSecret(ctx, q.Username, q.Key)
		if err != nil {
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[secret.Secret](http.StatusOK, "secret created successfully").WithData(secr)
	}

	return ghttp.Do("SecretsCreate", parseFn, execFn)
}

func (s *server) secretsDelete() http.HandlerFunc {
	type secretsDeleteRequest struct {
		Username string `json:"-"`
		Key      string `json:"-"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*secretsDeleteRequest, error) {
		splitPath := getPath(r.URL.Path)
		key := splitPath[1]

		if u, ok := authz.GetCaller(r); ok {
			return &secretsDeleteRequest{
				Username: u,
				Key:      key,
			}, nil
		}
		return nil, authz.ErrInvalidUser
	}

	var execFn = func(ctx context.Context, q *secretsDeleteRequest) *ghttp.Response[secret.Secret] {
		ctx, span := spanner.Start(ctx, "http.DeleteSecret:exec")
		defer span.End()

		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid request")
		}
		span.Add(attr.String("for_user", q.Username))

		err := s.s.DeleteSecret(ctx, q.Username, q.Key)
		if err != nil {
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[secret.Secret](http.StatusOK, "secret deleted successfully")
	}

	return ghttp.Do("SecretsDelete", parseFn, execFn)
}
