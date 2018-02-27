package old

import (
	"fmt"
)

type Contest struct {
	url string
	sess *session

}

func NewContest(url string) (c *Contest, err error) {
	c = &Contest{
		url: url,
		sess: nil,
	}
	sess, err := newSession(url)
	c.sess = &sess
	return c, err
}

func (c Contest) GetURL() string {
	return c.url
}

func NewContestFromId(t string, id int) (c *Contest, err error) {
	return NewContest(fmt.Sprintf("https://%s%03d.contest.atcoder.jp", t, id))
}
