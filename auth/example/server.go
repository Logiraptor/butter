package example

import (
	"fmt"
	"net/http"

	"github.com/Logiraptor/butter/auth"
	gauth "github.com/Logiraptor/butter/auth/gae"
)

func init() {
	googleAuth := gauth.NewStrategy("/login", "/home")
	auth.Register("GOOGLE", googleAuth)

	http.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(rw, "You logged in.")
	})
	http.HandleFunc("/logout", func(rw http.ResponseWriter, req *http.Request) {
		u, err := auth.Auth(req)
		if err != nil {
			fmt.Fprintln(rw, err.Error())
			return
		}

		err = u.Logout(rw, req)
		if err != nil {
			fmt.Fprintln(rw, err.Error())
			return
		}
	})
	http.HandleFunc("/home", func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(rw, "You logged out.")
	})
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		err := googleAuth.Login(rw, req)
		if err != nil {
			fmt.Fprintln(rw, err.Error())
		}
	})
}
