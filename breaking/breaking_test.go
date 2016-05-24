package breaking

import "testing"

func TestDeleted(t *testing.T) {
	names := []string{
		"FuncParameterAdded",
		"FuncParamTypeChanged",
		"FuncRetTypeChanged",
		"InterfaceMethodAdded",
		"InterfaceMethodDeleted",
		"InterfaceMethodParameterAdded",
		"InterfaceMethParamTypeChanged",
		"InterfaceMethRetTypeChanged",
		"VarDeleted",
		"VarToFunc",
		"VarTypeChanged",
	}

	report, err := CompareFiles("fixtures/a/a.go", "fixtures/b/b.go")
	if err != nil {
		t.Error(err)
	}

	for _, obj := range report.Deleted {
		found := false
		for _, name := range names {
			if name == obj.Name() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s mistakenly marked as deleted", obj.Name())
		}
	}
}
