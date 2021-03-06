package old

import (
	"errors"
	"fmt"
	"github.com/emakryo/adcoter/answer"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"strings"
)

func (sess *Session) Submit(ans answer.Answer) (id string, err error) {
	resp, err := sess.get("/submit")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)
	var tok html.Token
	var session string
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.StartTagToken && tok.Data == "input" &&
			find(tok.Attr, "name") == "__session" {
			session = find(tok.Attr, "value")
		}
	}
	if session == "" {
		err = errors.New(fmt.Sprintf("%s/submit : Parse Failure", sess.Url))
		return
	}

	taskID, err := sess.retrieveTaskID(ans.Id)
	if err != nil {
		return
	}

	data := url.Values{}
	data.Add("__session", session)
	data.Add("task_id", taskID)
	data.Add("language_id_"+taskID, ans.Language)
	data.Add("source_code", ans.Code)

	resp, err = sess.postForm("/submit", data)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.Request.URL.Path != "/submissions/me" {
		err = errors.New("Submission failure")
		return
	}

	id, err = submissionID(resp.Body)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s: Submission ID not retrieved",
			resp.Request.URL))
		return
	}
	return id, nil
}

func find(attrs []html.Attribute, key string) (val string) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (sess *Session) retrieveTaskID(problem string) (id string, err error) {
	resp, err := sess.get("/assignments")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)
	var tok html.Token
	var column int
	var match = false
	var submitURL string
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.StartTagToken && tok.Data == "tr" {
			column = 0
			match = false
		}
		if tok.Type == html.StartTagToken && tok.Data == "td" {
			column += 1
		}
		if tok.Type == html.TextToken && column == 1 && tok.Data == problem {
			match = true
		}
		if tok.Type == html.StartTagToken && column == 5 && tok.Data == "a" && match {
			submitURL = find(tok.Attr, "href")
			break
		}
	}

	if submitURL == "" {
		err = errors.New(fmt.Sprintf("%v: Problem ID not found", problem))
		return
	}

	parsed, err := url.Parse(submitURL)
	if err != nil {
		return
	}
	ids, ok := parsed.Query()["task_id"]
	if !ok {
		err = errors.New(fmt.Sprintf("%s/assignments : Parse error", sess.Url))
		return
	}
	return ids[0], nil
}

func submissionID(body io.Reader) (id string, err error) {
	tokenizer := html.NewTokenizer(body)
	done := false
	tbody := false
	for tokenizer.Next() != html.ErrorToken {
		tok := tokenizer.Token()
		if tok.Type == html.StartTagToken && tok.Data == "tbody" {
			tbody = true
		}
		if tok.Type == html.EndTagToken && tok.Data == "tbody" {
			tbody = false
		}
		if tbody && tok.Type == html.StartTagToken && tok.Data == "tr" {
			if done {
				break
			}
			done = true
		}
		if tbody && tok.Type == html.StartTagToken && tok.Data == "a" {
			link := find(tok.Attr, "href")
			if link != "" && strings.Contains(link, "submissions") {
				splited := strings.Split(link, "/")
				id = splited[len(splited)-1]
				return id, nil
			}
		}
	}

	err = errors.New("Submission ID not found")
	return
}
