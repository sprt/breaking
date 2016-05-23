package breaking

import "testing"

func TestDeleted(t *testing.T) {
	names := []string{
		"InterfaceMethodAdded",
		"InterfaceMethodDeleted",
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
			t.Errorf("%s not marked as deleted", obj.Name())
		}
	}
}
