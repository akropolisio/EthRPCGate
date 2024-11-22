package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type MochaSpecStats struct {
	Suites   int `json:"suites,omitempty"`
	Tests    int `json:"tests,omitempty"`
	Passes   int `json:"passes,omitempty"`
	Pending  int `json:"pending,omitempty"`
	Failures int `json:"failures,omitempty"`
	Flakes   int `json:"flakes,omitempty"`
}

type MochaSpecReceipt struct {
	TransactionHash   string        `json:"transactionHash,omitempty"`
	TransactionIndex  int           `json:"transactionIndex,omitempty"`
	BlockHash         string        `json:"blockHash,omitempty"`
	BlockNumber       int           `json:"blockNumber,omitempty"`
	From              string        `json:"from,omitempty"`
	To                string        `json:"to,omitempty"`
	CumulativeGasUsed big.Int       `json:"cumulativeGasUsed,omitempty"`
	GasUsed           big.Int       `json:"gasUsed,omitempty"`
	Logs              []interface{} `json:"logs,omitempty"`
	LogsBloom         string        `json:"logsBloom,omitempty"`
	Status            bool          `json:"status,omitempty"`
	RawLogs           []interface{} `json:"rawLogs,omitempty"`
}

type MochaSpecFailureError struct {
	Stack         string           `json:"stack,omitempty"`
	Message       string           `json:"message,omitempty"`
	Name          string           `json:"name,omitempty"`
	Transaction   string           `json:"tx,omitempty"`
	Receipt       MochaSpecReceipt `json:"receipt,omitempty"`
	Reason        string           `json:"reason,omitempty"`
	HijackedStack string           `json:"hijackedStack,omitempty"`
	ShowDiff      bool             `json:"showDiff,omitempty"`
	Actual        interface{}      `json:"actual,omitempty"`
	Expected      interface{}      `json:"expected,omitempty"`
	Operator      string           `json:"operator,omitempty"`
}

type MochaSpecFailure struct {
	FullTitle string                 `json:"fullTitle,omitempty"`
	Title     string                 `json:"title,omitempty"`
	Duration  int                    `json:"duration,omitempty"`
	Result    string                 `json:"result,omitempty"`
	Error     *MochaSpecFailureError `json:"err,omitempty"`
}

func (failure *MochaSpecFailure) String() string {
	return fmt.Sprintf("%s/%s", failure.FullTitle, failure.Title)
}

type MochaSpecJsonOutput struct {
	Stats    MochaSpecStats    `json:"stats,omitempty"`
	Failures MochaSpecFailures `json:"failures,omitempty"`
	Passes   MochaSpecFailures `json:"passes,omitempty"`
	Flakes   MochaSpecFailures `json:"flakes,omitempty"`
}

type MochaSpecFailures []MochaSpecFailure

func (a MochaSpecFailures) Len() int           { return len(a) }
func (a MochaSpecFailures) Less(i, j int) bool { return a[i].String() < a[j].String() }
func (a MochaSpecFailures) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

var (
	app = kingpin.New("truffleParser", "Parse Truffle JSON test output")

	expected = app.Flag("expected", "expected results").Envar("EXPECTED").File()
	input    = app.Flag("input", "json file to parse").Envar("INPUT").File()
	output   = app.Flag("output", "json file to parse").Envar("OUTPUT").String()
)

func action(pc *kingpin.ParseContext) error {
	if expected == nil {
		return errors.New("--expected parameter required")
	}

	if input == nil {
		return errors.New("--input parameter required")
	}

	if output == nil {
		return errors.New("--output parameter required")
	}

	expectedBytes, err := ioutil.ReadAll(*expected)
	if err != nil {
		return errors.Wrap(err, "Failed to read --expected file contents")
	}

	inputBytes, err := ioutil.ReadAll(*input)
	if err != nil {
		return errors.Wrap(err, "Failed to read --input file contents")
	}

	var unmarshalledExpected MochaSpecJsonOutput
	var unmarshalledInput MochaSpecJsonOutput

	err = json.Unmarshal(expectedBytes, &unmarshalledExpected)
	if err != nil {
		return errors.Wrap(err, "Failed to parse --expected file contents")
	}

	err = json.Unmarshal(inputBytes, &unmarshalledInput)
	if err != nil {
		return errors.Wrap(err, "Failed to parse --input file contents")
	}

	sort.Sort(unmarshalledExpected.Failures)
	sort.Sort(unmarshalledExpected.Passes)
	sort.Sort(unmarshalledExpected.Flakes)
	sort.Sort(unmarshalledInput.Failures)
	sort.Sort(unmarshalledInput.Passes)

	errs := compareReports(unmarshalledExpected, unmarshalledInput)

	prune(unmarshalledExpected.Failures)
	prune(unmarshalledExpected.Passes)
	prune(unmarshalledInput.Failures)
	prune(unmarshalledInput.Passes)
	unmarshalledInput.Flakes = unmarshalledExpected.Flakes[:]

	prunedInput, err := json.MarshalIndent(unmarshalledInput, "", " ")
	if err != nil {
		panic(err)
	}

	if len(errs) != 0 {
		fmt.Println("Update the the input file with this contents for this to pass if these tests are expected")
		fmt.Println("=======================================")
		fmt.Println(string(prunedInput))
		fmt.Println("=======================================")
	}

	if output != nil {
		err = ioutil.WriteFile(*output, prunedInput, 0777)
		if err != nil {
			return errors.Wrap(err, "Failed to write pruned --input to --output")
		}
		fmt.Printf("Wrote expected output to %s\n", *output)
	}

	if len(errs) == 0 {
		fmt.Println("Files match")
		os.Exit(0)
	}

	fmt.Printf("%d errors occurred comparing test results\n", len(errs))
	for i, err := range errs {
		fmt.Printf("%d) %s\n", i+1, err)
	}

	os.Exit(len(errs))
	return nil
}

func compareReports(expected MochaSpecJsonOutput, got MochaSpecJsonOutput) []error {
	errs := []error{}

	// flakes can end up in passes/failures
	// so if expected has any flakes
	// we need to detect if they are in got.Passes or got.Failures
	// then decrement the passes/failure count by those flakes
	gotFailedFlakes := getMultiple(expected.Flakes, got.Failures)
	gotPassedFlakes := getMultiple(expected.Flakes, got.Passes)

	failures := pruneMultiple(gotFailedFlakes, got.Failures)
	passes := pruneMultiple(gotPassedFlakes, got.Passes)

	got.Stats.Failures -= len(got.Failures) - len(failures)
	got.Stats.Passes -= len(got.Passes) - len(passes)

	got.Failures = failures
	got.Passes = passes

	if expected.Stats.Tests != got.Stats.Tests {
		errs = append(errs, errors.Errorf("Total tests ran don't match: expected: %d got: %d", expected.Stats.Tests, got.Stats.Tests))
	}
	if expected.Stats.Passes != got.Stats.Passes {
		errs = append(errs, errors.Errorf("Total test passes ran don't match: expected: %d got: %d", expected.Stats.Passes, got.Stats.Passes))
	}
	if expected.Stats.Pending != got.Stats.Pending {
		errs = append(errs, errors.Errorf("Total tests pending ran don't match: expected: %d got: %d", expected.Stats.Pending, got.Stats.Pending))
	}
	if expected.Stats.Failures != got.Stats.Failures {
		errs = append(errs, errors.Errorf("Total test failures ran don't match: expected: %d got: %d", expected.Stats.Failures, got.Stats.Failures))
	}

	errs = append(errs, compare("Unexpected test failure in output", got.Failures, expected.Failures)...)
	errs = append(errs, compare("Expected test failure missing in output", expected.Failures, got.Failures)...)
	errs = append(errs, compare("Unexpected test skipped in output", got.Passes, expected.Passes)...)
	errs = append(errs, compare("Expected skipped test missing in output", expected.Passes, got.Passes)...)

	if len(errs) == 0 {
		// tests should be sorted in both inputs
		// make sure that tests are exactly the same as there can be duplicates
		expectIdentical(expected.Failures, got.Failures)
		expectIdentical(expected.Passes, got.Passes)
	}

	return errs
}

func compare(message string, left MochaSpecFailures, right MochaSpecFailures) []error {
	errs := []error{}
	for _, test := range left {
		expectedFailure := get(test, right)
		if expectedFailure == nil {
			byts, err := json.MarshalIndent(test, "", " ")
			if err != nil {
				errs = append(errs, errors.Wrap(err, "Failed to marshal to json"))
			}
			errs = append(errs, errors.Errorf("%s:\n%s", message, string(byts)))
		}
	}
	return errs
}

func expectIdentical(left MochaSpecFailures, right MochaSpecFailures) []error {
	var errs []error
	for i := range right {
		rightTest := right[i]
		leftTest := left[i]
		if leftTest.String() != rightTest.String() {
			errs = append(errs, errors.Errorf("Index %d doesn't match '%s' != '%s'", i, leftTest.String(), rightTest.String()))
		}
	}
	return errs
}

func getMultiple(flakes MochaSpecFailures, got MochaSpecFailures) MochaSpecFailures {
	results := MochaSpecFailures{}

	for _, flake := range flakes {
		expectedFlake := get(flake, got)
		if expectedFlake != nil {
			results = append(results, *expectedFlake)
		}
	}

	return results
}

func get(failure MochaSpecFailure, from MochaSpecFailures) *MochaSpecFailure {
	// there will be < 2000 tests here, o^n complexity is fine
	for _, fromFailure := range from {
		if fromFailure.String() == failure.String() {
			return &fromFailure
		}
	}

	return nil
}

func prune(specs MochaSpecFailures) {
	for i := range specs {
		specs[i].Result = ""
		specs[i].Error = nil
		specs[i].Duration = 0
	}
}

func pruneMultiple(failuresToPrune MochaSpecFailures, pruneTarget MochaSpecFailures) MochaSpecFailures {
	result := MochaSpecFailures{}

	failuresToPruneMap := make(map[string]MochaSpecFailure)
	for _, failure := range failuresToPrune {
		failuresToPruneMap[failure.String()] = failure
	}

	for _, target := range pruneTarget {
		_, ok := failuresToPruneMap[target.String()]
		if !ok {
			result = append(result, target)
		}
	}

	return result
}

func Run() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func init() {
	app.Action(action)
}

func main() {
	Run()
}
