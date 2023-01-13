package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zalgonoise/x/ghttp"
)

func (s *server) usersGet() http.HandlerFunc {
	pFn := func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}
		return &q, nil
	}

	qFn := func(ctx context.Context, q *string) *ghttp.Response[User] {
		if q == nil || *q == "" {
			return ghttp.NewResponse[User](http.StatusBadRequest, "invalid username")
		}

		for _, user := range s.users {
			if *q == user.Username {
				return ghttp.NewResponse[User](http.StatusOK, "user fetched successfully").WithData(user)
			}
		}
		return ghttp.NewResponse[User](http.StatusNotFound, "user not found")
	}

	return ghttp.Do("GetUser", pFn, qFn)
}

func (s *server) usersList() http.HandlerFunc {
	qFn := func(ctx context.Context, q *string) *ghttp.Response[[]*User] {
		var u = make([]*User, len(s.users))
		for idx, user := range s.users {
			u[idx] = user
		}

		return ghttp.NewResponse[[]*User](http.StatusOK, "users listed successfully").WithData(&u)
	}

	return ghttp.Do("ListUsers", nil, qFn)
}

func (s *server) usersGetListRoute(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path[len(path):]) == 0 {
			s.usersList()(w, r)
			return
		}
		s.usersGet()(w, r)
	}
}

func (s *server) usersCreate() http.HandlerFunc {
	pFn := func(ctx context.Context, r *http.Request) (*User, error) {
		u, err := ghttp.ReadBody[User](ctx, r)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user data from request body: %v", err)
		}

		return u, nil
	}

	qFn := func(ctx context.Context, q *User) *ghttp.Response[User] {
		if q == nil {
			return ghttp.NewResponse[User](http.StatusBadRequest, "empty request")
		}
		if q.Name == "" {
			return ghttp.NewResponse[User](http.StatusBadRequest, "no name provided")
		}
		if q.Username == "" {
			return ghttp.NewResponse[User](http.StatusBadRequest, "no username provided")
		}

		for _, user := range s.users {
			if q.Username == user.Username {
				return ghttp.NewResponse[User](http.StatusUnauthorized, "username already taken")
			}
		}

		id := len(s.users)
		s.users[id] = &User{
			ID:       id,
			Name:     q.Name,
			Username: q.Username,
		}

		q.ID = id

		return ghttp.NewResponse[User](http.StatusOK, "user added successfully").WithData(q)
	}

	return ghttp.Do("AddUser", pFn, qFn)
}

func (s *server) usersUpdate() http.HandlerFunc {
	pFn := func(ctx context.Context, r *http.Request) (*User, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}

		u, err := ghttp.ReadBody[User](ctx, r)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user data from request body: %v", err)
		}

		if u.Username != "" && q != u.Username {
			return nil, errors.New("invalid target username")
		}
		u.Username = q

		return u, nil
	}

	qFn := func(ctx context.Context, q *User) *ghttp.Response[User] {
		if q == nil {
			return ghttp.NewResponse[User](http.StatusBadRequest, "empty request")
		}
		if q.Name == "" {
			return ghttp.NewResponse[User](http.StatusBadRequest, "no name provided")
		}

		for _, user := range s.users {
			if q.Username == user.Username {
				if q.Name == user.Name {
					return ghttp.NewResponse[User](http.StatusOK, "no changes required")
				}
				user.Name = q.Name
				return ghttp.NewResponse[User](http.StatusOK, "user updated successfully").WithData(user)

			}
		}

		return ghttp.NewResponse[User](http.StatusNotFound, "user not found")
	}

	return ghttp.Do("UpdateUser", pFn, qFn)
}

func (s *server) usersDelete() http.HandlerFunc {
	pFn := func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}
		return &q, nil
	}

	qFn := func(ctx context.Context, q *string) *ghttp.Response[string] {
		if q == nil || *q == "" {
			return ghttp.NewResponse[string](http.StatusBadRequest, "empty request")
		}

		for id, user := range s.users {
			if *q == user.Username {
				delete(s.users, id)
				return ghttp.NewResponse[string](http.StatusOK, "user deleted successfully")
			}
		}

		return ghttp.NewResponse[string](http.StatusNotFound, "user not found")
	}

	return ghttp.Do("DeleteUser", pFn, qFn)
}
