package core

import (
	"crypto/rand"
	"encoding/base64"
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	_	"github.com/gocraft/dbr/v2"
	"github.com/rbaderts/spacerace/auth"
	"net/http"
	"net/url"
	"os"
)

type handler struct{}


/*
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
 */


//func LoginHandler(c *gin.Context) {
func LoginHandler(w http.ResponseWriter, r *http.Request) {

		// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("LoginHandler\n")

	//data := []byte("string of data")
	encodedData := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(encodedData, b)
	state := string(encodedData)

	//state := base64.StdEncoding.EncodeToString(b)
	//encodedData := &bytes.Buffer{}
	//encoder := base64.NewEncoder(base64.StdEncoding, encodedData)
	//defer encoder.Close()
	//encoder.Write(data)


	session, err := auth.AuthStore.Get(r, "auth-session")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	err = session.Save(r, w)
	fmt.Printf("LoginHandler 1\n")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authenticator, err := auth.NewAuthenticator()
	fmt.Printf("err == %v\n", err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("LoginHandler 3\n")
	http.Redirect(w, r, authenticator.Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}


func LogoutHandler(w http.ResponseWriter, r *http.Request) {
//func LogoutHandler( c *gin.Context) {

	domain := os.Getenv("SPACERACE_DOMAIN")

	logoutUrl, err := url.Parse("https://" + domain)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logoutUrl.Path += "/v2/logout"
	parameters := url.Values{}

	var scheme string
	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" +  r.Host + "/login")
	fmt.Printf("returnTo: %v\n", returnTo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("SPACERACE_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
}

func AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
///func AuthCallbackHandler(c *gin.Context) {

	fmt.Printf("AuthCallbackHandler 1\n");
	session, err := auth.AuthStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("session.Value['state'] = %v\n", session.Values["state"])

	if (r.URL.Query().Get("state") != session.Values["state"]) {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	authenticator, err := auth.NewAuthenticator()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("r.URL.Query() = %v\n", r.URL.Query().Get("code"))
	token, err := authenticator.Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
	//token, err := authenticator.Config.Exchange(authenticator.Ctx, r.URL.Query().Get("code"))
	if err != nil {
		fmt.Printf("no token found: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: os.Getenv("SPACERACE_CLIENT_ID"),
	}

	idToken, err := authenticator.Provider.Verifier(oidcConfig).Verify(context.TODO(), rawIDToken)

	//var verifier = authenticator.Provider.Verifier(&oidc.Config{ClientID: v})

    //dToken, err := verifier.Verify(context.TODO(), rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: " + err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract custom claims
	/*
	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		// handle error
	}
	email := claims.Email
	*/

	//	idToken, err := authenticator.Provider.Verifier(oidcConfig).Verify(context.TODO(), rawIDToken)

	// Getting now the userInfo
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("other claims: %v\n", profile)
	email := profile["email"].(string)
	issuer := profile["iss"].(string)
	subject := profile["sub"].(string)
	given_name := profile["given_name"].(string)
	/*
	subject := idToken.Subject
	issuer := idToken.Issuer
	fmt.Printf("subject: %s, issuer %s\n", subject, issuer)
	 */

	//var user *User

	fmt.Printf("email: %v, issuer: %v, subject: %v\n", email, issuer, subject)

	dbSession := DB.NewSession(nil)
	var user *User
	user, err = LoadUserByEmail(dbSession, email)

	fmt.Printf("user = %v\n", user)

	if err != nil || user == nil {
		//user = &User{Subject: subject, Name: profile["name"].(string), Provider: issuer}
		user, err = AddProvidedUser(dbSession, email, issuer, subject)
//		db *dbr.Session, email string, provider string, subject string

	}

	session.Values["id_token"] = rawIDToken
	session.Values["given_name"] = given_name
//	session.Values["user_id"] = user.Id


	session.Values["access_token"] = token.AccessToken
	session.Values["profile"] = profile
	session.Values["uid"] = user.Id
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to logged in page
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
