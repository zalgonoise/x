package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

func (s *server) usersGet() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}
		return &q, nil
	}

	var execFn = func(ctx context.Context, q *string) *ghttp.Response[user.User] {
		if q == nil || *q == "" {
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		dbuser, err := s.s.GetUser(ctx, *q)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.User](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.User](http.StatusOK, "user fetched successfully").WithData(dbuser)
	}

	return ghttp.Do("UsersGet", parseFn, execFn)
}

func (s *server) usersList() http.HandlerFunc {
	var execFn = func(ctx context.Context, q *string) *ghttp.Response[[]*user.User] {
		dbuser, err := s.s.ListUsers(ctx)
		if err != nil {
			return ghttp.NewResponse[[]*user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[[]*user.User](http.StatusOK, "user fetched successfully").WithData(&dbuser)
	}

	return ghttp.Do("UsersList", nil, execFn)
}

func (s *server) usersCreate() http.HandlerFunc {
	type usersCreateRequest struct {
		Name     string `json:"name,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*usersCreateRequest, error) {
		u, err := ghttp.ReadBody[usersCreateRequest](ctx, r)
		if err != nil {
			return nil, err
		}
		return u, nil
	}

	var execFn = func(ctx context.Context, q *usersCreateRequest) *ghttp.Response[user.User] {
		ctx, span := spanner.Start(ctx, "UsersCreate")
		span.Add(attr.New("req", q))

		if q == nil {
			span.Event("empty object error")
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		dbuser, err := s.s.CreateUser(ctx, q.Username, q.Password, q.Name)
		if err != nil {
			span.Event("operation error", attr.String("error", err.Error()))

			if errors.Is(sqlite.ErrAlreadyExistsUser, err) {
				return ghttp.NewResponse[user.User](http.StatusConflict, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		span.Event("operation successful", attr.New("user", dbuser))
		return ghttp.NewResponse[user.User](http.StatusOK, "user created successfully").WithData(dbuser)
	}

	return ghttp.Do("UsersCreate", parseFn, execFn)
}

func (s *server) usersUpdate() http.HandlerFunc {
	type usersUpdateRequest struct {
		Name     string `json:"name,omitempty"`
		Username string `json:"username,omitempty"`
	}

	var parseFn = func(ctx context.Context, r *http.Request) (*usersUpdateRequest, error) {
		prefix := "/users/"
		username := r.URL.Path[len(prefix):]

		if username == "" {
			return nil, errors.New("no username provided")
		}

		u, err := ghttp.ReadBody[usersUpdateRequest](ctx, r)
		if err != nil {
			return nil, err
		}

		u.Username = username
		return u, nil
	}

	var execFn = func(ctx context.Context, q *usersUpdateRequest) *ghttp.Response[user.User] {
		if q == nil {
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		u, err := s.s.GetUser(ctx, q.Username)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.User](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		u.Name = q.Name
		err = s.s.UpdateUser(ctx, q.Username, u)
		if err != nil {
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.User](http.StatusOK, "user updated successfully").WithData(u)
	}

	return ghttp.Do("UsersUpdate", parseFn, execFn)
}

func (s *server) usersDelete() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		username := r.URL.Path[len(prefix):]

		if username == "" {
			return nil, errors.New("no username provided")
		}

		return &username, nil
	}

	var execFn = func(ctx context.Context, q *string) *ghttp.Response[user.User] {
		if q == nil || *q == "" {
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		err := s.s.DeleteUser(ctx, *q)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.User](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.User](http.StatusOK, "user deleted successfully")
	}

	return ghttp.Do("UsersDelete", parseFn, execFn)
}
