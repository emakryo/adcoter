package old

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Session struct {
	url    *url.URL
	client *http.Client
}

func New(rawurl string) (sess *Session, err error) {
	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	return &Session{
		url:    parsedUrl,
		client: &http.Client{Jar: jar},
	}, nil
}

func ContestURL(t string, id int) string {
	return fmt.Sprintf("https://%s%03d.contest.atcoder.jp", t, id)
}

func (sess *Session) SetCookies(cookies []*http.Cookie) {
	sess.client.Jar.SetCookies(sess.url, cookies)
}

func (sess *Session) Cookies() []*http.Cookie {
	return sess.client.Jar.Cookies(sess.url)
}

func (sess *Session) get(path string) (*http.Response, error) {
	return sess.client.Get(sess.url.String() + path)
}

func (sess *Session) postForm(path string, values url.Values) (*http.Response, error) {
	return sess.client.PostForm(sess.url.String()+path, values)
}

func (sess *Session) Valid() bool {
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

func (sess *Session) Login(name, password string) (err error) {
	var values = url.Values{}
	values.Add("name", name)
	values.Add("password", password)
	_, err = sess.postForm("/login", values)
	return err
}

func (sess *Session) IsLoggedin() bool {
	resp, err := sess.get("/submit")
	if err != nil {
		return false
	}
	return resp.Request.URL.Path == "/submit"
}
