package handlers

import (
	"log"
	"net/http"

	"github.com/oktalz/present-poc/hash"
)

func APILogin(userPwd, adminPwd string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		var err error

		pass, err = hash.Hash(pass)
		if err != nil {
			log.Println(err)
			return
		}

		passwordOK := hash.Equal(pass, userPwd) || hash.Equal(pass, adminPwd)
		if !passwordOK {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		log.Println("/api/login", user, "OK")

		cookieSet := http.Cookie{
			Name:  "present",
			Value: pass,
			Path:  "/",
		}
		http.SetCookie(w, &cookieSet)
	})
}
