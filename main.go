package main

import (
	"flag"
	"fmt"
	"github.com/emakryo/adcoter/answer"
	"github.com/emakryo/adcoter/session/beta"
	"github.com/emakryo/adcoter/session/old"
	"github.com/emakryo/adcoter/status"
	"github.com/emakryo/adcoter/util"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var program string
var logger *util.Logger

const retry_count = 5

func fatal(v interface{}) {
	fmt.Printf("%s: %v\n", program, v)
	os.Exit(1)
}

func exitWithUsage(err interface{}) {
	fmt.Printf("%s: %v\n", program, err)
	flag.Usage()
	os.Exit(255)
}

func main() {
	program = os.Args[0]
	var arg = parseArgs()

	submissionId, err := arg.contest.Submit(*arg.answer)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Judging")
	var stat status.Status
	var retry = 0
	for {
		stat, err = arg.contest.Status(submissionId)
		logger.Println(stat.Summary)
		logger.Println(err)
		if err != nil {
			logger.Printf("Error: %v", err)
			retry += 1
			if retry > retry_count {
				logger.Fatalf("Existing after %d retry", retry)
			}
			logger.Printf("Retrying...(%d)", retry)
		} else if stat.Summary != "WJ" {
			break
		}
		fmt.Printf(".")
		time.Sleep(time.Second)
	}
	fmt.Printf("\n")
	stat.Output()
}

var languages = []struct {
	language string
	id       string
	suffix   string
}{
	{"C++14 (GCC 5.4.1)", "3003", ".cpp"},
	{"Haskell (GHC 7.10.3)", "3014", ".hs"},
}

func detectLanguage(filename string) string {
	for _, l := range languages {
		if strings.HasSuffix(filename, l.suffix) {
			return l.id
		}
	}

	printAvailable()
	os.Exit(255)
	return ""
}

func printAvailable() {
	for _, l := range languages {
		fmt.Printf("%s: %s\n", l.language, l.suffix)
	}
}

type argument struct {
	contest *Contest
	answer  *answer.Answer
	verbose bool
	debug   bool
	beta    bool
}

func parseArgs() (arg argument) {
	isBeta := flag.Bool("beta", false, "Use beta.atcoder.jp")
	arc := flag.Int("arc", -1, "ARC")
	abc := flag.Int("abc", -1, "ABC")
	agc := flag.Int("agc", -1, "AGC")
	url := flag.String("u", "", "URL for the contest")
	prob := flag.String("p", "", "Problem ID")
	lang := flag.String("l", "", "Language ID")
	verb := flag.Bool("v", false, "Verbose output")
	logFile := flag.String("log", "", "Log file")

	flag.Parse()

	var logWriter io.Writer
	var err error
	if *logFile != "" {
		logWriter, err = os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			exitWithUsage(err.Error())
		}
	} else {
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Could not find current user")
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		filename := filepath.Join(usr.HomeDir, ".adcoter", "log"+timestamp)
		logWriter, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			exitWithUsage(err.Error())
		}
	}

	if *verb {
		logWriter = io.MultiWriter(logWriter, os.Stdout)
	}
	logger = util.NewLogger(logWriter, "", log.Lshortfile)
	logger.Println("Verbose output enabled")
	beta.Logger = logger
	old.Logger = logger

	x := 0
	if *arc > 0 {
		x += 1
	}
	if *abc > 0 {
		x += 1
	}
	if *agc > 0 {
		x += 1
	}
	if *url != "" {
		x += 1
	}
	if x > 1 {
		exitWithUsage("Multiple types of contests are specified")
	}
	if x < 1 {
		exitWithUsage("No contest is specified")
	}

	var contest_type string
	var contest_id int
	if *arc > 0 {
		contest_type = "arc"
		contest_id = *arc
		*isBeta = true
	} else if *abc > 0 {
		contest_type = "abc"
		contest_id = *abc
		*isBeta = true
	} else if *agc > 0 {
		contest_type = "agc"
		contest_id = *agc
		*isBeta = true
	}

	var rawurl string
	arg.beta = *isBeta
	if *url == "" {
		if *isBeta {
			rawurl = beta.ContestURL(contest_type, contest_id)
		} else {
			rawurl = old.ContestURL(contest_type, contest_id)
		}
	} else {
		rawurl = *url
	}
	arg.contest, err = newContest(rawurl, *isBeta)
	if err != nil {
		fatal(err)
	}

	if flag.NArg() < 1 {
		exitWithUsage("No source file is specified")
	} else if flag.NArg() > 1 {
		exitWithUsage("Too many source files are specified")
	}
	source := flag.Arg(0)

	code, err := ioutil.ReadFile(source)
	if err != nil {
		fatal(err)
	}

	id := ""
	if *prob == "" {
		basename := filepath.Base(source)
		id = strings.ToUpper(strings.Split(basename, ".")[0])
	} else {
		id = strings.ToUpper(*prob)
	}

	language := *lang
	if *lang == "" {
		language = detectLanguage(source)
	}

	arg.answer = &answer.Answer{
		Id:       id,
		Code:     string(code),
		Language: language,
	}

	return arg
}
