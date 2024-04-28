package processinfo

import (
	"fmt"
	"strings"
	"sync"
)

type variableInformation struct {
	latestDashVariableName string
	variableNamesSet       map[string]struct{}
}

var varInfoOnce sync.Once
var varInfoInstance *variableInformation

func VariableInformation() *variableInformation {
	varInfoOnce.Do(func() {
		varInfoInstance = &variableInformation{
			latestDashVariableName: "_",
			variableNamesSet:       nil,
		}
	})

	return varInfoInstance
}

func (v *variableInformation) GetLatestDashVariableName() string {
	var sb strings.Builder
	if _, err := sb.WriteString(v.latestDashVariableName); err != nil {
		fmt.Println("error copying string!")
		fmt.Println(err)
		return ""
	}

	return sb.String()
}

func (v *variableInformation) SetLatestDashVariableName(name string) {
	//add mutex and waitgroup if thread safety is required
	v.latestDashVariableName = name
}

func (v *variableInformation) GetVariableNamesSet() map[string]struct{} {
	return v.variableNamesSet
}

func (v *variableInformation) SetVariableNamesSet(namesSet map[string]struct{}) {
	v.variableNamesSet = namesSet
}

func (v *variableInformation) NameIsUsed(name string) bool {
	_, ok := v.variableNamesSet[name]
	return ok
}
