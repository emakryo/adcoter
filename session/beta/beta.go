package beta

import (
	"net/url"
	"net/http"
//	"net/http/cookiejar"
	"github.com/emakryo/adcoter/answer"
	"github.com/emakryo/adcoter/status"
)

type Session struct {
	client *http.Client
	url *url.URL
}

func New(url string) (sess *Session, err error) {
	return
}

func ContestURL(t string, id int) string {
	return ""
}

func (sess *Session) SetCookies(cookies []*http.Cookie) {
	sess.client.Jar.SetCookies(sess.url, cookies)
}

func (sess *Session) Cookies() []*http.Cookie{
	return sess.client.Jar.Cookies(sess.url)
}

func (sess *Session) Submit(ans answer.Answer) (id string, err error) {
	return
}

func (sess *Session) Status(string) (stat status.Status, err error) {
	return
}

func (sess *Session) Valid() bool {
	return false
}

func (sess *Session) Login(user string, password string) error {
	return nil
}

func (sess *Session) IsLoggedin() bool {
	return false
}
