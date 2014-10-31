package auth

import "net/http"

// User is a base type for implementing authentication.
type User interface {
	IsAdmin() bool
	// Logout invalidates the user session. This will take control of the request flow.
	Logout(http.ResponseWriter, *http.Request) error
}
