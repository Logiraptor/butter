package auth

import (
	"errors"
	"net/http"
)

var (
	// ErrNotLoggedIn means no user is currently logged in
	ErrNotLoggedIn = errors.New("no user is logged in")
)

// AuthStrategy defines a method of authenticating users.
type AuthStrategy interface {
	// Auth is expected to attempt to validate a session token/header/etc.
	Auth(req *http.Request) (User, error)
	// Login is expected to take control of the request flow
	// until the user is authenticated or authentication fails
	Login(rw http.ResponseWriter, req *http.Request) error
}

var methods = map[string]AuthStrategy{}

// Register adds a strategy to the set of available strategies.
func Register(name string, strategy AuthStrategy) {
	methods[name] = strategy
}

// Auth tries all registered authentication mechanisms and uses the first successful one.
func Auth(req *http.Request) (User, error) {
	for _, strategy := range methods {
		u, err := strategy.Auth(req)
		if err == nil {
			return u, nil
		}
	}
	return nil, errors.New("authentication failed")
}
