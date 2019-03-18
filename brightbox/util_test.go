package brightbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
)

func TestValidateKeys(t *testing.T) {
	testCases := []StringValidationTestCase{
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
	es := testStringValidationCases(testCases, http1Keys)
	if len(es) > 0 {
		t.Errorf("Failed to validate keys: %v", es)
	}
}

type StringValidationTestCase struct {
	TestName    string
	Value       map[string]interface{}
	ExpectError bool
}

func testStringValidationCases(cases []StringValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	es := make([]error, 0)
	for _, c := range cases {
		es = append(es, testStringValidation(c, validationFunc)...)
	}

	return es
}

func testStringValidation(testCase StringValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	_, es := validationFunc(testCase.Value, testCase.TestName)
	if testCase.ExpectError {
		if len(es) > 0 {
			return nil
		} else {
			return []error{fmt.Errorf("Didn't see expected error in case \"%s\" with string \"%s\"", testCase.TestName, testCase.Value)}
		}
	}

	return es
}
