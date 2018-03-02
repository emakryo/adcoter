package old

import (
	"errors"
	"github.com/emakryo/adcoter/status"
	"golang.org/x/net/html"
	"strings"
)

func (sess *Session) Status(id string) (stat status.Status, err error) {
	resp, err := sess.get("/submissions/" + id)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	node, err := html.Parse(resp.Body)
	if err != nil {
		return
	}
	return parseSubmission(node)
}

func parseSubmission(node *html.Node) (stat status.Status, err error) {
	switch node.Type {
	case html.DocumentNode:
		for next := node.FirstChild; next != nil; next = next.NextSibling {
			stat, err = parseSubmission(next)
			if err == nil {
				return
			}
		}
	case html.ElementNode:
		if node.Data == "h4" && node.FirstChild != nil {
			if strings.Contains(node.FirstChild.Data, "Test case") {
				for sib := node.NextSibling; sib != nil; sib = sib.NextSibling {
					if sib.Type == html.ElementNode && sib.Data == "table" {
						return parseTable(sib)
					}
				}
			}
		}
		for next := node.FirstChild; next != nil; next = next.NextSibling {
			stat, err = parseSubmission(next)
			if err == nil {
				return
			}
		}
	}

	return stat, errors.New("Not found")
}

func parseTable(node *html.Node) (stat status.Status, err error) {
	var tbody *html.Node
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		if next.Type == html.ElementNode && next.Data == "tbody" {
			tbody = next
			break
		}
	}
	if tbody == nil {
		return stat, errors.New("No tbody")
	}
	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type != html.ElementNode || tr.Data != "tr" {
			continue
		}
		var col = 0
		if tr.FirstChild == nil {
			return stat, errors.New("No column in the row")
		}
		var state, name string
		for td := tr.FirstChild; td != nil; td = td.NextSibling {
			if td.Type != html.ElementNode || td.Data != "td" {
				continue
			}
			if td.FirstChild == nil {
				return stat, errors.New("No items in td")
			}
			if col == 0 {
				name = td.FirstChild.Data
			}
			if col == 1 {
				if td.FirstChild.FirstChild == nil {
					return stat, errors.New("Invalid state")
				}
				state = td.FirstChild.FirstChild.Data
				stat.Add(name, state)
			}
			col += 1
		}
	}

	return stat, nil
}
