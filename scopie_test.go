package scopie

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/miniscruff/scopie-go/then"
)

type testAllowedScenario struct {
	ID        string            `json:"id"`
	Actor     string            `json:"actor"`
	Scopes    string            `json:"scopes"`
	Result    bool              `json:"result"`
	Variables map[string]string `json:"variables"`
	Error     string            `json:"error"`
}

type testValidScenario struct {
	ID    string `json:"id"`
	Scope string `json:"scope"`
	Error string `json:"error"`
}

type coreTestCase struct {
	Version         string                `json:"version"`
	IsAllowedTests  []testAllowedScenario `json:"isAllowedTests"`
	ScopeValidTests []testValidScenario   `json:"scopeValidTests"`
	Benchmarks      []testAllowedScenario `json:"benchmarks"`
}

var testCases coreTestCase

func TestMain(m *testing.M) {
	testFile, err := os.Open("testdata/scopie_scenarios.json")
	if err != nil {
		fmt.Println("unable to read scenarios", err)
		os.Exit(1)
	}

	err = json.NewDecoder(testFile).Decode(&testCases)
	if err != nil {
		fmt.Println("unable to decode scenarios", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func Test_IsAllowed(t *testing.T) {
	for _, scenario := range testCases.IsAllowedTests {
		t.Run(scenario.ID, func(t *testing.T) {
			res, err := IsAllowed(scenario.Variables, scenario.Scopes, scenario.Actor)
			if scenario.Error != "" {
				then.NotNil(t, err)
				then.Equals(t, scenario.Error, err.Error())
			} else {
				then.Nil(t, err)
				then.Equals(t, scenario.Result, res)
			}
		})
	}

	// Also run our benchmarks as test cases separate from running benchmarks
	for _, scenario := range testCases.Benchmarks {
		t.Run(scenario.ID, func(t *testing.T) {
			_, err := IsAllowed(scenario.Variables, scenario.Scopes, scenario.Actor)
			then.Nil(t, err)
		})
	}
}

func Test_ScopeValid(t *testing.T) {
	for _, scenario := range testCases.ScopeValidTests {
		t.Run(scenario.ID, func(t *testing.T) {
			err := ValidateScope(scenario.Scope)
			if scenario.Error == "" {
				then.Nil(t, err)
			} else {
				then.NotNil(t, err)
				then.Equals(t, scenario.Error, err.Error())
			}
		})
	}
}

func Test_JumpAfterSep_WhenFound(t *testing.T) {
	value := "hello_world"
	afterIndex := jumpAfterSeperator(&value, 0, '_')
	then.Equals(t, 6, afterIndex)
	then.Equals(t, 'w', value[afterIndex])
}

func Test_JumpAfterSep_WhenNotFound(t *testing.T) {
	value := "hello_world"
	afterIndex := jumpAfterSeperator(&value, 0, '$')
	then.Equals(t, 11, afterIndex)
}

func Test_JumpBlockOrScope_WhenBlock(t *testing.T) {
	value := "alpha/beta/ceti"
	afterIndex := jumpBlockOrScopeSeperator(&value, 0)
	then.Equals(t, 6, afterIndex)
}

func Test_JumpBlockOrScope_WhenScope(t *testing.T) {
	value := "alpha,beta,ceti"
	afterIndex := jumpBlockOrScopeSeperator(&value, 0)
	then.Equals(t, 6, afterIndex)
}

func Test_JumpBlockOrScope_WhenNeitherScopeOrBlock(t *testing.T) {
	value := "alphabetaceti"
	afterIndex := jumpBlockOrScopeSeperator(&value, 0)
	then.Equals(t, 13, afterIndex)
}

func Test_CompareStringsAfterIndexes_WithMatch(t *testing.T) {
	a := "allow/alpha/beta"
	b := "alpha/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, nil)
	then.True(t, doesMatch)
	then.Equals(t, 12, nextA)
	then.Equals(t, 6, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_NoMatch(t *testing.T) {
	a := "allow/alpha/beta"
	b := "centi/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, nil)
	then.False(t, doesMatch)
	then.Equals(t, 6, nextA)
	then.Equals(t, 0, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_DiffLengths(t *testing.T) {
	a := "allow/alpha/beta"
	b := "unicorn/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, nil)
	then.False(t, doesMatch)
	then.Equals(t, 6, nextA)
	then.Equals(t, 0, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_AtEnd(t *testing.T) {
	a := "allow/alpha/beta"
	b := "alpha/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 12, &b, 6, nil)
	then.True(t, doesMatch)
	then.Equals(t, len(a)+1, nextA)
	then.Equals(t, len(b)+1, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_WithWildcard(t *testing.T) {
	a := "allow/*/beta"
	b := "alpha/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, nil)
	then.True(t, doesMatch)
	then.Equals(t, 8, nextA)
	then.Equals(t, 6, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_WithArrays(t *testing.T) {
	a := "allow/alpha|beta|ceti|omega|tango/beta"
	b := "omega/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, nil)
	then.True(t, doesMatch)
	then.Equals(t, 34, nextA)
	then.Equals(t, 6, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_WithArraysInList(t *testing.T) {
	a := "allow/alpha|beta|ceti|omega|tango,allow/alpha/delta"
	b := "omega/beta"
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, nil)
	then.True(t, doesMatch)
	then.Equals(t, 34, nextA)
	then.Equals(t, 6, nextB)
	then.Nil(t, err)
}

func Test_CompareStringsAfterIndexes_WithVar(t *testing.T) {
	a := "allow/@me/beta"
	b := "omega/beta"
	vars := map[string]string{
		"me": "omega",
	}
	nextA, nextB, doesMatch, err := compareFrom(&a, 6, &b, 0, vars)
	then.True(t, doesMatch)
	then.Equals(t, 10, nextA)
	then.Equals(t, 6, nextB)
	then.Nil(t, err)
}

func Benchmark_Validations(b *testing.B) {
	for _, scenario := range testCases.Benchmarks {
		b.Run(scenario.ID, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := IsAllowed(scenario.Variables, scenario.Scopes, scenario.Actor)
				then.Nil(b, err)
			}
		})
	}
}
