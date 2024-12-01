package scopie

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/miniscruff/scopie-go/then"
)

type testAllowedScenario struct {
	ID           string            `json:"id"`
	ActorRules   []string          `json:"actorRules"`
	ActionScopes []string          `json:"actionScopes"`
	Result       bool              `json:"result"`
	Variables    map[string]string `json:"variables"`
	Error        string            `json:"error"`
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
			res, err := IsAllowed(scenario.ActionScopes, scenario.ActorRules, scenario.Variables)
			if scenario.Error != "" {
				then.NotNil(t, err)
				then.Equals(t, scenario.Error, err.Error())
			} else {
				then.Nil(t, err)
				then.Equals(t, scenario.Result, res)
			}
		})
	}
}

func Test_IsAllowedBenchmarks(t *testing.T) {
	// Also run our benchmarks as test cases separate from running benchmarks
	for _, scenario := range testCases.Benchmarks {
		t.Run(scenario.ID, func(t *testing.T) {
			res, err := IsAllowed(scenario.ActionScopes, scenario.ActorRules, scenario.Variables)
			then.Equals(t, scenario.Result, res)
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

type compareTestCase struct {
	name   string
	actor  string
	action string
	vars   map[string]string
	err    error
	res    bool
}

func Test_CompareActorToRule(t *testing.T) {
	for _, tc := range []compareTestCase{
		{
			name:   "basic equality",
			actor:  "allow/alpha/beta",
			action: "alpha/beta",
			res:    true,
		},
		{
			name:   "first inequality",
			actor:  "allow/alpha/beta",
			action: "delta/beta",
			res:    false,
		},
		{
			name:   "last inequality",
			actor:  "allow/alpha/beta/ceti/delta",
			action: "alpha/beta/ceti/epsilon",
			res:    false,
		},
		{
			name:   "wildcard equality",
			actor:  "allow/alpha/beta/*/delta",
			action: "alpha/beta/ceti/delta",
			res:    true,
		},
		{
			name:   "super wildcard equality",
			actor:  "allow/alpha/beta/**",
			action: "alpha/beta/ceti/delta",
			res:    true,
		},
		{
			name:   "variable usage",
			actor:  "allow/alpha/@user",
			action: "alpha/our_user",
			vars: map[string]string{
				"user": "our_user",
			},
			res: true,
		},
		{
			name:   "first array value",
			actor:  "allow/alpha/beta|ceti|delta",
			action: "alpha/beta",
			res:    true,
		},
		{
			name:   "last array value",
			actor:  "allow/alpha/beta|ceti|delta",
			action: "alpha/delta", // last array value of epsilon
			res:    true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			doesMatch, err := compareActorToAction(&tc.actor, &tc.action, tc.vars)
			if tc.err == nil {
				then.Nil(t, err)
				then.Equals(t, tc.res, doesMatch)
			} else {
				then.Err(t, tc.err, err)
			}
		})
	}
}

func Benchmark_Validations(b *testing.B) {
	for _, scenario := range testCases.Benchmarks {
		b.Run(scenario.ID, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := IsAllowed(scenario.ActionScopes, scenario.ActorRules, scenario.Variables)
				then.Nil(b, err)
			}
		})
	}
}

func ExampleIsAllowed() {
	userScopes := []string{"allow/blog/create|update"}

	allowed, err := IsAllowed([]string{"blog/create"}, userScopes, nil)
	if err != nil {
		panic("invalid scopes or rules")
	}

	if !allowed {
		panic("can not create a new blog")
	}

	// create the blog here
}
