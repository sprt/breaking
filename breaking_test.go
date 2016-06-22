package breaking

import "testing"

const (
	dira = "testdata/a"
	dirb = "testdata/b"
)

func TestObjectDiffNew(t *testing.T) {
	d := &ObjectDiff{nil, &Object{obj: nil}}
	n := d.New()
	if n != nil {
		t.Error("d.New(): expected nil, got", n)
	}
}

func TestBreaking(t *testing.T) {
	names := []string{
		"FuncParameterAdded",
		"FuncParamTypeChanged",
		"FuncResAdded",
		"FuncRetTypeChanged",
		"InterfaceMethodAdded",
		"InterfaceMethodParameterAdded",
		"InterfaceMethParamTypeChanged",
		"InterfaceMethRetTypeChanged",
		"StructExportedAddedUnexported",
		"StructExportedPrependedExported",
		"StructExportedRemoved",
		"StructExportedRepositioned",
		"StructExportedTypeChanged",
		"StructFieldRenamed",
		"StructMixedExportedRemoved",
		"TypeStructToVar",
		"VarDeleted",
		"VarToFunc",
		"VarTypeChanged",
	}

	diffs, err := ComparePackages(dira, dirb)
	if err != nil {
		t.Error(err)
	}

	for _, name := range names {
		deleted := false
		for _, obj := range diffs {
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

func TestNonBreaking(t *testing.T) {
	names := []string{
		"FuncParamRenamed",
		"FuncResRenamed",
		"InterfaceMethodDeleted",
		"NamedType",
		"StructEmptyAddedExported",
		"StructEmptyAddedUnexported",
		"StructExportedAppendedExported",
		"StructMixedAddedExported",
		"StructMixedRepositionedExported",
		"StructMixedRepositionedUnexported",
		"StructNamedType",
		"StructRecursive",
		"StructUnexportedAddedExported",
		"StructUnexportedAddedUnexported",
		"StructUnexportedRemoved",
		"StructUnexportedRepositioned",
	}

	diffs, err := ComparePackages(dira, dirb)
	if err != nil {
		t.Error(err)
	}

	for _, name := range names {
		deleted := false
		for _, obj := range diffs {
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
