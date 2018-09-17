package status

import "fmt"

type Status struct {
	Summary string
	Error string
	caseNames  []string
	caseStates []string
}

func (stat *Status) Add(name, state string) {
	stat.caseNames = append(stat.caseNames, name)
	stat.caseStates = append(stat.caseStates, state)
}

func (stat Status) Output() {
	if stat.Summary == "AC" {
		fmt.Printf("AC (%d cases)\n", len(stat.caseStates))
		return
	}
	if stat.Summary == "CE" {
		fmt.Println("Compile Error")
		fmt.Println(stat.Error)
		return
	}

	for i, n := range stat.caseNames {
		fmt.Printf("%s\t%s\n", stat.caseStates[i], n)
	}
}
