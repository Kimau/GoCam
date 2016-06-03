package google

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"GoCamCapture/web"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type ClientSecret struct {
	Id     string `json:"client_id"`
	Secret string `json:"client_secret"`
}

func Login(wf *web.WebFace, clientScopes []string) (*oauth2.Token, error) {
	var Token *oauth2.Token

	// X
	secret, err := loadClientSecret("_secret.json")
	if err != nil {
		log.Fatalln("Secret Missing: %s", err)
		return nil, err
	}

	config := &oauth2.Config{
		ClientID:     secret.Id,
		ClientSecret: secret.Secret,
		Endpoint:     google.Endpoint,
		Scopes:       clientScopes,
	}

	ctx := context.Background()

	{
		var err error
		cacheFile := tokenCacheFile(config)
		Token, err = tokenFromFile(cacheFile)

		if err != nil || (time.Now().After(Token.Expiry)) {
			Token = tokenFromWeb(ctx, config, wf)
			saveToken(cacheFile, Token)
		} else {
			log.Printf("Using cached token")
		}
	}

	c := config.Client(ctx, Token)

	setupClients(c)

	return Token, nil
}

func loadClientSecret(filename string) (*ClientSecret, error) {
	jsonBlob, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cs ClientSecret
	err = json.Unmarshal(jsonBlob, &cs)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func tokenCacheFile(config *oauth2.Config) string {
	hash := fnv.New32a()
	hash.Write([]byte(config.ClientID))
	hash.Write([]byte(config.ClientSecret))
	hash.Write([]byte(strings.Join(config.Scopes, " ")))
	fn := fmt.Sprintf("_cache-tok%v", hash.Sum32())
	return url.QueryEscape(fn)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token, err := DecodeToken(f)
	return token, err
}

func saveToken(file string, token *oauth2.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	EncodeToken(token, f)
}

func EncodeToken(tok *oauth2.Token, dst io.Writer) {
	gob.NewEncoder(dst).Encode(tok)
}

func DecodeToken(src io.Reader) (*oauth2.Token, error) {
	token := new(oauth2.Token)
	err := gob.NewDecoder(src).Decode(token)
	return token, err
}

func tokenFromWeb(ctx context.Context, config *oauth2.Config, wf *web.WebFace) *oauth2.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())

	config.RedirectURL = "http://" + wf.Addr + "/login"

	{ // Auto
		authURL := config.AuthCodeURL(randState)

		wf.RedirectHandler = func(rw http.ResponseWriter, req *http.Request) {

			if req.URL.Path == "/favicon.ico" {
				http.Error(rw, "", 404)
				return
			}

			if !strings.HasPrefix(req.URL.Path, "/login") {
				log.Println("Redirect ", req.URL.Path, strings.HasPrefix(req.URL.Path, "/login"))
				http.Redirect(rw, req, authURL, 302)
				return
			}

			if req.FormValue("state") != randState {
				log.Printf("State doesn't match: req = %#v", req)
				http.Error(rw, "", 500)
				return
			}

			if code := req.FormValue("code"); code != "" {
				wf.RedirectHandler = nil
				http.Redirect(rw, req, "http://"+wf.Addr+"/", 302)
				ch <- code
				return
			}
		}

		log.Println("Awaiting Authorize Token")
	}

	code := <-ch
	log.Printf("Got code: %s", code)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return token
}
