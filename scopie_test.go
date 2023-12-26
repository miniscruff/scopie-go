package scopie

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/neilotoole/slogt"

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
	slogDef := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(slogDef)
	})

	testLog := slogt.New(t, slogt.JSON())
	slog.SetDefault(testLog)

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
