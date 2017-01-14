package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dgiagio/getpass"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type problem struct {
	ContestID string
	ProblemID string

	url          *url.URL
	taskID       string
	languageID   int
	client       *http.Client
	cacheFile    string
	submissionID string
}

type status struct {
	caseName  []string
	caseState []string
}

var prog string
var logger *log.Logger
var debug_out io.WriteCloser

func newProblem(contestID, problemID string) (p problem, err error) {
	p.ContestID = contestID
	p.ProblemID = problemID
	rawContestURL := fmt.Sprintf("https://%s.contest.atcoder.jp", contestID)
	contestURL, err := url.Parse(rawContestURL)
	if err != nil {
		return
	}
	p.url = contestURL

	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	p.client = &http.Client{Jar: jar}
	resp, err := p.client.Get(rawContestURL)
	defer resp.Body.Close()
	if !p.validContest() {
		err = errors.New(contestID + ": Invalid contest")
		return
	}
	err = p.retrieveTaskID()
	if err != nil {
		return
	}

	usr, err := user.Current()
	if err != nil {
		return
	}
	cacheDir := usr.HomeDir + "/.adcoter"
	os.MkdirAll(cacheDir, 0700)
	p.cacheFile = usr.HomeDir + "/.adcoter/cookies"
	return p, nil
}

func (p *problem) validContest() bool {
	resp, err := p.client.Get(p.url.String())
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)

	var tok html.Token
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.TextToken && tok.Data == "404" {
			return false
		}
	}

	return true
}

func (p *problem) retrieveTaskID() (err error) {
	resp, err := p.client.Get(p.url.String() + "/assignments")
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
		if tok.Type == html.TextToken && column == 1 && tok.Data == p.ProblemID {
			match = true
		}
		if tok.Type == html.StartTagToken && column == 5 && tok.Data == "a" && match {
			submitURL = find(tok.Attr, "href")
			break
		}
	}

	if submitURL == "" {
		return errors.New(fmt.Sprintf("%v: Problem ID not found", p.ProblemID))
	}

	parsed, err := url.Parse(p.url.String() + submitURL)
	if err != nil {
		return
	}
	vals, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return
	}
	id, ok := vals["task_id"]
	if !ok {
		return errors.New(fmt.Sprintf("%s/assignments : Parse error", p.url.String()))
	}
	p.taskID = id[0]
	return nil
}

func (p *problem) loginSuccess() bool {
	resp, err := p.client.Get(p.url.String() + "/submit")
	if err != nil {
		return false
	}
	return resp.Request.URL.Path == "/submit"
}

func (p *problem) authorize() (err error) {
	err = p.loadCookies()
	if err != nil {
		return
	}
	if !p.loginSuccess() {
		return errors.New("Invalid cookies")
	}
	//log.Println("Login with cached cookies");
	return nil
}

func (p *problem) login() (err error) {
	var name, password string
	fmt.Printf("User ID: ")
	fmt.Scan(&name)
	password, err = getpass.GetPassword("Password: ")
	if err != nil {
		return
	}
	var values = url.Values{}
	values.Add("name", name)
	values.Add("password", password)
	resp, err := p.client.PostForm(p.url.String()+"/login", values)
	if err != nil {
		return
	}
	resp.Body.Close()
	if !p.loginSuccess() {
		return errors.New("login: Invalid user ID or password")
	}
	//log.Printf("Login successful");
	p.saveCookies()
	return nil
}

func (p *problem) saveCookies() {
	cookies := p.client.Jar.Cookies(p.url)
	marshaled, err := json.Marshal(cookies)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(p.cacheFile, marshaled, 0400)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *problem) loadCookies() (err error) {
	var cookies []*http.Cookie
	marshaled, err := ioutil.ReadFile(p.cacheFile)
	if err != nil {
		//log.Println(err);
		return
	}
	err = json.Unmarshal(marshaled, &cookies)
	if err != nil {
		//log.Println(err);
		return
	}
	p.client.Jar.SetCookies(p.url, cookies)
	return nil
}

func (p *problem) submit(codePath string, languageID string) (err error) {
	resp, err := p.client.Get(p.url.String() + "/submit")
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
		return errors.New(fmt.Sprintf("%s/submit : Parse Failure", p.url.String()))
	}

	content, err := ioutil.ReadFile(codePath)
	if err != nil {
		return
	}

	data := url.Values{}
	data.Add("__session", session)
	data.Add("task_id", p.taskID)
	data.Add("language_id_"+p.taskID, languageID)
	data.Add("source_code", string(content))

	resp, err = p.client.PostForm(p.url.String()+"/submit", data)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.Request.URL.Path != "/submissions/me" {
		return errors.New(fmt.Sprintf("%s: Submission failure", codePath))
	}

	wr, err := os.Create("first_submission.html")
	defer wr.Close()
	if err != nil {
		return
	}
	rd := io.TeeReader(resp.Body, wr)

	err = p.retrieveSubmissionID(rd)

	if err != nil {
		return errors.New(fmt.Sprintf("%s: Submission ID not retrieved",
			resp.Request.URL))
	}

	logger.Printf("submission ID: %s\n", p.submissionID)

	return nil
}

func (p *problem) retrieveSubmissionID(body io.Reader) (err error) {
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
				p.submissionID = splited[len(splited)-1]
				return nil
			}
		}
	}

	return errors.New("Submission ID not found")

}

func (p *problem) status() (stat status, err error) {
	resp, err := p.client.Get(p.url.String() + "/submissions/" + p.submissionID)
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
	fmt.Printf("%s: %v\n", prog, v)
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
	prog = os.Args[0]

	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()
	var out io.Writer
	var err error
	if *verbose {
		out = os.Stdout
	} else {
		out, err = os.Create(os.DevNull)
		if err != nil {
			fatal(err)
		}
	}
	logger = log.New(out, "", log.LstdFlags|log.Lshortfile)
	if flag.NArg() < 4 {
		fatal("Too small number of arguments")
	}
	contestType := flag.Arg(0)
	contestID, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fatal(fmt.Sprintf("%v: Contest ID must be an integer", flag.Arg(1)))
	}
	problemID := strings.ToUpper(flag.Arg(2))
	sourcePath := flag.Arg(3)

	contest := fmt.Sprintf("%s%03d", strings.ToLower(contestType), contestID)
	p, err := newProblem(contest, problemID)
	if err != nil {
		fatal(err)
	}

	err = p.authorize()
	if err != nil {
		err = p.login()
		if err != nil {
			fatal(err)
		}
	}

	err = p.submit(sourcePath, "14")
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
		if *verbose {
			debug_out.Close()
			out, err0 := os.Create(fmt.Sprintf("submissions%d.html", cnt))
			debug_out = out
			cnt += 1
			if err0 != nil {
				fatal(err0)
			}
		}
		stat, err = p.status()
		fmt.Printf(".")
		logger.Println(err)
		time.Sleep(time.Second)
	}
	fmt.Printf("\n")
	output(stat)
}
