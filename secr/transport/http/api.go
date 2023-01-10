package http

import "net/http"

type API interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	ChangePassword(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)

	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	ListUsers(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)

	CreateSecret(w http.ResponseWriter, r *http.Request)
	GetSecret(w http.ResponseWriter, r *http.Request)
	ListSecrets(w http.ResponseWriter, r *http.Request)
	DeleteSecret(w http.ResponseWriter, r *http.Request)
}
