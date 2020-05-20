package brightbox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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
