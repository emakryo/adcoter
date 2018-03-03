package old

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"github.com/emakryo/adcoter/session"
)

type Session struct {
	*session.SessionBase
}

func New(rawurl string) (*Session, error) {
	sess, err := session.New(rawurl)
	return &Session{sess}, err
}

func ContestURL(t string, id int) string {
	return fmt.Sprintf("https://%s%03d.contest.atcoder.jp", t, id)
}

func (sess *Session) get(path string) (*http.Response, error) {
	return sess.Client.Get(sess.Url.String() + path)
}

func (sess *Session) postForm(path string, values url.Values) (*http.Response, error) {
	return sess.Client.PostForm(sess.Url.String()+path, values)
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
