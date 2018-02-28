package old

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgiagio/getpass"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/user"
)

type Contest struct {
	client     *http.Client
	url        *url.URL
	cookiePath string
}

func New(rawurl string) (c *Contest, err error) {
	c = &Contest{
		client: nil,
		url:    nil,
	}
	c.url, err = url.Parse(rawurl)
	if err != nil {
		return
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	c.client = &http.Client{Jar: jar}
	if !c.valid() {
		err = errors.New(rawurl + ": Invalid contest")
		return
	}

	usr, err := user.Current()
	if err != nil {
		return
	}
	cacheDir := usr.HomeDir + "/.adcoter"
	os.MkdirAll(cacheDir, 0700)
	c.cookiePath = usr.HomeDir + "/.adcoter/cookies"

	err = c.authorize()
	if err != nil {
		err = c.login()
		if err != nil {
			return
		}
	}

	return c, nil
}

func NewFromId(t string, id int) (c *Contest, err error) {
	return New(fmt.Sprintf("https://%s%03d.contest.atcoder.jp", t, id))
}

func (c *Contest) get(path string) (resp *http.Response, err error) {
	resp, err = c.client.Get(c.url.String() + path)
	return
}

func (c *Contest) postForm(path string, values url.Values) (resp *http.Response, err error) {
	resp, err = c.client.PostForm(c.url.String()+path, values)
	return
}

func (c *Contest) valid() bool {
	resp, err := c.get("")
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

func (c *Contest) authorize() (err error) {
	err = c.loadCookies()
	if err != nil {
		return
	}
	if !c.loginSuccess() {
		return errors.New("Invalid cookies")
	}
	return nil
}

func (c *Contest) login() (err error) {
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
	resp, err := c.postForm("/login", values)
	if err != nil {
		return
	}
	resp.Body.Close()
	if !c.loginSuccess() {
		return errors.New("login: Invalid user ID or password")
	}
	//log.Printf("Login successful");
	c.saveCookies()
	return nil
}

func (c *Contest) loginSuccess() bool {
	resp, err := c.get("/submit")
	if err != nil {
		return false
	}
	return resp.Request.URL.Path == "/submit"
}

func (c *Contest) saveCookies() error {
	cookies := c.client.Jar.Cookies(c.url)
	marshaled, err := json.Marshal(cookies)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(c.cookiePath, marshaled, 0400)
	if err != nil {
		return err
	}
	return nil
}

func (c *Contest) loadCookies() (err error) {
	var cookies []*http.Cookie
	marshaled, err := ioutil.ReadFile(c.cookiePath)
	if err != nil {
		//log.Println(err);
		return
	}
	err = json.Unmarshal(marshaled, &cookies)
	if err != nil {
		//log.Println(err);
		return
	}
	c.client.Jar.SetCookies(c.url, cookies)
	return nil
}
