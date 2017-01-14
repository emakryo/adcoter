package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgiagio/getpass"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/user"
)

type session struct {
	client *http.Client
	url    string
}

var cacheFile string

func newSession(contestURL string) (sess session, err error) {
	sess.url = contestURL
	if err != nil {
		return
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	sess.client = &http.Client{Jar: jar}
	if !sess.valid() {
		err = errors.New(contestURL + ": Invalid contest")
		return
	}

	usr, err := user.Current()
	if err != nil {
		return
	}
	cacheDir := usr.HomeDir + "/.adcoter"
	os.MkdirAll(cacheDir, 0700)
	cacheFile = usr.HomeDir + "/.adcoter/cookies"

	err = sess.authorize()
	if err != nil {
		err = sess.login()
		if err != nil {
			fatal(err)
		}
	}

	return sess, nil
}

func (sess *session) get(path string) (resp *http.Response, err error) {
	resp, err = sess.client.Get(sess.url + path)
	return
}

func (sess *session) postForm(path string, values url.Values) (resp *http.Response, err error) {
	resp, err = sess.client.PostForm(sess.url+path, values)
	return
}

func (sess *session) valid() bool {
	resp, err := sess.get("")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)

	var tok html.Token
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.TextToken && tok.Data == "404" {
			return false
		}
	}

	return true
}

func (sess *session) loginSuccess() bool {
	resp, err := sess.get("/submit")
	if err != nil {
		return false
	}
	return resp.Request.URL.Path == "/submit"
}

func (sess *session) authorize() (err error) {
	err = sess.loadCookies()
	if err != nil {
		return
	}
	if !sess.loginSuccess() {
		return errors.New("Invalid cookies")
	}
	logger.Println("Login with cached cookies")
	return nil
}

func (sess *session) login() (err error) {
	var name, password string
	fmt.Printf("User ID: ")
	fmt.Scan(&name)
	password, err = getpass.GetPassword("Password: ")
	if err != nil {
		return
	}
	var values = url.Values{}
	values.Add("name", name)
	values.Add("password", password)
	resp, err := sess.postForm("/login", values)
	if err != nil {
		return
	}
	resp.Body.Close()
	if !sess.loginSuccess() {
		return errors.New("login: Invalid user ID or password")
	}
	//log.Printf("Login successful");
	sess.saveCookies()
	return nil
}

func (sess *session) saveCookies() {
	parsed, err := url.Parse(sess.url)
	if err != nil {
		fatal(err)
	}
	cookies := sess.client.Jar.Cookies(parsed)
	marshaled, err := json.Marshal(cookies)
	if err != nil {
		fatal(err)
	}
	err = ioutil.WriteFile(cacheFile, marshaled, 0400)
	if err != nil {
		fatal(err)
	}
}

func (sess *session) loadCookies() (err error) {
	var cookies []*http.Cookie
	marshaled, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		//log.Println(err);
		return
	}
	err = json.Unmarshal(marshaled, &cookies)
	if err != nil {
		//log.Println(err);
		return
	}
	parsed, err := url.Parse(sess.url)
	sess.client.Jar.SetCookies(parsed, cookies)
	return nil
}
