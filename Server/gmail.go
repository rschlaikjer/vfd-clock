package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/gob"
	"fmt"
	gmail "google.golang.org/api/gmail/v1"
	"hash/fnv"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var oauth_config = &oauth.Config{
	ClientId:     "715152452181-5r97br6c5jj2a9v5alh18lnp8is24020.apps.googleusercontent.com",
	ClientSecret: "fcMvtIJOHlAgQA-ePVvVk0n9",
	Scope:        gmail.MailGoogleComScope,
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
}

type message struct {
	size    int64
	gmailID string
	date    string // retrieved from message header
	snippet string
}

func get_num_unread_emails() int {
	oauth_client := getOAuthClient(oauth_config)
	mail_svc, err := gmail.New(oauth_client)
	if err != nil {
		log.Println(err)
		return 0
	}
	pageToken := ""
	unread_mails := 0
	for {
		req := mail_svc.Users.Messages.List("me").Q("in:inbox is:unread")
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			log.Println("Failed to recv messages: %v", err)
			return 0
		}
		unread_mails += len(r.Messages)
		//for _, m := range r.Messages {
		//msg, err := mail_svc.Users.Messages.Get("me", m.Id).Do()
		//if err != nil {
		//	log.Println("Failed to retrieve message: %v", err)
		//	return 0
		//}
		//log.Println(msg.Snippet)
		//}
		if r.NextPageToken == "" {
			break
		}
	}
	return unread_mails
}

func tokenCacheFile(config *oauth.Config) string {
	hash := fnv.New32a()
	hash.Write([]byte(config.ClientId))
	hash.Write([]byte(config.ClientSecret))
	hash.Write([]byte(config.Scope))
	fn := fmt.Sprintf("go-api-demo-tok%v", hash.Sum32())
	return filepath.Join(filepath.Join(os.Getenv("HOME"), ".cache"), url.QueryEscape(fn))
}

func tokenFromFile(file string) (*oauth.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}
func tokenFromWeb(config *oauth.Config) *oauth.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authUrl := config.AuthCodeURL(randState)
	go openUrl(authUrl)
	log.Printf("Authorize this app at: %s", authUrl)
	code := <-ch
	log.Printf("Got code: %s", code)

	t := &oauth.Transport{
		Config:    config,
		Transport: http.DefaultTransport,
	}
	_, err := t.Exchange(code)
	if err != nil {
		log.Println("Token exchange error: %v", err)
		return nil
	}
	return t.Token
}

func openUrl(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.")
}

func saveToken(file string, token *oauth.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

func getOAuthClient(config *oauth.Config) *http.Client {
	cacheFile := tokenCacheFile(config)
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token = tokenFromWeb(config)
		saveToken(cacheFile, token)
	}

	t := &oauth.Transport{
		Token:     token,
		Config:    config,
		Transport: http.DefaultTransport,
	}
	return t.Client()
}
