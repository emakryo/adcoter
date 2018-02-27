package contest

import "fmt"

type Contest interface {
	GetURL() string
	Login()
	Submit(Answer) (string, error)
	Status(string) (Status, error)
}

type Base struct {
	url string
}

func NewContest(url string) *Base {
	return &Base{url}
}

func (c Base) GetURL() string {
	return c.url
}

type Answer struct {
	Id string
	Language string
	Source string
}

type Status struct {
	CaseName  []string
	CaseState []string
}

func (stat Status) Output() {
//	logger.Printf("%d test cases\n", len(stat.caseName))
	ac := true
	for _, s := range stat.CaseState {
		if s != "AC" {
			ac = false
			break
		}
	}

	if ac {
		fmt.Printf("AC (%d cases)\n", len(stat.CaseState))
		return
	}

	for i, n := range stat.CaseName {
		fmt.Printf("%s\t%s\n", stat.CaseState[i], n)
	}
}

