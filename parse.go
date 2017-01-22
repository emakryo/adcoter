package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type argument struct {
	url     string
	verbose bool
	debug   bool
	problem string
	source  string
	language string
}

func exitWithUsage(err interface{}) {
	fmt.Printf("%s: %v\n", program, err)
	flag.Usage()
	os.Exit(255)
}

func parseArg() (arg argument) {
	minf := -1000000
	verb := flag.Bool("v", false, "Verbose output")
	debug := flag.Bool("debug", false, "For debug")
	arc := flag.Int("arc", minf, "ARC")
	abc := flag.Int("abc", minf, "ABC")
	agc := flag.Int("agc", minf, "AGC")
	url := flag.String("u", "", "URL for the contest")
	prob := flag.String("p", "", "Problem ID")
	lang := flag.String("l", "", "Language ID")

	flag.Parse()

	arg.verbose = *verb
	arg.debug = *debug

	if flag.NArg() < 1 {
		exitWithUsage("No source file")
	} else if flag.NArg() > 1 {
		exitWithUsage("Too many source files")
	}

	x := 0
	if *arc != minf {
		x += 1
	}
	if *abc != minf {
		x += 1
	}
	if *agc != minf {
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

	if *arc != minf {
		if *arc < 0 {
			exitWithUsage("Invalid contest ID")
		}
		arg.url = fmt.Sprintf("arc%03d.contest.atcoder.jp", *arc)
	} else if *abc != minf {
		if *abc < 0 {
			exitWithUsage("Invalid contest ID")
		}
		arg.url = fmt.Sprintf("abc%03d.contest.atcoder.jp", *abc)
	} else if *agc != minf {
		if *agc < 0 {
			exitWithUsage("Invalid contest ID")
		}
		arg.url = fmt.Sprintf("agc%03d.contest.atcoder.jp", *agc)
	}

	if *url != "" {
		arg.url = *url
	}

	arg.url = "https://" + arg.url

	if *prob == "" {
		basename := path.Base(flag.Arg(0))
		arg.problem = strings.ToUpper(strings.Split(basename, ".")[0])
	} else {
		arg.problem = strings.ToUpper(*prob)
	}

	arg.source = flag.Arg(0)

	arg.language = *lang
	if *lang == "" {
		arg.language = detectLanguage(arg.source)
	} else {
		arg.language = *lang
	}

	return arg
}

var languages = []struct{
	language string
	id string
	suffix string
}{
	{"G++", "14", ".cpp"},
	{"GHC", "11", ".hs"},
}

func detectLanguage(filename string) (id string) {
	for _, l := range languages {
		if strings.HasSuffix(filename, l.suffix){
			return l.id
		}
	}

	printAvailable();
	os.Exit(255);
	return "";
}

func printAvailable() {
	for _, l := range languages {
		fmt.Printf("%s: %s\n", l.language, l.suffix)
	}
}
