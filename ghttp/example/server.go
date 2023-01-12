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
	qFn := func(ctx context.Context, q *string) (int, string, *[]*User, map[string]string, error) {
		var u = make([]*User, len(users), len(users))
		for idx, user := range users {
			u[idx] = user
		}

		return 200, "users listed successfully", &u, nil, nil
	}
	return ghttp.Query("ListUsers", nil, qFn)
}

func getUser() http.HandlerFunc {
	qFn := func(ctx context.Context, q *string) (int, string, *User, map[string]string, error) {
		if q == nil || *q == "" {
			return 400, "invalid username", nil, nil, errors.New("no user provided")
		}

		for _, user := range users {
			if *q == user.Username {
				return 200, "user fetched successfully", user, nil, nil
			}
		}
		return 404, "user not found", nil, nil, errors.New("user not found")
	}

	pFn := func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}
		return &q, nil
	}

	return ghttp.Query("GetUser", pFn, qFn)
}

func createUser() http.HandlerFunc {
	qFn := func(ctx context.Context, q *User) (int, string, *User, map[string]string, error) {
		if q == nil {
			return 400, "empty request", nil, nil, errors.New("no user provided")
		}
		if q.Name == "" {
			return 400, "no name provided", nil, nil, errors.New("no name provided")
		}
		if q.Username == "" {
			return 400, "no username provided", nil, nil, errors.New("no username provided")
		}

		for _, user := range users {
			if q.Username == user.Username {
				return 400, "username already taken", nil, nil, errors.New("username unavailable")
			}
		}

		id := len(users)
		users[id] = &User{
			ID:       id,
			Name:     q.Name,
			Username: q.Username,
		}

		q.ID = id

		return 200, "user added successfully", q, nil, nil
	}

	pFn := func(ctx context.Context, r *http.Request) (*User, error) {
		u, err := ghttp.ReadBody[User](ctx, r)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user data from request body: %v", err)
		}

		return u, nil
	}

	return ghttp.Query("AddUser", pFn, qFn)
}

func deleteUser() http.HandlerFunc {
	qFn := func(ctx context.Context, q *string) (int, string, map[string]string, error) {
		if q == nil || *q == "" {
			return 400, "empty request", nil, errors.New("no user provided")
		}

		for id, user := range users {
			if *q == user.Username {
				delete(users, id)
				return 200, "user deleted successfully", nil, nil
			}
		}

		return 404, "user not found", nil, errors.New("user not found")
	}

	pFn := func(ctx context.Context, r *http.Request) (*string, error) {
		prefix := "/users/"
		q := r.URL.Path[len(prefix):]

		if q == "" {
			return nil, errors.New("no username provided")
		}
		return &q, nil
	}

	return ghttp.Exec("DeleteUser", pFn, qFn)
}

func updateUser() http.HandlerFunc {
	qFn := func(ctx context.Context, q *User) (int, string, *User, map[string]string, error) {
		if q == nil {
			return 400, "empty request", nil, nil, errors.New("no user provided")
		}
		if q.Name == "" {
			return 400, "no name provided", nil, nil, errors.New("no name provided")
		}

		for _, user := range users {
			if q.Username == user.Username {
				if q.Name == user.Name {
					return 200, "no changes required", user, nil, nil
				}
				user.Name = q.Name
				return 200, "user updated successfully", user, nil, nil
			}
		}

		return 404, "user not found", nil, nil, errors.New("user not found")
	}

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

	return ghttp.Query("UpdateUser", pFn, qFn)
}

func usersHandler() ghttp.Handler {
	p := "/users/"
	return ghttp.Handler{
		Path: p,
		Fn: func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				if len(r.URL.Path[len(p):]) == 0 {
					listUsers()(w, r)
					return
				}
				getUser()(w, r)
				return
			case http.MethodPost:
				createUser()(w, r)
			case http.MethodDelete:
				deleteUser()(w, r)
			case http.MethodPut:
				updateUser()(w, r)
			default:
				ghttp.ErrResponse(404, "not found", errors.New("path not found"), nil).WriteHTTP(r.Context(), w)
			}
		},
	}
}

func endpoints() ghttp.Endpoints {
	e := ghttp.NewEndpoints()

	e.Set(usersHandler())

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
