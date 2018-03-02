package status

import "fmt"

type Status struct {
	caseNames  []string
	caseStates []string
}

func (stat *Status) Add(name, state string) {
	stat.caseNames = append(stat.caseNames, name)
	stat.caseStates = append(stat.caseStates, state)
}

func (stat Status) Output() {
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
