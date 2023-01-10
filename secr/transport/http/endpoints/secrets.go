package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/user"
)

func (e endpoints) CreateSecret(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.CreateSecret")
	defer s.End()

	// TODO: get user from token
	secr, err := readBody[secret.WithOwner](ctx, r)
	if err != nil {
		res := NewResponse[secret.Secret](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	dbuser, err := e.s.GetUser(ctx, secr.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to fetch user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	err = e.s.CreateSecret(ctx, dbuser.Username, secr.Key, secr.Value)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[secret.Secret](500, "failed to create secret", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse[secret.Secret](200, "secret created successfully", nil, nil)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) GetSecret(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.GetSecret")
	defer s.End()

	// TODO: get user from token
	secr, err := readBody[secret.WithOwner](ctx, r)
	if err != nil {
		res := NewResponse[secret.Secret](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	dbuser, err := e.s.GetUser(ctx, secr.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to fetch user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	dbsecr, err := e.s.GetSecret(ctx, dbuser.Username, secr.Key)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[secret.Secret](500, "failed to fetch secret", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "secret fetched successfully", nil, dbsecr)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) ListSecrets(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.ListSecrets")
	defer s.End()

	// TODO: get user from token
	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}
	dbuser, err := e.s.GetUser(ctx, u.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to fetch user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	secrets, err := e.s.ListSecrets(ctx, dbuser.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[secret.Secret](500, "failed to list secrets", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "secrets listed successfully", nil, &secrets)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.DeleteSecret")
	defer s.End()

	// TODO: get user from token
	secr, err := readBody[secret.WithOwner](ctx, r)
	if err != nil {
		res := NewResponse[secret.WithOwner](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}
	dbuser, err := e.s.GetUser(ctx, secr.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to fetch user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	err = e.s.DeleteSecret(ctx, dbuser.Username, secr.Key)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[secret.Secret](500, "failed to delete secret", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse[secret.Secret](200, "secret deleted successfully", nil, nil)
	res.WriteHTTP(ctx, w)
}
