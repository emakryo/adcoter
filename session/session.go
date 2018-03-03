package session

import (
	"github.com/emakryo/adcoter/answer"
	"github.com/emakryo/adcoter/status"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Session interface {
	Submit(answer.Answer) (string, error)
	Status(string) (status.Status, error)
	Valid() bool
	Login(string, string) error
	IsLoggedin() bool
	SetCookies([]*http.Cookie)
	Cookies() []*http.Cookie
}

type SessionBase struct {
	Url    *url.URL
	Client *http.Client
}

func New(rawurl string) (sess *SessionBase, err error) {
	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	return &SessionBase{
		Url:    parsedUrl,
		Client: &http.Client{Jar: jar},
	}, nil
}

func (sess *SessionBase) SetCookies(cookies []*http.Cookie) {
	sess.Client.Jar.SetCookies(sess.Url, cookies)
}

func (sess *SessionBase) Cookies() []*http.Cookie {
	return sess.Client.Jar.Cookies(sess.Url)
}
