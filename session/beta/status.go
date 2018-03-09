package beta

import (
	//"os"
	"fmt"
	"github.com/emakryo/adcoter/status"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

func (sess *Session) Status(id string) (stat status.Status, err error) {
	statusURL := fmt.Sprintf("%s/submissions/%s", sess.Url.String(), id)
	resp, err := sess.Client.Get(statusURL)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Status of %s is not 200 OK: %v", statusURL, resp.StatusCode)
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	table := getLastTable(doc)
	if table == nil {
		err = fmt.Errorf("No table found in %v", statusURL)
	}

	return getStatus(table)
}

func getLastTable(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "table" {
		return node
	}

	for c := node.LastChild; c != nil; c = c.PrevSibling {
		m := getLastTable(c)
		if m != nil {
			return m
		}
	}
	return nil
}

func getStatus(table *html.Node) (stat status.Status, err error) {
	var tbody *html.Node
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			tbody = c
			break
		}
	}
	if tbody == nil {
		err = fmt.Errorf("No tbody")
		return
	}

	for c := tbody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "tr" {
			continue
		}
		n, s := getTestCase(c)
		if n == "" || s == "" {
			err = fmt.Errorf("Failed to parse row")
		}
		stat.Add(n, s)
	}

	return
}

func getTbody(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "tbody" {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		m := getTbody(c)
		if m != nil {
			return m
		}
	}
	return nil
}

func getTestCase(node *html.Node) (n string, s string) {
	var row_idx = 0
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			if row_idx == 0 {
				n = innerText(c)
			} else if row_idx == 1 {
				s = innerText(c)
			}
			row_idx += 1
		}
	}
	return
}

func innerText(node *html.Node) string {
	if node.Type == html.TextNode && node.FirstChild == nil {
		return strings.TrimSpace(node.Data)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		s := innerText(c)
		if s != "" {
			return s
		}
	}

	return ""
}
