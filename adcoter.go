package main

import (
	"fmt"
	"log"
	"os"
	"flag"
	"strings"
	"strconv"
	"errors"
	"os/user"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"golang.org/x/net/html"
	"github.com/dgiagio/getpass"
);

type problem struct {
	ContestID string
	ProblemID string

	url *url.URL
	taskID string
	languageID int
	client *http.Client
	cacheFile string
}

var prog string;

func newProblem(contestID, problemID string) (p problem, err error){
	p.ContestID = contestID;
	p.ProblemID = problemID;
	rawContestURL := fmt.Sprintf("https://%s.contest.atcoder.jp", contestID);
	contestURL, err := url.Parse(rawContestURL);
	if err != nil { return; }
	p.url = contestURL;

	jar, err := cookiejar.New(nil);
	if err != nil { return; }
	p.client = &http.Client{Jar: jar};
	resp, err := p.client.Get(rawContestURL);
	defer resp.Body.Close()
	if !p.validContest() {
		err = errors.New(contestID+": Invalid contest");
		return;
	}
	err = p.retrieveTaskID();
	if err != nil{ return ; }

	usr, err := user.Current();
	if err != nil { return; };
	cacheDir := usr.HomeDir+"/.adcoter";
	os.MkdirAll(cacheDir, 0700);
	p.cacheFile = usr.HomeDir+"/.adcoter/cookies";
	return p, nil
}

func (p *problem)validContest() bool {
	resp, err := p.client.Get(p.url.String());
	if err != nil { return false; }
	defer resp.Body.Close();
	tokenizer := html.NewTokenizer(resp.Body);

	var tok html.Token;
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.TextToken && tok.Data == "404" {
			return false;
		}
	}

	return true;
}

func (p *problem)retrieveTaskID() (err error) {
	resp, err := p.client.Get(p.url.String()+"/assignments");
	if err != nil { return; }
	defer resp.Body.Close();
	tokenizer := html.NewTokenizer(resp.Body);
	var tok html.Token;
	var column int;
	var match = false;
	var submitURL string;
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.StartTagToken && tok.Data == "tr" {
			column = 0;
			match = false;
		}
		if tok.Type == html.StartTagToken && tok.Data == "td" {
			column += 1;
		}
		if tok.Type == html.TextToken && column == 1 && tok.Data == p.ProblemID {
			match = true;
		}
		if tok.Type == html.StartTagToken && column == 5 && tok.Data == "a" && match {
			submitURL = find(tok.Attr, "href");
			break;
		}
	}

	if submitURL == "" {
		return errors.New(fmt.Sprintf("%v: Problem ID not found", p.ProblemID));
	}

	parsed, err := url.Parse(p.url.String()+submitURL);
	if err != nil { return; }
	vals, err := url.ParseQuery(parsed.RawQuery);
	if err != nil { return; }
	id, ok := vals["task_id"];
	if !ok {
		return errors.New(fmt.Sprintf("%s/assignments : Parse error",p.url.String()));
	}
	p.taskID = id[0];
	return nil;
}

func (p *problem)loginSuccess() bool {
	resp, err := p.client.Get(p.url.String()+"/submit");
	if err != nil { return false; }
	return resp.Request.URL.Path == "/submit"
}

func (p *problem) authorize() (err error) {
	err = p.loadCookies();
	if err != nil { return; }
	if !p.loginSuccess() {
		return  errors.New("Invalid cookies");
	}
	//log.Println("Login with cached cookies");
	return nil
}

func (p *problem)login() (err error) {
	var name, password string;
	fmt.Printf("User ID: ");
	fmt.Scan(&name);
	password, err = getpass.GetPassword("Password: ");
	if err != nil { return; }
	var values = url.Values{};
	values.Add("name", name);
	values.Add("password", password);
	resp, err := p.client.PostForm(p.url.String() + "/login", values);
	if err != nil { return; }
	resp.Body.Close();
	if !p.loginSuccess() {
		return errors.New("login: Invalid user ID or password");
	}
	//log.Printf("Login successful");
	p.saveCookies();
	return nil
}

func (p *problem)saveCookies() {
	cookies := p.client.Jar.Cookies(p.url);
	marshaled, err := json.Marshal(cookies);
	if err != nil {
		log.Fatal(err);
	}
	err = ioutil.WriteFile(p.cacheFile, marshaled, 0400);
	if err != nil {
		log.Fatal(err);
	}
}

func (p *problem)loadCookies() (err error) {
	var cookies []*http.Cookie;
	marshaled, err := ioutil.ReadFile(p.cacheFile);
	if err != nil {
		//log.Println(err);
		return;
	}
	err = json.Unmarshal(marshaled, &cookies);
	if err != nil {
		//log.Println(err);
		return;
	}
	p.client.Jar.SetCookies(p.url, cookies);
	return nil;
}

func (p *problem)submit(codePath string, languageID string) (err error) {
	resp, err := p.client.Get(p.url.String()+"/submit");
	if err != nil { return; }
	defer resp.Body.Close();
	tokenizer := html.NewTokenizer(resp.Body);
	var tok html.Token;
	var session string;
	for ; tokenizer.Next() != html.ErrorToken; tok = tokenizer.Token() {
		if tok.Type == html.StartTagToken && tok.Data == "input" &&
			find(tok.Attr, "name") == "__session" {
			session = find(tok.Attr, "value");
		}
	}
	if session == "" {
		return errors.New(fmt.Sprintf("%s/submit : Parse Failure", p.url.String()));
	}

	content, err := ioutil.ReadFile(codePath);
	if err != nil { return; }

	data := url.Values{};
	data.Add("__session", session);
	data.Add("task_id", p.taskID);
	data.Add("language_id_"+p.taskID, languageID);
	data.Add("source_code", string(content));

	resp, err = p.client.PostForm(p.url.String() + "/submit", data);
	if err != nil { return; }
	defer resp.Body.Close();
	if resp.Request.URL.Path != "/submissions/me" {
		return errors.New(fmt.Sprintf("%s: Submission failure", codePath));
	}

	return nil;
}

func find(attrs []html.Attribute, key string) (val string) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val;
		}
	}
	return "";
}

func fatal(v interface{}) {
	fmt.Printf("%s: %v\n", prog, v);
	os.Exit(1);
}

func main(){
	prog = os.Args[0];

	flag.Parse();
	if flag.NArg() < 4 {
		fatal("Too small number of arguments");
	}
	contestType := flag.Arg(0);
	contestID, err := strconv.Atoi(flag.Arg(1));
	if err != nil {
		fatal(fmt.Sprintf("%v: Contest ID must be an integer", flag.Arg(1)));
	}
	problemID := strings.ToUpper(flag.Arg(2));
	sourcePath := flag.Arg(3);

	contest := fmt.Sprintf("%s%03d", strings.ToLower(contestType), contestID);
	p, err := newProblem(contest, problemID);
	if err != nil {
		fatal(err);
	}

	err = p.authorize();
	if err != nil {
		err = p.login();
		if err != nil {
			fatal(err);
		}
	}

	err = p.submit(sourcePath, "14");
	if err != nil {
		fatal(err);
	}
}
