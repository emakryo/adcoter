package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var program string
var logger *log.Logger
var debug_out io.WriteCloser

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

func find(attrs []html.Attribute, key string) (val string) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func fatal(v interface{}) {
	fmt.Printf("%s: %v\n", program, v)
	os.Exit(1)
}

func output(stat status) {
	logger.Printf("%d test cases\n", len(stat.caseName))
	ac := true
	for _, s := range stat.caseState {
		if s != "AC" {
			ac = false
			break
		}
	}

	if ac {
		fmt.Printf("AC (%d cases)\n", len(stat.caseState))
		return
	}

	for i, n := range stat.caseName {
		fmt.Printf("%s\t%s\n", stat.caseState[i], n)
	}
}

func main() {
	program = os.Args[0]

	var arg = parseArg()
	var out io.Writer
	var err error
	if arg.verbose {
		out = os.Stdout
	} else {
		out, err = os.Create(os.DevNull)
		if err != nil {
			fatal(err)
		}
	}
	logger = log.New(out, "", log.LstdFlags|log.Lshortfile)

	logger.Println(arg.url, arg.problem)
	sess, err := newSession(arg.url)
	if err != nil {
		fatal(err)
	}

	submissionID, err := sess.submit(arg.problem, arg.source, "14")
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Judging")
	var stat status
	var cnt = 0
	debug_out, err = os.Create(os.DevNull)
	if err != nil {
		fatal(err)
	}
	err = errors.New("Dummy Error")
	for err != nil {
		if arg.debug {
			debug_out.Close()
			out, err0 := os.Create(fmt.Sprintf("submissions%d.html", cnt))
			debug_out = out
			cnt += 1
			if err0 != nil {
				fatal(err0)
			}
		}
		stat, err = sess.status(submissionID)
		fmt.Printf(".")
		logger.Println(err)
		time.Sleep(time.Second)
	}
	fmt.Printf("\n")
	output(stat)
}
