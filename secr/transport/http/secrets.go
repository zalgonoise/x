package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/sqlite"
)

func (s *server) secretsGet() http.HandlerFunc {
	type secretsGetRequest struct {
		Username string `json:"username,omitempty"`
		Key      string `json:"key,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*secretsGetRequest, error) {
		splitPath := getPath(r.URL.Path)
		username, key := splitPath[1], splitPath[3]

		return &secretsGetRequest{
			Username: username,
			Key:      key,
		}, nil

	}

	var execFn = func(ctx context.Context, q *secretsGetRequest) *ghttp.Response[secret.Secret] {
		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid username")
		}

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
		username := getPath(r.URL.Path)[1]

		if username == "" {
			return nil, errors.New("no username provided")
		}
		return &username, nil
	}

	var execFn = func(ctx context.Context, q *string) *ghttp.Response[[]*secret.Secret] {
		if q == nil || *q == "" {
			return ghttp.NewResponse[[]*secret.Secret](http.StatusBadRequest, "invalid username")
		}

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
		Username string `json:"username,omitempty"`
		Key      string `json:"key,omitempty"`
		Value    string `json:"value,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*secretsCreateRequest, error) {
		username := getPath(r.URL.Path)[1]
		secr, err := ghttp.ReadBody[secretsCreateRequest](ctx, r)
		if err != nil {
			return nil, err
		}

		secr.Username = username
		return secr, nil
	}

	var execFn = func(ctx context.Context, q *secretsCreateRequest) *ghttp.Response[secret.Secret] {
		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid username")
		}

		err := s.s.CreateSecret(ctx, q.Username, q.Key, []byte(q.Value))
		if err != nil {
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[secret.Secret](http.StatusOK, "secret created successfully")
	}

	return ghttp.Do("SecretsCreate", parseFn, execFn)
}

func (s *server) secretsDelete() http.HandlerFunc {
	type secretsDeleteRequest struct {
		Username string `json:"username,omitempty"`
		Key      string `json:"key,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*secretsDeleteRequest, error) {
		splitPath := getPath(r.URL.Path)
		username, key := splitPath[1], splitPath[3]

		return &secretsDeleteRequest{
			Username: username,
			Key:      key,
		}, nil
	}

	var execFn = func(ctx context.Context, q *secretsDeleteRequest) *ghttp.Response[secret.Secret] {
		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid request")
		}

		err := s.s.DeleteSecret(ctx, q.Username, q.Key)
		if err != nil {
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[secret.Secret](http.StatusOK, "secret deleted successfully")
	}

	return ghttp.Do("SecretsDelete", parseFn, execFn)
}
