package hatenaauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/sessions"
	oauth1 "github.com/rmatsuoka/dghubble-oauth1"
)

var (
	sessionKey = os.Getenv("SESSION_KEY")

	consumerKey      = os.Getenv("CONSUMER_KEY")
	consumerSecret   = os.Getenv("CONSUMER_SECRET")
	callbackURL      = os.Getenv("CALLBACK_URL")
	requestTokenPath = "https://www.hatena.com/oauth/initiate"
	authorizePath    = "https://www.hatena.ne.jp/oauth/authorize"
	accessTokenPath  = "https://www.hatena.com/oauth/token"
)

var store = sessions.NewCookieStore([]byte(sessionKey))

type Auth struct {
	config *oauth1.Config
}

func New() *Auth {
	return &Auth{
		config: &oauth1.Config{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			CallbackURL:    callbackURL,
			Endpoint: oauth1.Endpoint{
				RequestTokenURL: requestTokenPath,
				AuthorizeURL:    authorizePath,
				AccessTokenURL:  accessTokenPath,
			},
			Realm:      "",
			Signer:     nil,
			Noncer:     nil,
			HTTPClient: http.DefaultClient,
		},
	}
}

func (a *Auth) Install(handle func(string, http.Handler)) {
	handleFunc := func(pattern string, h http.HandlerFunc) {
		handle(pattern, h)
	}

	handleFunc("GET /signin", a.signin)
	handleFunc("GET /callback", a.callback)
}

func (a *Auth) signin(w http.ResponseWriter, req *http.Request) {
	token, secret, err := a.config.RequestToken(url.Values{"scope": []string{"read_public"}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := store.Get(req, "user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["request_token"] = token
	session.Values["request_secret"] = secret
	if err := session.Save(req, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u, err := a.config.AuthorizationURL(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, u.String(), http.StatusFound)
}

type SessionKey string

const (
	sessionKeyRequestToken    = "request_token"
	sessionKeyRequestSecret   = "request_secret"
	sessionKeyURLName         = "url_name"
	sessionKeyDisplayName     = "display_name"
	sessionKeyProfileImageURL = "profile_image_url"
)

func (a *Auth) callback(w http.ResponseWriter, req *http.Request) {
	_, verifier, err := oauth1.ParseAuthorizationCallback(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, err := store.Get(req, "user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	requestToken, ok := session.Values["request_token"]
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	delete(session.Values, "request_token")

	requestSecret, ok := session.Values["request_secret"]
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	delete(session.Values, "request_secret")

	accessToken, accessSecret, err := a.config.AccessToken(requestToken.(string), requestSecret.(string), verifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println(accessToken, accessSecret)

	client := a.config.Client(req.Context(), oauth1.NewToken(accessToken, accessSecret))

	my, err := getHatenaMy(req.Context(), client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["url_name"] = my.URLName
	session.Values["display_name"] = my.DisplayName
	session.Values["profile_image_url"] = my.ProfileImageURL

	if err := session.Save(req, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

type HatenaMy struct {
	URLName         string `json:"url_name"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

func getHatenaMy(ctx context.Context, client *http.Client) (HatenaMy, error) {
	newReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://n.hatena.com/applications/my.json", http.NoBody)
	if err != nil {
		return HatenaMy{}, fmt.Errorf("auth.getHatenaMy: %w", err)
	}
	res, err := client.Do(newReq)
	if err != nil {
		return HatenaMy{}, fmt.Errorf("auth.getHatenaMy: %w", err)
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return HatenaMy{}, fmt.Errorf("auth.getHatenaMy: %w", err)
	}
	log.Printf("%s\n", buf)
	var my HatenaMy
	err = json.Unmarshal(buf, &my)
	if err != nil {
		return HatenaMy{}, fmt.Errorf("auth.getHatenaMy: %w", err)
	}
	slog.Info("login", "my", my)
	return my, nil
}

type ctxkey int

// userKey is a key of username contained in context. Its associated value type is string.
var userKey = ctxkey(0)

func MyFromContext(ctx context.Context) (HatenaMy, bool) {
	my, ok := ctx.Value(userKey).(HatenaMy)
	return my, ok
}

func AuthHandler(handler http.Handler, fallback func(w http.ResponseWriter, req *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session, err := store.Get(req, "user")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		my := HatenaMy{}
		URLName, ok := session.Values["url_name"]
		if !ok {
			fallback(w, req)
			return
		}
		my.URLName = URLName.(string)
		my.DisplayName = session.Values["display_name"].(string)
		my.ProfileImageURL = session.Values["profile_image_url"].(string)

		ctx := context.WithValue(req.Context(), userKey, my)
		req = req.WithContext(ctx)

		if err := session.Save(req, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		handler.ServeHTTP(w, req)
	})
}
