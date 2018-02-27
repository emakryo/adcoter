package old

import (
	"fmt"
	"github.com/emakryo/adcoter/contest"
)

type Contest struct {
	*contest.Base
}

func NewContestFromId(t string, id int) *Contest {
	return &Contest{contest.NewContest(
		fmt.Sprintf("https://%s%03d.contest.atcoder.jp", t, id),
	)}
}

func (c *Contest) Login() {
}

func (c *Contest) Submit(answer contest.Answer) (id string, err error) {
	return
}

func (c *Contest) Status(id string) (s contest.Status, err error) {
	return
}
