package main

import (
	"encoding/json"
	"errors"
	"github.com/dgiagio/getpass"
	"io/ioutil"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"github.com/emakryo/adcoter/session"
	"github.com/emakryo/adcoter/session/beta"
	"github.com/emakryo/adcoter/session/old"
	"github.com/emakryo/adcoter/status"
	"github.com/emakryo/adcoter/answer"
)

type Contest struct {
	sess session.Session
}

func newContest(url string, isBeta bool) (c *Contest, err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}
	cacheDir := usr.HomeDir + "/.adcoter"
	os.MkdirAll(cacheDir, 0700)
	var cookiePath = usr.HomeDir + "/.adcoter/cookies"
	if isBeta {
		sess, err := beta.New(url)
		if err != nil {
			return nil, err
		}
		c = &Contest{sess: sess}
		cookiePath += "_beta"
	} else {
		sess, err := old.New(url)
		if err != nil {
			return nil, err
		}
		c = &Contest{sess: sess}
	}

	if !c.sess.Valid() {
		return nil, errors.New("Invalid contest")
	}

	err = c.loadCookies(cookiePath)
	if err != nil {
		return
	}

	if c.sess.IsLoggedin() {
		return
	}

	var name, password string
	fmt.Printf("User ID: ")
	fmt.Scan(&name)
	password, err = getpass.GetPassword("Password: ")
	if err != nil {
		return
	}

	err = c.sess.Login(name, password)
	if err != nil {
		return
	}

	if c.sess.IsLoggedin() {
		err = c.saveCookies(cookiePath)
	} else {
		err = errors.New("Login Failured")
	}
	return
}

func (c *Contest) loadCookies(cookiePath string) (err error) {
	var cookies []*http.Cookie
	marshaled, err := ioutil.ReadFile(cookiePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(marshaled, &cookies)
	if err != nil {
		return
	}
	c.sess.SetCookies(cookies)
	return nil
}

func (c *Contest) saveCookies(cookiePath string) error {
	cookies := c.sess.Cookies()
	marshaled, err := json.Marshal(cookies)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cookiePath, marshaled, 0400)
	if err != nil {
		return err
	}
	return nil
}

func (c *Contest) Submit(ans answer.Answer) (id string, err error) {
	return c.sess.Submit(ans)
}

func (c * Contest) Status(id string) (stat status.Status, err error) {
	return c.sess.Status(id)
}
