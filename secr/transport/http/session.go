package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

func (s *server) login() http.HandlerFunc {
	type loginRequest struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}

	var parseFn = func(ctx context.Context, r *http.Request) (*loginRequest, error) {
		u, err := ghttp.ReadBody[loginRequest](ctx, r)
		if err != nil {
			return nil, err
		}

		if u.Password == "" {
			token := r.Header.Get("Authorization")
			if token != "" {
				t := strings.TrimPrefix(token, "Bearer: ")
				if t != "" {
					u.Password = t
					return u, nil
				}
			}
		}
		return nil, errors.New("failed to read password or token in request")
	}

	var execFn = func(ctx context.Context, q *loginRequest) *ghttp.Response[user.Session] {
		if q == nil {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		dbsession, err := s.s.Login(ctx, q.Username, q.Password)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			if errors.Is(sqlite.ErrIncorrectPassword, err) {
				return ghttp.NewResponse[user.Session](http.StatusBadRequest, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "user logged in successfully").WithData(dbsession)
	}

	return ghttp.Do("Login", parseFn, execFn)
}
func (s *server) logout() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*user.User, error) {
		token := r.Header.Get("Authorization")
		if token != "" {
			t := strings.TrimPrefix(token, "Bearer: ")
			if t != "" {
				u, err := s.s.ParseToken(ctx, t)
				if err != nil {
					return u, nil
				}
			}
		}
		return nil, errors.New("failed to read password or token in request")
	}

	var execFn = func(ctx context.Context, q *user.User) *ghttp.Response[user.Session] {
		if q == nil || q.Username == "" {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		err := s.s.Logout(ctx, q.Username)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "user logged out successfully")
	}

	return ghttp.Do("Logout", parseFn, execFn)
}

func (s *server) changePassword() http.HandlerFunc {
	var execFn = func(ctx context.Context, q *user.NewPassword) *ghttp.Response[user.Session] {
		if q == nil {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		err := s.s.ChangePassword(ctx, q.User.Username, q.User.Password, q.Password)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			if errors.Is(sqlite.ErrIncorrectPassword, err) {
				return ghttp.NewResponse[user.Session](http.StatusBadRequest, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "password changed successfully")
	}

	return ghttp.Do("ChangePassword", ghttp.ReadBody[user.NewPassword], execFn)
}

func (s *server) refresh() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*user.Session, error) {
		s, err := ghttp.ReadBody[user.Session](ctx, r)
		if err != nil {
			return nil, err
		}

		if s.Token == "" {
			token := r.Header.Get("Authorization")
			if token != "" {
				t := strings.TrimPrefix(token, "Bearer: ")
				if t != "" {
					s.Token = t
					return s, nil
				}
			}
		}
		return nil, errors.New("failed to read user or token in request")
	}

	var execFn = func(ctx context.Context, q *user.Session) *ghttp.Response[user.Session] {
		if q == nil {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		token, err := s.s.Refresh(ctx, q.Username, q.Token)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			if errors.Is(sqlite.ErrIncorrectPassword, err) {
				return ghttp.NewResponse[user.Session](http.StatusBadRequest, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "token refreshed successfully").WithData(token)
	}

	return ghttp.Do("ChangePassword", parseFn, execFn)

}
