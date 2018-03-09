package beta

import (
	//"os"
	"errors"
	"github.com/emakryo/adcoter/answer"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

//html.Render(os.Stdout, firstRow)
func (sess *Session) Submit(ans answer.Answer) (id string, err error) {
	submitURL := sess.Url.String() + "/submit"
	resp, err := sess.Client.Get(submitURL)
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	form := getForm(doc, sess.Url.Path+"/submit")
	if form == nil {
		return "", errors.New("No form in /submit")
	}

	tok := getCSRFToken(form)
	if tok == "" {
		return "", errors.New("No token in form")
	}

	task, err := getTaskValue(form, ans.Id)
	if err != nil {
		return
	}

	values := url.Values{}
	values.Add("data.TaskScreenName", task)
	values.Add("data.LanguageId", ans.Language)
	values.Add("sourceCode", ans.Code)
	values.Add("csrf_token", tok)

	resp, err = sess.Client.PostForm(submitURL, values)
	if err != nil {
		return
	}

	doc, err = html.Parse(resp.Body)
	if err != nil {
		return
	}
	table := getTable(doc)
	if table == nil {
		return "", errors.New("No table found")
	}
	detailURL, err := getDetailURL(table)
	if err != nil {
		return
	}
	splitted := strings.Split(detailURL, "/")
	submissionId := splitted[len(splitted)-1]

	return submissionId, nil
}

func getTaskValue(node *html.Node, id string) (string, error) {
	if node.Type == html.ElementNode && node.Data == "select" {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode || c.Data != "option" {
				continue
			}
			if c.FirstChild.Type != html.TextNode {
				return "", errors.New("Invalid element in <option>")
			}
			if c.FirstChild.Data[0] == id[0] {
				for _, a := range c.Attr {
					if a.Key == "value" {
						return a.Val, nil
					}
				}
				return "", errors.New("No value attribute in <option>")
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		val, err := getTaskValue(c, id)
		if err != nil || val != "" {
			return val, err
		}
	}
	return "", nil
}

func getTable(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "table" {
		return node
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		m := getTable(c)
		if m != nil {
			return m
		}
	}
	return nil
}

func getDetailURL(table *html.Node) (string, error) {
	if table.Type != html.ElementNode || table.Data != "table" {
		return "", errors.New("Invalid input")
	}

	var tbody *html.Node = nil
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			tbody = c
		}
	}
	if tbody == nil {
		return "", errors.New("<tbody> not found")
	}

	var firstRow *html.Node = nil
	for c := tbody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			firstRow = c
			break
		}
	}
	if firstRow == nil {
		return "", errors.New("No row in <tbody>")
	}

	var detailCol *html.Node = nil
	for c := firstRow.LastChild; c != nil; c = c.PrevSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			detailCol = c
			break
		}
	}
	if detailCol == nil {
		return "", errors.New("No column in <tr>")
	}

	var detail *html.Node = nil
	for c := detailCol.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" {
			detail = c
		}
	}
	if detail == nil {
		return "", errors.New("No link in the last column")
	}
	for _, a := range detail.Attr {
		if a.Key == "href" {
			return a.Val, nil
		}
	}
	return "", errors.New("No link in <a>")
}
