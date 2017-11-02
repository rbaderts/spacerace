package core

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net"
	"net/http"
	"os"

	"sort"

	"log"
	//	"math"

	//"github.com/gorilla/pat"
	//	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
	//	"github.com/markbates/goth/providers/auth0"
	_ "github.com/markbates/goth/providers/facebook"
	_ "github.com/markbates/goth/providers/github"
	_ "github.com/markbates/goth/providers/gplus"
	_ "github.com/markbates/goth/providers/openidConnect"
	_ "github.com/markbates/goth/providers/paypal"
	_ "github.com/markbates/goth/providers/twitter"
)

func init() {
	//store := sessions.NewFilesystemStore(os.TempDir(), []byte("spacerace-auth"))

	// set the maxLength of the cookies stored on the disk to a larger number to prevent issues with:
	// securecookie: the value is too long
	// when using OpenID Connect , since this can contain a large amount of extra information in the id_token

	// Note, when using the FilesystemStore only the session.ID is written to a browser cookie, so this is explicit for the storage on disk
	//store.MaxLength(math.MaxInt64)

	//gothic.Store = store
}

func SetupAuth(router *mux.Router) {

	oauth_host := os.Getenv("OAUTH_HOST")
	if len(oauth_host) <= 1 {
		oauth_host = "localhost:8080"
	}

	callback := fmt.Sprintf("http://%s", oauth_host)

	fmt.Printf("Oauth callback = %s\n", callback)

	//	ip := GetOutboundIP()
	//	callback := fmt.Sprintf("http://%s:8080/auth/gplus/callback", ip.String())

	goth.UseProviders(

		//facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), "http://localhost:3000/auth/facebook/callback"),
		gplus.New(os.Getenv("GPLUS_KEY"), os.Getenv("GPLUS_SECRET"), callback),
		//github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), "http://localhost:3000/auth/github/callback"),
		//amazon.New(os.Getenv("AMAZON_KEY"), os.Getenv("AMAZON_SECRET"), "http://localhost:3000/auth/amazon/callback"),

		//Pointed localhost.com to http://localhost:3000/auth/yahoo/callback through proxy as yahoo
		// does not allow to put custom ports in redirection uri
		//By default paypal production auth urls will be used, please set PAYPAL_ENV=sandbox as environment variable for testing
		//in sandbox environment
	)

	// OpenID Connect is based on OpenID Connect Auto Discovery URL (https://openid.net/specs/openid-connect-discovery-1_0-17.html)
	// because the OpenID Connect provider initialize it self in the New(), it can return an error which should be handled or ignored
	// ignore the error for now
	//	openidConnect, _ := openidConnect.New(os.Getenv("OPENID_CONNECT_KEY"), os.Getenv("OPENID_CONNECT_SECRET"), "http://localhost:8080/auth/openid-connect/callback", os.Getenv("OPENID_CONNECT_DISCOVERY_URL"))
	//	if openidConnect != nil {
	//		goth.UseProviders(openidConnect)
	//	}

	m := make(map[string]string)
	//m["amazon"] = "Amazon"
	//m["facebook"] = "Facebook"
	//m["github"] = "Github"
	m["gplus"] = "Google Plus"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	//	p := pat.New()
	//router.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {
	router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {

		session, err := Store.Get(r, "session")
		if err != nil {
			fmt.Printf("get session store error = %v\n", err)
		}

		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
			var user *User
			user, err = LoadUserByEmail(DB, gothUser.Email)

			if user == nil {
				user, err = AddProvidedUser(DB, gothUser.Email, gothUser.Provider)
			}

			if err == nil {
				session.Values["loggedin"] = user.Id
				session.Values["loggedin_email"] = user.Email
				session.Save(r, w)
				http.Redirect(w, r, "/lobby", 302)
				return
			}
			//t, _ := template.New("foo").Parse(userTemplate)
			//t.Execute(w, gothUser)
		} else {
			fmt.Printf("error = %v\n", gothUser)
		}
		//		t, _ := template.New("foo").Parse(userTemplate)
		//		t.Execute(w, user)
	}).Methods("GET")

	//p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
	router.HandleFunc("/login/{provider}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("login/{proivder}")

		gothic.Logout(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}).Methods("GET")

	router.HandleFunc("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		// try to get the user without re-authenticating
		fmt.Printf("auth/{proivder}")
		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {

			var user *User
			user, err = LoadUserByEmail(DB, gothUser.Email)

			if user == nil {
				AddProvidedUser(DB, gothUser.Email, gothUser.Provider)
			}

			t, _ := template.New("foo").Parse(userTemplate)
			t.Execute(w, gothUser)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}).Methods("GET")

	//p.Get("/auth", func(res http.ResponseWriter, req *http.Request) {
	router.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {

		fmt.Printf("auth")
		provider, err := gothic.GetProviderName(r)
		if err != nil {
			fmt.Printf("GetProviderNmae err = %v\n", err)
		}

		email := providerIndex.ProvidersMap["Email"]

		var user *User
		user, err = LoadUserByEmail(DB, email)
		if user == nil {
			AddProvidedUser(DB, email, provider)
		}

		http.Redirect(w, r, "/lobby", 302)
		//t, _ := template.New("foo").Parse(indexTemplate)
		//t.Execute(w, providerIndex)
	}).Methods("GET")
	//log.Fatal(http.ListenAndServe(":3000", p))
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
