package contest

import "fmt"

type Contest interface {
	Submit(Answer) (string, error)
	Status(string) (Status, error)
}

type Answer struct {
	Id string
	Language string
	Source string
}

type Status struct {
	caseNames  []string
	caseStates []string
}

func (stat *Status) Add(name, state string) {
	stat.caseNames = append(stat.caseNames, name)
	stat.caseStates = append(stat.caseStates, state)
}

func (stat Status) Output() {
//	logger.Printf("%d test cases\n", len(stat.caseName))
	ac := true
	for _, s := range stat.caseStates {
		if s != "AC" {
			ac = false
			break
		}
	}

	if ac {
		fmt.Printf("AC (%d cases)\n", len(stat.caseStates))
		return
	}

	for i, n := range stat.caseNames {
		fmt.Printf("%s\t%s\n", stat.caseStates[i], n)
	}
}
