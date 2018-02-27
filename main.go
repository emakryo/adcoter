package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"github.com/emakryo/adcoter/contest"
)

var program string
var logger *log.Logger
var debug_out io.WriteCloser

func fatal(v interface{}) {
	fmt.Printf("%s: %v\n", program, v)
	os.Exit(1)
}

func main() {
	program = os.Args[0]

	var arg = parseArgs()
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

	logger.Println(arg.contest.GetURL(), arg.answer.Id)

	submissionID, err := arg.contest.Submit(arg.answer)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Judging")
	var stat contest.Status
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
		stat, err = arg.contest.Status(submissionID)
		fmt.Printf(".")
		logger.Println(err)
		time.Sleep(time.Second)
	}
	fmt.Printf("\n")
	stat.Output()
}
