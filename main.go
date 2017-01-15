package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var program string
var logger *log.Logger
var debug_out io.WriteCloser

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
