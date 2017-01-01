package main

import (
	"log"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	file, err := os.Open("first_submission.html")
	if err != nil {
		fatal(err)
	}
	defer file.Close()
	p := &problem{}
	err = p.retrieveSubmissionID(file)
	logger.Println(p.submissionID)
}
