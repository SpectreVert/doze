package doze

// We probably don't need so many tests for this part of the code
// but this is my first test file so I wanted to hit the 100%
// coverage and test decently the capacities of the code.
// I will keep trying to test all the code as I think this may
// force me to architecture it more robustly.

import (
	"errors"
	"testing"
)

func TestProcedureID(t *testing.T) {
	for i, tc := range []struct {
		input           ProcedureID
		expectNamespace string
		expectName      string
	}{
		{
			input:           "bar",
			expectNamespace: "",
			expectName:      "bar",
		},
		{
			input:           "bar:boz",
			expectNamespace: "bar",
			expectName:      "boz",
		},
		{
			input:           "bar:boz:big:ben",
			expectNamespace: "bar:boz:big",
			expectName:      "ben",
		},
		{
			input:           "",
			expectNamespace: "",
			expectName:      "",
		},
	} {
		actualNamespace := tc.input.Namespace()
		if actualNamespace != tc.expectNamespace {
			t.Errorf("case %d: expected namespace %s but got %s", i, tc.expectNamespace, actualNamespace)
		}
		actualName := tc.input.Name()
		if actualName != tc.expectName {
			t.Errorf("case %d: expected name %s but got %s", i, tc.expectName, actualName)
		}
	}
}

type TestProcedure struct{}

func (TestProcedure) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:foo:bar",
		New: func() Procedure { return new(TestProcedure) },
	}
}

func TestRegisterProcedure(t *testing.T) {
	for i, tc := range []struct {
		input            Procedure
		expectProcInfoID string
	}{
		{
			input:            TestProcedure{},
			expectProcInfoID: "test:foo:bar",
		},
	} {
		RegisterProcedure(tc.input)

		procsMutex.Lock()
		defer procsMutex.Unlock()

		if actualProcInfo, ok := procedures[tc.expectProcInfoID]; !ok {
			t.Errorf("case %d: expected to find procedure %s in procedures map but got %s", i, tc.expectProcInfoID, actualProcInfo.ID)
		}
	}
}

// These tests are testing that we panic in case of error during procedure registration.
// We can only panic once per test case so all cases have to be separated. Sigh.

// Also, all procedures are registered in the same test runtime namespace, so we cannot
// register one instance of a procedure multiple times, otherwise that panic will shadow
// the actual panic we are testing for.

func assertRegisterProcedurePanic(t *testing.T, registerF func(Procedure), instance Procedure, expectPanicMsg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected to panic but did not")
		} else {
			switch x := r.(type) {
			case string:
				if x != expectPanicMsg {
					t.Errorf("expected to panic with error %s but got %s", expectPanicMsg, x)
				}
			default:
				t.Errorf("expected to panic with error %s but got an unknown panic type", expectPanicMsg)
			}
		}
	}()
	registerF(instance)
}

type TestProcedureCopyCat struct{}

func (TestProcedureCopyCat) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:foo:bar",
		New: func() Procedure { return new(TestProcedureCopyCat) },
	}
}

func TestRegisterCopyCatProcedure(t *testing.T) {
	for _, tc := range []struct {
		base           Procedure
		input          Procedure
		expectPanicMsg string
	}{
		{
			base:           TestProcedure{},
			input:          TestProcedureCopyCat{},
			expectPanicMsg: "procedure already registered: test:foo:bar",
		},
	} {
		assertRegisterProcedurePanic(t, RegisterProcedure, tc.input, tc.expectPanicMsg)
	}
}

type TestProcedureMissingID struct{}

func (TestProcedureMissingID) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "",
		New: func() Procedure { return new(TestProcedureMissingID) },
	}
}

func TestRegisterMissingIDProcedure(t *testing.T) {
	for _, tc := range []struct {
		input          Procedure
		expectPanicMsg string
	}{
		{
			input:          TestProcedureMissingID{},
			expectPanicMsg: "procedure ID missing",
		},
	} {
		assertRegisterProcedurePanic(t, RegisterProcedure, tc.input, tc.expectPanicMsg)
	}
}

type TestProcedureMissingNew struct{}

func (TestProcedureMissingNew) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID: "test:foo:boz",
	}
}

func TestRegisterMissingNew(t *testing.T) {
	for _, tc := range []struct {
		input          Procedure
		expectPanicMsg string
	}{
		{
			input:          TestProcedureMissingNew{},
			expectPanicMsg: "missing ProcedureInfo.New",
		},
	} {
		assertRegisterProcedurePanic(t, RegisterProcedure, tc.input, tc.expectPanicMsg)
	}
}

type TestProcedureMissingInstance struct{}

func (TestProcedureMissingInstance) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:foo:bar:boz",
		New: func() Procedure { return nil },
	}
}

func TestRegisterMissingInstance(t *testing.T) {
	for _, tc := range []struct {
		input          Procedure
		expectPanicMsg string
	}{
		{
			input:          TestProcedureMissingInstance{},
			expectPanicMsg: "ProcedureInfo.New must return a non-nil procedure instance",
		},
	} {
		assertRegisterProcedurePanic(t, RegisterProcedure, tc.input, tc.expectPanicMsg)
	}
}

// Hang on, for the next test we need a bunch of testing structures...

type TestProcedureTccObjectFile struct{}

func (TestProcedureTccObjectFile) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:lang:c:tcc:objectFile",
		New: func() Procedure { return new(TestProcedureTccObjectFile) },
	}
}

type TestProcedureTccExecutable struct{}

func (TestProcedureTccExecutable) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:lang:c:tcc:executable",
		New: func() Procedure { return new(TestProcedureTccExecutable) },
	}
}

type TestProcedureGccObjectFile struct{}

func (TestProcedureGccObjectFile) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:lang:c:gcc:objectFile",
		New: func() Procedure { return new(TestProcedureGccObjectFile) },
	}
}

type TestProcedureGccExecutable struct{}

func (TestProcedureGccExecutable) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:lang:c:gcc:executable",
		New: func() Procedure { return new(TestProcedureGccExecutable) },
	}
}

type TestProcedureGoBuild struct{}

func (TestProcedureGoBuild) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:lang:go:build",
		New: func() Procedure { return new(TestProcedureGoBuild) },
	}
}

type TestProcedureZigCompile struct{}

func (TestProcedureZigCompile) DozeProcedure() ProcedureInfo {
	return ProcedureInfo{
		ID:  "test:lang:zig:compile",
		New: func() Procedure { return new(TestProcedureZigCompile) },
	}
}

func init() {
	// The test is not about registering procedures so we do it one time at the start.
	RegisterProcedure(TestProcedureTccObjectFile{})
	RegisterProcedure(TestProcedureTccExecutable{})
	RegisterProcedure(TestProcedureGccObjectFile{})
	RegisterProcedure(TestProcedureGccExecutable{})
	RegisterProcedure(TestProcedureZigCompile{})
	RegisterProcedure(TestProcedureGoBuild{})
}

// These tests are fragile and probably easily broken. First, what happens when more modules
// are added into the default configuration? I will try to improve them later.
func TestGetProcedures(t *testing.T) {
	for caseNb, tc := range []struct {
		input         string
		expectProcIDs []string
	}{
		{
			input:         "",
			expectProcIDs: []string{"test:foo:bar", "test:lang:c:gcc:executable", "test:lang:c:gcc:objectFile", "test:lang:c:tcc:executable", "test:lang:c:tcc:objectFile", "test:lang:go:build", "test:lang:zig:compile"},
		},
		{
			input:         "test:lang:c",
			expectProcIDs: []string{"test:lang:c:gcc:executable", "test:lang:c:gcc:objectFile", "test:lang:c:tcc:executable", "test:lang:c:tcc:objectFile"},
		},
		{
			input:         "test:lang:c:gcc:executable:not_exist",
			expectProcIDs: []string{},
		},
	} {
		procs := GetProcedures(tc.input)
		if len(procs) != len(tc.expectProcIDs) {
			t.Errorf("case %d: expected %d procedures but got %d", caseNb, len(tc.expectProcIDs), len(procs))
		}

		for i, actualProcInfo := range procs {
			if tc.expectProcIDs[i] != string(actualProcInfo.ID) {
				t.Errorf("case %d: expected procedure %s but got %s", caseNb, tc.expectProcIDs[i], actualProcInfo.ID)
			}
		}
	}
}

func TestGetOneProcedure(t *testing.T) {
	for caseNb, tc := range []struct {
		input        string
		expectErr    error
		expectProcID ProcedureID
	}{
		{
			input:        "test:not_exist",
			expectErr:    ProcedureError{"test:not_exist"},
			expectProcID: "",
		},
		{
			input:        "test:lang:go:build",
			expectErr:    nil,
			expectProcID: "test:lang:go:build",
		},
	} {
		procInfo, actualErr := GetProcedure(tc.input)
		if actualErr != nil {
			if !errors.Is(actualErr, tc.expectErr) {
				t.Errorf("case %d: expected error\n %v`\nbut got\n %v`\ninstead", caseNb, tc.expectErr, actualErr)
			}
		}
		actualProcID := procInfo.ID
		if tc.expectProcID != actualProcID {
			t.Errorf("case %d: expected ProcID %s but got %s instead", caseNb, tc.expectProcID, actualProcID)
		}
	}
}
