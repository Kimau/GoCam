package google

import (
	"net/http"

	oauth "golang.org/x/oauth2"
	oauthGoogle "google.golang.org/api/oauth2/v2"
)

func GetIdentity(Token *oauth.Token) (*oauthGoogle.Tokeninfo, error) {
	tokenCall := oauthSvc.Tokeninfo()
	tokenCall.AccessToken(Token.AccessToken)
	token, err := tokenCall.Do()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func GetAuth(getUrl string) (resp *http.Response, err error) {
	<-driveThrottle // Rate Limit
	r, e := loginClient.Get(getUrl)

	return r, e
}
