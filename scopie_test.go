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
	Rules     string            `json:"rules"`
	Result    string            `json:"result,omitempty"`
	Variables map[string]string `json:"variables"`
}

type validationTestCase struct {
	Version   string         `json:"version"`
	Scenarios []testScenario `json:"scenarios"`
}

func LoadScenarios(t testing.TB) validationTestCase {
	testFile, err := os.Open("testdata/scopie_scenarios.json")
	then.Nil(t, err)

	var tc validationTestCase
	err = json.NewDecoder(testFile).Decode(&tc)
	then.Nil(t, err)

	return tc
}

func Test_Validations(t *testing.T) {
	tc := LoadScenarios(t)
	for _, scenario := range tc.Scenarios {
		t.Run(scenario.ID, func(t *testing.T) {
			res, err := Process(scenario.Variables, scenario.Actor, scenario.Rules)
			// TODO: handle invalid test

			then.Nil(t, err)
			then.Equals(t, scenario.Result, string(res))
		})
	}
}

func Benchmark_Validations(b *testing.B) {
	tc := LoadScenarios(b)
	for _, scenario := range tc.Scenarios {
		// TODO: only benchmark non-error results

		b.Run(scenario.ID, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := Process(scenario.Variables, scenario.Actor, scenario.Rules)
				then.Nil(b, err)
			}
		})
	}
}

/*
func Benchmark_StringAlloc(b *testing.B) {
	left := strings.Repeat("abcde", 50) + "peach"
	right := "peach"

	stringChecker := func(leftIndex int, left, right string) bool {
		for k := 0; k < len(right); k++ {
			if left[leftIndex+k] != right[k] {
				return false
			}
		}
		return true
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < len(left)-5; j++ {
			if stringChecker(j, left, right) {
				return
			}
		}
	}
}
*/
