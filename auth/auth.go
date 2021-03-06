package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jaumecapdevila/chat/config"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")
	if err == http.ErrNoCookie {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		// Some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

// MustAuth returns a new authHandler object
func Must(handler http.Handler) http.Handler {
	// Wrap authHandler around the received handler
	return &authHandler{next: handler}
}

// LoginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
// hardcoded github provider
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	switch action {
	case "login":
		provider, err := gomniauth.Provider("github")
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}
		loginURL, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf(
					"Error when trying to GetBeingAuthURL for %s %s",
					provider,
					err),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", loginURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider("github")
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}
		cred, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when trying to complete auth for %s: %s", provider, err),
				http.StatusInternalServerError,
			)
			return
		}
		user, err := provider.GetUser(cred)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when trying to get user from %s: %s", provider, err),
				http.StatusInternalServerError,
			)
			return
		}
		authCookieValue := objx.New(map[string]interface{}{
			"name": user.Name(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}

func init() {
	config := config.Load()
	gomniauth.SetSecurityKey(signature.RandomKey(64))
	gomniauth.WithProviders(
		github.New(
			config.Oauth.Github.Key,
			config.Oauth.Github.Secret,
			config.Oauth.Github.Callback,
		),
	)
}
