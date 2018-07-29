package beta

import (
	"errors"
	"fmt"
	"github.com/emakryo/adcoter/session"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
)

var Logger *log.Logger

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
	resp, err := sess.Client.Get(sess.Url.String())
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	} else {
		return true
	}
}

func (sess *Session) Login(user string, password string) error {
	loginURL := "https://beta.atcoder.jp/login"
	resp, err := sess.Client.Get(loginURL)
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	form := getForm(doc, "/login")
	if form == nil {
		return errors.New("Input form not found")
	}
	tok := getCSRFToken(form)
	if tok == "" {
		return errors.New("Login token not found")
	}

	values := url.Values{}
	values.Add("username", user)
	values.Add("password", password)
	values.Add("csrf_token", tok)
	Logger.Println("Post form to", loginURL, ":", values)
	resp, err = sess.Client.PostForm(loginURL, values)
	Logger.Println("Response by POST to", loginURL, ":", resp)

	if err != nil {
		return err
	}

	return nil
}

func getForm(node *html.Node, action string) *html.Node {
	if node.Type == html.ElementNode && node.Data == "form" {
		for _, a := range node.Attr {
			if a.Key == "action" && a.Val == action {
				return node
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		m := getForm(c, action)
		if m != nil {
			return m
		}
	}
	return nil
}

func getCSRFToken(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "input" {
		var inputType, inputName, inputValue string
		for _, a := range node.Attr {
			if a.Key == "type" {
				inputType = a.Val
			}
			if a.Key == "name" {
				inputName = a.Val
			}
			if a.Key == "value" {
				inputValue = a.Val
			}
		}
		if inputType == "hidden" && inputName == "csrf_token" {
			return inputValue
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		tok := getCSRFToken(c)
		if tok != "" {
			return tok
		}
	}
	return ""
}

func (sess *Session) IsLoggedin() bool {
	submitURL := sess.Url.String() + "/submit"
	resp, err := sess.Client.Get(submitURL)
	Logger.Println("Response to", submitURL, ":", resp)
	if err != nil {
		Logger.Println(err)
		return false
	}
	if resp.StatusCode == http.StatusOK && resp.Request.URL.String() == submitURL {
		Logger.Println("Logged in")
		return true
	}
	Logger.Println("Not logged in because", resp.StatusCode, "is not", http.StatusOK, "or",
		resp.Request.URL.String(), "is not", submitURL)
	return false
}
