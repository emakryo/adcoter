package beta

import (
	"errors"
	"io"
	"fmt"
	"github.com/emakryo/adcoter/status"
	"github.com/emakryo/adcoter/util"
	"golang.org/x/net/html"
	"net/http"
)

func (sess *Session) Status(id string) (stat status.Status, err error) {
	statusURL := fmt.Sprintf("%s/submissions/%s", sess.Url.String(), id)
	Logger.Printf("GET '%s'\n", statusURL)
	resp, err := sess.Client.Get(statusURL)
	Logger.Printf("GET '%s': %s\n", statusURL, resp.Status)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("GET '%s': %s\n", statusURL, resp.Status)
		return
	}

	return getStatus(resp.Body)
}

func getStatus(r io.Reader) (stat status.Status, err error) {
	Logger.Println("---Body of response---")
	tee := io.TeeReader(r, Logger)
	doc, err := html.Parse(tee)
	if err != nil {
		return
	}
	Logger.Println("---end of body---")


	tables := util.FindAll(doc, util.NewElemCond("table"))
	Logger.Printf("Found %d table(s)\n", len(tables))

	summary := getSummary(tables[0])
	stat.Summary = summary
	if tables == nil {
		err = fmt.Errorf("No table found")
	}

	if summary == "" {
		err = errors.New("Failed to get the status")
		return
	}

	if summary == "WJ" {
		return
	} else if summary == "CE" {
		stat.Error, err = getCompileError(doc)
		return
	}

	if len(tables) < 3 {
		Logger.Println("Too few tables")
		Logger.Println("Changing status to WJ")
		stat.Summary = "WJ"
		return
	}

	err = getAllTestCases(tables[2], &stat)
	return
}

func getSummary(table *html.Node) (summary string) {
	t := util.ParseTable(table)
	for _, r := range t {
		if r[0] == "Status" {
			summary = r[1]
		}
	}
	Logger.Println("Summary: " + summary)
	return summary
}

func getAllTestCases(table *html.Node, stat *status.Status) (err error) {
	var tbody = util.FindFirst(table, util.NewElemCond("tbody"))
	if tbody == nil {
		err = fmt.Errorf("No tbody found in the test case table")
		return
	}

	t := util.ParseTable(tbody)

	for _, r := range t {
		stat.Add(r[0], r[1])
	}
	return
}

func getCompileError(doc *html.Node) (string, error) {
	pre := util.FindAll(doc, util.NewElemCond("pre"))
	if len(pre) < 2 {
		return "", fmt.Errorf("Too few <pre> elements: %d", len(pre))
	}
	return util.InnerText(pre[1]), nil
}
