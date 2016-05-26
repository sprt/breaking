package breaking

import "testing"

const (
	filenamea = "fixtures/a/a.go"
	filenameb = "fixtures/b/b.go"
)

func TestDeleted(t *testing.T) {
	names := []string{
		"FuncParameterAdded",
		"FuncParamTypeChanged",
		"FuncResAdded",
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

	report, err := CompareFiles(filenamea, filenameb)
	if err != nil {
		t.Error(err)
	}

	for _, name := range names {
		deleted := false
		for _, obj := range report.Deleted {
			if obj.Name() == name {
				deleted = true
				break
			}
		}
		if !deleted {
			t.Errorf("not marked as deleted: %s", name)
		}
	}
}

func TestNotDeleted(t *testing.T) {
	names := []string{
		"FuncParamRenamed",
		"FuncResRenamed",
	}

	report, err := CompareFiles(filenamea, filenameb)
	if err != nil {
		t.Error(err)
	}

	for _, name := range names {
		deleted := false
		for _, obj := range report.Deleted {
			if obj.Name() == name {
				deleted = true
				break
			}
		}
		if deleted {
			t.Errorf("marked as deleted: %s", name)
		}
	}
}