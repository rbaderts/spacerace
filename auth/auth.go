package auth

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)
var (
	SessionKey = "SpaceRace"
//	AuthStore sessions.CookieStore
	//AuthStore  = sessions.NewCookieStore([]byte(SessionKey))
	AuthStore *sessions.FilesystemStore

)

type Authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

func init() {
	AuthStore = sessions.NewFilesystemStore("spacerace_store", []byte(SessionKey))
	gob.Register(map[string]interface{}{})

}

func NewAuthenticator() (*Authenticator, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, os.Getenv("SPACERACE_DOMAIN"))
	if err != nil {
		log.Printf("failed to get provider: %v", err)
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     os.Getenv("SPACERACE_CLIENT_ID"),
		ClientSecret: os.Getenv("SPACERACE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint: 	  provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
		Ctx:      ctx,
	}, nil
}


func AuthenticationRequired(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("AuthenticationRequired\n")
		//func AuthenticationRequired(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		session, err := AuthStore.Get(r, "auth-session")
		if err != nil {
			fmt.Printf("err = %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values["profile"]; !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)

		} else {
			fmt.Printf("Authenticated\n")

			uid := session.Values["uid"]
			fmt.Printf("uid in session %d\n", uid)
			ctx := context.WithValue(r.Context(), "uid", uid)
			h.ServeHTTP(w, r.WithContext(ctx))

		}

	})
}


/*
func AuthenticationRequired(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("AuthenticationRequired\n")
		//func AuthenticationRequired(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		session, err := AuthStore.Get(r, "auth-session")
		if err != nil {
			fmt.Printf("err = %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("f3")
		if _, ok := session.Values["profile"]; !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			//v := session.Values["uid"]
			//c.Set("uid", v)

			fmt.Printf("f4")
			fmt.Printf("Authenticated\n")
			h.ServeHTTP(w, r)
		}

		fmt.Printf("f5")
	}
	return http.HandlerFunc(fn)

}

/*
func AuthenticationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("AuthenticationRequired\n")
		//func AuthenticationRequired(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		session, err := AuthStore.Get(c.Request, "auth-session")
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values["profile"]; !ok {
			http.Redirect(c.Writer, c.Request, "/", http.StatusSeeOther)
		} else {

			v := session.Values["uid"]
		    c.Set("uid", v)

			fmt.Printf("Authenticated\n")
			c.Next()
		}
	}
}
 */

func containsString(strings []string, checkFor string) bool {
	for _, s := range strings {
		if (s == checkFor) {
			return true
		}
	}
	return false

}