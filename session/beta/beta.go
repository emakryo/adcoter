package beta

import (
	"fmt"
	"github.com/emakryo/adcoter/answer"
	"github.com/emakryo/adcoter/status"
	"github.com/emakryo/adcoter/session"
)

type Session struct {
	*session.SessionBase
}

func New(url string) (*Session, error) {
	sess, err := session.New(url)
	return &Session{sess}, err
}

func ContestURL(t string, id int) string {
	return fmt.Sprintf("https://beta.atcoder.jp/contests/%s%03d", t, id)
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

func (sess *Session) Submit(ans answer.Answer) (id string, err error) {
	return
}

func (sess *Session) Status(string) (stat status.Status, err error) {
	return
}
