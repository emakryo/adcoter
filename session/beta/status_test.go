package beta

import (
	"golang.org/x/net/html"
	"os"
	"testing"
)

func TestStatusParser(t *testing.T) {
	statusFile := "test_files/abc10C_6176545.htm"
	f, err := os.Open(statusFile)
	defer f.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
	doc, err := html.Parse(f)
	if err != nil {
		t.Errorf("%v", err)
	}
	table := getLastTable(doc)
	//html.Render(os.Stdout, table)

	if table == nil {
		t.Errorf("No table found")
	}

	stat, err := getStatus(table)
	if err != nil {
		t.Errorf("%v", err)
	}

	stat.Output()
}
