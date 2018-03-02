package session

import (
	"net/http"
	"github.com/emakryo/adcoter/status"
	"github.com/emakryo/adcoter/answer"
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
