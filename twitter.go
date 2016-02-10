package main 

import {
	"github.com/ChimeraCoder/anaconda"
}

const (
MINUTES_TO_WAIT = 2
APPROVED_USER = "EvilKimau"
)

type ClientSecret struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`

	AccessToken  string `json:"access_token"`
	AccessSecret string `json:"access_token_secret"`
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

func startTwitterAPI() (*anaconda.TwitterApi, error) {
	secret, err := loadClientSecret("_secret.json")
	if err != nil {
		log.Fatalln("Secret Missing: %s", err)
		return nil, err
	}

	anaconda.SetConsumerKey(secret.Key)
	anaconda.SetConsumerSecret(secret.Secret)
	api := anaconda.NewTwitterApi(secret.AccessToken, secret.AccessSecret)
	return api, nil
}


func postImageTweet(api *anaconda.TwitterApi, gifFile string, t *anaconda.Tweet) error {
	// Post

	data, err := ioutil.ReadFile(gifFile)
	if err != nil {
		return err
	}

	mediaResponse, err := api.UploadMedia(base64.StdEncoding.EncodeToString(data))
	if err != nil {
		return err
	}

	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))
	v.Set("in_reply_to_status_id", t.IdStr)

	tweetString := fmt.Sprintf("@%s here are your fireworks", t.User.ScreenName)

	_, err = api.PostTweet(tweetString, v)
	if err != nil {
		return err
	} else {
		// fmt.Println(result)
	}

	return nil
}

func twitterLoop(chan string) {
	api, _ := startTwitterAPI()
	var startTime = time.Now()

	// Refresh Loop
	var lastId int64 = 0
	var err error
	var hasNewBits bool = true
	var loopMe = true
	for loopMe {
		// Sleep
		time.Sleep(time.Minute * MINUTES_TO_WAIT)

		if hasNewBits {
			fmt.Printf("\nRefreshing")
			hasNewBits = false
		}

		// Get Mentions
		v := url.Values{}
		v.Set("count", "15")
		if lastId != 0 {
			v.Set("since_id", strconv.FormatInt(lastId, 10))
		}

		// Tweets
		var tweets []anaconda.Tweet
		tweets, err = api.GetMentionsTimeline(v)
		if len(tweets) > 0 {
			fmt.Printf("\nRetrieved %d mentions. \n", len(tweets))
			hasNewBits = true
		} else {
			fmt.Printf(".")
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, t := range tweets {
			// Get Last ID
			if lastId < t.Id {
				lastId = t.Id
			}

			ttime, _ := t.CreatedAtTime()
			timeDiff := startTime.Sub(ttime)

			if(t.User.ScreenName != APPROVED_USER) {
				fmt.Printf("%s not approved \n", t.User.ScreenName)
				body := fmt.SPrintf("%s tried to tell me to [%s] \n", t.User.ScreenName, t.Text)
				api.PostDMToScreenName(APPROVED_USER, body)
			}

			if timeDiff > 0 {
				// Old Tweet
				if timeDiff > time.Hour {
					fmt.Printf("Ignoring tweet from %s because its from %.0f hours ago \n", t.User.ScreenName, timeDiff.Hours())
				} else if timeDiff > time.Minute {
					fmt.Printf("Ignoring tweet from %s because its from %.0f minutes ago \n", t.User.ScreenName, timeDiff.Minutes())
				} else {
					fmt.Printf("Ignoring tweet from %s because its from %.0f seconds ago \n", t.User.ScreenName, timeDiff.Seconds())
				}
				continue
			}

			// NEXT Twet
		}

		// Next Round
	}
}