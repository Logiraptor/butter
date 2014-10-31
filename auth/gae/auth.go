package gae

import (
	"net/http"

	"github.com/Logiraptor/butter/auth"

	"appengine"
	"appengine/user"
)

// GAEUser wraps appengine's user type
type GAEUser struct {
	user      *user.User
	logoutURL string
}

// IsAdmin returns true if g is an appengine administrator
func (g *GAEUser) IsAdmin() bool {
	return g.user.Admin
}

// Logout invalidates the session established earlier
func (g *GAEUser) Logout(rw http.ResponseWriter, req *http.Request) error {
	c := appengine.NewContext(req)
	url, err := user.LogoutURL(c, g.logoutURL)
	if err != nil {
		return err
	}
	http.Redirect(rw, req, url, 302)
	return nil
}

// NewStrategy returns an auth strategy which manages google accounts.
// Upon logging in, the user is redirected to login,
// Upon logging out, the user is redirected to logout.
func NewStrategy(loginURL, logoutURL string) auth.AuthStrategy {
	return &gaeStrategy{
		LogoutURL: logoutURL,
		LoginURL:  loginURL,
	}
}

// gaeStrategy implements an AuthStrategy for google accounts on appengine.
type gaeStrategy struct {
	LogoutURL string
	LoginURL  string
}

// Auth is expected to attempt to validate a session token/header/etc.
func (g *gaeStrategy) Auth(req *http.Request) (auth.User, error) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	if u == nil {
		return nil, auth.ErrNotLoggedIn
	}
	return &GAEUser{
		user:      u,
		logoutURL: g.LogoutURL,
	}, nil
}

// Login is expected to take control of the request flow
// until the user is authenticated or authentication fails
func (g *gaeStrategy) Login(rw http.ResponseWriter, req *http.Request) error {
	c := appengine.NewContext(req)
	url, err := user.LoginURL(c, g.LoginURL)
	if err != nil {
		return err
	}
	http.Redirect(rw, req, url, 302)
	return nil
}
