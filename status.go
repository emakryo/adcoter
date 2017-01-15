package main

import (
	"errors"
	"golang.org/x/net/html"
	"io"
	"strings"
)

type status struct {
	caseName  []string
	caseState []string
}

func (sess *session) status(id string) (stat status, err error) {
	resp, err := sess.get("/submissions/" + id)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	rd := io.TeeReader(resp.Body, debug_out)
	node, err := html.Parse(rd)
	if err != nil {
		return
	}
	return parseSubmission(node)
}

func parseSubmission(node *html.Node) (stat status, err error) {
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

func parseTable(node *html.Node) (stat status, err error) {
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
		for td := tr.FirstChild; td != nil; td = td.NextSibling {
			if td.Type != html.ElementNode || td.Data != "td" {
				continue
			}
			if td.FirstChild == nil {
				return stat, errors.New("No items in td")
			}
			if col == 0 {
				stat.caseName = append(stat.caseName, td.FirstChild.Data)
			}
			if col == 1 {
				if td.FirstChild.FirstChild == nil {
					return stat, errors.New("Invalid state")
				}
				stat.caseState = append(stat.caseState, td.FirstChild.FirstChild.Data)
			}
			col += 1
		}
	}

	return stat, nil
}
