package simple

import (
	"net/http"

	"github.com/Logiraptor/butter/auth"
)

// SimpleUser represents a basic email/password combo
// Embed this type to add on to the base fields
type SimpleUser struct {
	Email    string `json:"-"`
	Password []byte `json:"-"`
}

type strategy struct {
	LogoutURL string
	LoginURL  string
}

func (s *strategy) Auth(req *http.Request) (auth.User, error) {
	return nil, nil
}

func (s *strategy) Login(rw http.ResponseWriter, req *http.Request) error {
	return nil
}
