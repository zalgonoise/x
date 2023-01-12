package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zalgonoise/x/ghttp"
)

type User struct {
	ID       int
	Name     string
	Username string
}

var users = map[int]*User{
	0: {ID: 0, Name: "Me", Username: "me"},
	1: {ID: 1, Name: "The other", Username: "the_other"},
	2: {ID: 2, Name: "Someone else", Username: "someone_else"},
}

func listUsers() http.HandlerFunc {
	qFn := func(ctx context.Context, q *string) *ghttp.Response[[]*User] {
		var u = make([]*User, len(users), len(users))
		for idx, user := range users {
			u[idx] = user
		}

		res := ghttp.NewResponse[[]*User](http.StatusOK, "users listed successfully")
		res.Data = &u
		return res
	}

	return ghttp.Query("ListUsers", nil, qFn)
}

func getUser() http.HandlerFunc {
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

		for _, user := range users {
			if *q == user.Username {
				res := ghttp.NewResponse[User](http.StatusOK, "user fetched successfully")
				res.Data = user
				return res
			}
		}
		return ghttp.NewResponse[User](http.StatusNotFound, "user not found")
	}

	return ghttp.Query("GetUser", pFn, qFn)
}

func createUser() http.HandlerFunc {
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

		for _, user := range users {
			if q.Username == user.Username {
				return ghttp.NewResponse[User](http.StatusUnauthorized, "username already taken")
			}
		}

		id := len(users)
		users[id] = &User{
			ID:       id,
			Name:     q.Name,
			Username: q.Username,
		}

		q.ID = id

		res := ghttp.NewResponse[User](http.StatusOK, "user added successfully")
		res.Data = q
		return res
	}

	return ghttp.Query("AddUser", pFn, qFn)
}

func deleteUser() http.HandlerFunc {
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

		for id, user := range users {
			if *q == user.Username {
				delete(users, id)
				return ghttp.NewResponse[string](http.StatusOK, "user deleted successfully")
			}
		}

		return ghttp.NewResponse[string](http.StatusNotFound, "user not found")
	}

	return ghttp.Exec("DeleteUser", pFn, qFn)
}

func updateUser() http.HandlerFunc {
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

		for _, user := range users {
			if q.Username == user.Username {
				if q.Name == user.Name {
					return ghttp.NewResponse[User](http.StatusOK, "no changes required")
				}
				user.Name = q.Name
				res := ghttp.NewResponse[User](http.StatusOK, "user updated successfully")
				res.Data = user
				return res
			}
		}

		return ghttp.NewResponse[User](http.StatusNotFound, "user not found")
	}

	return ghttp.Query("UpdateUser", pFn, qFn)
}

func usersHandler() []ghttp.Handler {
	p := "/users/"

	return []ghttp.Handler{
		{
			Method: http.MethodGet,
			Path:   p,
			Fn: func(w http.ResponseWriter, r *http.Request) {
				if len(r.URL.Path[len(p):]) == 0 {
					listUsers()(w, r)
					return
				}
				getUser()(w, r)
			},
		},
		{
			Method: http.MethodPost,
			Path:   p,
			Fn:     createUser(),
		},
		{
			Method: http.MethodDelete,
			Path:   p,
			Fn:     deleteUser(),
		},
		{
			Method: http.MethodPut,
			Path:   p,
			Fn:     updateUser(),
		},
	}
}

func endpoints() ghttp.Endpoints {
	e := ghttp.NewEndpoints()

	e.Set(usersHandler()...)

	return e
}

func main() {
	e := endpoints()
	s := ghttp.NewServer(e, 8080)

	err := s.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
