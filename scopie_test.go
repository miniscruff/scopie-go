package scopie

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/miniscruff/scopie-go/then"
)

type testScenario struct {
	ID        string            `json:"id"`
	Actor     string            `json:"actor"`
	Scopes    string            `json:"scopes"`
	Result    bool              `json:"result"`
	Variables map[string]string `json:"variables"`
	Error     string            `json:"error"`
}

type coreTestCase struct {
	Version     string         `json:"version"`
	Validations []testScenario `json:"validations"`
	Benchmarks  []testScenario `json:"benchmarks"`
}

func LoadScenarios(t testing.TB) coreTestCase {
	testFile, err := os.Open("testdata/scopie_scenarios.json")
	then.Nil(t, err)

	var tc coreTestCase
	err = json.NewDecoder(testFile).Decode(&tc)
	then.Nil(t, err)

	return tc
}

func Test_Validations(t *testing.T) {
	tc := LoadScenarios(t)
	for _, scenario := range tc.Validations {
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

	for _, scenario := range tc.Benchmarks {
		t.Run(scenario.ID, func(t *testing.T) {
			_, err := IsAllowed(scenario.Variables, scenario.Scopes, scenario.Actor)
			then.Nil(t, err)
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
	tc := LoadScenarios(b)
	for _, scenario := range tc.Benchmarks {
		b.Run(scenario.ID, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := IsAllowed(scenario.Variables, scenario.Scopes, scenario.Actor)
				then.Nil(b, err)
			}
		})
	}
}
