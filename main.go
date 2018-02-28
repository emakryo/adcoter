package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"
	"strings"
	"github.com/emakryo/adcoter/contest"
	"github.com/emakryo/adcoter/contest/old"
)

var program string

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

	submissionId, err := arg.contest.Submit(arg.answer)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Judging")
	var stat contest.Status
	for {
		stat, err = arg.contest.Status(submissionId)
		if err == nil {
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
	contest contest.Contest
	answer  contest.Answer
	verbose  bool
	debug    bool
}

func parseArgs() (arg argument) {
	beta := flag.Bool("beta", false, "use beta.atcoder.jp")
	arc := flag.Int("arc", -1, "ARC")
	abc := flag.Int("abc", -1, "ABC")
	agc := flag.Int("agc", -1, "AGC")
	url := flag.String("u", "", "URL for the contest")
	prob := flag.String("p", "", "Problem ID")
	lang := flag.String("l", "", "Language ID")

	flag.Parse()

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

	if *beta {
		exitWithUsage("Beta version is not implemented yet");
	} else {
		var c *old.Contest
		var err error
		if *url != "" {
			c, err = old.New(*url)
		} else if *arc > 0 {
			c, err = old.NewFromId("arc", *arc)
		} else if *abc > 0 {
			c, err = old.NewFromId("abc", *abc)
		} else if *agc > 0 {
			c, err = old.NewFromId("agc", *agc)
		}
		if err != nil {
			fatal(err)
		}
		arg.contest = c
	}

	if flag.NArg() < 1 {
		exitWithUsage("No source file is specified")
	} else if flag.NArg() > 1 {
		exitWithUsage("Too many source files are specified")
	}
	source := flag.Arg(0)

	id := ""
	if *prob == "" {
		basename := path.Base(source)
		id = strings.ToUpper(strings.Split(basename, ".")[0])
	} else {
		id = strings.ToUpper(*prob)
	}

	language := *lang
	if *lang == "" {
		language = detectLanguage(source)
	}

	arg.answer = contest.Answer{
		Id: id,
		Source: source,
		Language: language,
	}

	return arg
}
