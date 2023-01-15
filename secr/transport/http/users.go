package http

import (
	"context"
	"errors"
	"net/http"

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
	var execFn = func(ctx context.Context, q *user.User) *ghttp.Response[user.User] {
		if q == nil {
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		dbuser, err := s.s.CreateUser(ctx, q.Username, q.Password, q.Name)
		if err != nil {
			if errors.Is(sqlite.ErrAlreadyExistsUser, err) {
				return ghttp.NewResponse[user.User](http.StatusConflict, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.User](http.StatusOK, "user created successfully").WithData(dbuser)
	}

	return ghttp.Do("UsersCreate", ghttp.ReadBody[user.User], execFn)
}

func (s *server) usersUpdate() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*user.User, error) {
		prefix := "/users/"
		username := r.URL.Path[len(prefix):]

		if username == "" {
			return nil, errors.New("no username provided")
		}

		u, err := ghttp.ReadBody[user.User](ctx, r)
		if err != nil {
			return nil, err
		}

		if u.Username != username {
			return nil, errors.New("username mismatch")
		}
		return u, nil
	}

	var execFn = func(ctx context.Context, q *user.User) *ghttp.Response[user.User] {
		if q == nil {
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		err := s.s.UpdateUser(ctx, q.Username, q)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.User](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.User](http.StatusOK, "user updated successfully").WithData(q)
	}

	return ghttp.Do("UsersUpdate", parseFn, execFn)
}

func (s *server) usersDelete() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*user.User, error) {
		prefix := "/users/"
		username := r.URL.Path[len(prefix):]

		if username == "" {
			return nil, errors.New("no username provided")
		}

		u, err := ghttp.ReadBody[user.User](ctx, r)
		if err != nil {
			return nil, err
		}

		if u.Username != username {
			return nil, errors.New("username mismatch")
		}
		return u, nil
	}

	var execFn = func(ctx context.Context, q *user.User) *ghttp.Response[user.User] {
		if q == nil {
			return ghttp.NewResponse[user.User](http.StatusBadRequest, "invalid username")
		}

		err := s.s.DeleteUser(ctx, q.Username)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.User](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.User](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.User](http.StatusOK, "user deleted successfully").WithData(q)
	}

	return ghttp.Do("UsersDelete", parseFn, execFn)
}
