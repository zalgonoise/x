package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

func (s server) secretsGet() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*secret.WithOwner, error) {
		splitPath := strings.Split(r.URL.Path, ",")
		username, key := splitPath[1], splitPath[3]

		return &secret.WithOwner{
			User: user.User{
				Username: username,
			},
			Secret: secret.Secret{
				Key: key,
			},
		}, nil
	}

	var execFn = func(ctx context.Context, q *secret.WithOwner) *ghttp.Response[secret.Secret] {
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
func (s server) secretsList() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}
		return &q, nil
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
func (s server) secretsCreate() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*secret.WithOwner, error) {
		username := strings.Split(r.URL.Path, ",")[1]
		secr, err := ghttp.ReadBody[secret.Secret](ctx, r)
		if err != nil {
			return nil, err
		}

		return &secret.WithOwner{
			User: user.User{
				Username: username,
			},
			Secret: *secr,
		}, nil
	}

	var execFn = func(ctx context.Context, q *secret.WithOwner) *ghttp.Response[secret.Secret] {
		if q == nil {
			return ghttp.NewResponse[secret.Secret](http.StatusBadRequest, "invalid username")
		}

		err := s.s.CreateSecret(ctx, q.Username, q.Key, q.Value)
		if err != nil {
			return ghttp.NewResponse[secret.Secret](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[secret.Secret](http.StatusOK, "secret created successfully")
	}

	return ghttp.Do("SecretsCreate", parseFn, execFn)
}

func (s server) secretsDelete() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*secret.WithOwner, error) {
		splitPath := strings.Split(r.URL.Path, ",")
		username, key := splitPath[1], splitPath[3]

		return &secret.WithOwner{
			User: user.User{
				Username: username,
			},
			Secret: secret.Secret{
				Key: key,
			},
		}, nil
	}

	var execFn = func(ctx context.Context, q *secret.WithOwner) *ghttp.Response[secret.Secret] {
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
