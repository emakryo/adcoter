package session

import (
	"github.com/emakryo/adcoter/answer"
	"github.com/emakryo/adcoter/status"
	"net/http"
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
