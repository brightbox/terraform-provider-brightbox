package brightbox

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gotest.tools/v3/assert"
)

func TestTimeFromFloat(t *testing.T) {
	ts := "1564670787.9459848"
	tf, _ := strconv.ParseFloat(ts, 64)
	checkString := fmt.Sprintf("%.9f", tf)
	checkTime := timeFromFloat(tf)
	resultString := fmt.Sprintf("%d.%d", checkTime.Unix(), checkTime.Nanosecond())
	if checkString != resultString {
		t.Errorf("Time not converted properly, expected %s, got %s", checkString, resultString)
	}

}

func TestValidateCron(t *testing.T) {
	testCases := []StringValidationTestCase{
		{"Valid Cron", "5 4 * * *", false},
		{"Invalid Cron", "apple", true},
	}
	es := testStringValidationCases(testCases, ValidateCronString)
	if len(es) > 0 {
		t.Errorf("Failed to validate cron: %v", es)
	}
}

func TestValidateKeys(t *testing.T) {
	testCases := []StringMapValidationTestCase{
		{
			TestName: "Valid keys",
			Value: map[string]interface{}{
				"foo":     "bar",
				"foo-bar": "foobar",
				"bar1^":   "ok",
			},
		},
		{
			TestName: "Invalid keys",
			Value: map[string]interface{}{
				"foo_bar": "bar",
				"foo bar": "bar",
				"€uro":    "bar",
				"Foo":     "Bar",
				"Foo-Bar": "foobar",
			},
			ExpectError: true,
		},
	}
	es := testStringMapValidationCases(testCases, http1Keys)
	if len(es) > 0 {
		t.Errorf("Failed to validate keys: %v", es)
	}
}

type StringMapValidationTestCase struct {
	TestName    string
	Value       map[string]interface{}
	ExpectError bool
}

type StringValidationTestCase struct {
	TestName    string
	Value       string
	ExpectError bool
}

func testStringMapValidationCases(cases []StringMapValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	es := make([]error, 0)
	for _, c := range cases {
		es = append(es, testStringMapValidation(c, validationFunc)...)
	}

	return es
}

func testStringValidationCases(cases []StringValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	es := make([]error, 0)
	for _, c := range cases {
		es = append(es, testStringValidation(c, validationFunc)...)
	}

	return es
}

func testStringMapValidation(testCase StringMapValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	_, es := validationFunc(testCase.Value, testCase.TestName)
	if testCase.ExpectError {
		if len(es) > 0 {
			return nil
		}
		return []error{fmt.Errorf("Didn't see expected error in case \"%s\" with string \"%s\"", testCase.TestName, testCase.Value)}
	}

	return es
}

func testStringValidation(testCase StringValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	_, es := validationFunc(testCase.Value, testCase.TestName)
	if testCase.ExpectError {
		if len(es) > 0 {
			return nil
		}
		return []error{fmt.Errorf("Didn't see expected error in case \"%s\" with string \"%s\"", testCase.TestName, testCase.Value)}
	}

	return es
}

var testNameRe = regexp.MustCompile(`^foo-\d+|^bar-\d+|^baz-\d+|^initial$`)

func isTestName(name string) bool {
	return testNameRe.MatchString(name)
}

func TestFilter(t *testing.T) {
	var testData []int
	for i := 1; i < 100; i++ {
		testData = append(testData, i)
	}

	output := filter(testData, func(v int) bool {
		return v >= 30 && v <= 33
	})

	assert.DeepEqual(t, []int{30, 31, 32, 33}, output)
}
