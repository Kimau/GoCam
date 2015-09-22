package google

import (
	"log"
	"net/http"
	"time"

	drive "google.golang.org/api/drive/v2"
	oauth "google.golang.org/api/oauth2/v2"
)

var (
	loginClient   *http.Client
	oauthSvc      *oauth.Service
	drvSvc        *drive.Service
	driveThrottle <-chan time.Time
)

func init() {
	rate := time.Second / 10
	driveThrottle = time.Tick(rate)
}

func GetClientScope() []string {
	return []string{
		drive.DriveScope,
		oauth.PlusMeScope,
		oauth.UserinfoEmailScope}
}

func setupClients(client *http.Client) {
	var err error

	loginClient = client

	oauthSvc, err = oauth.New(client)
	if err != nil {
		log.Fatalf("Unable to create OAuth service: %v", err)
	}

	drvSvc, err = drive.New(client)
	if err != nil {
		log.Fatalf("Unable to create Drive service: %v", err)
	}
}
