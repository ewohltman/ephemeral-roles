package environment

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

const testVariablesFile = "testdata/variables.json"

func TestLookup(t *testing.T) {
	const testVal = "testVal"

	_, err := Lookup()
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	expected, err := expectedResults()
	if err != nil {
		t.Fatalf("Unable to obtain expected test results: %s", err)
	}

	preset := []string{
		BotToken,
		LogLevel,
	}

	setTestEnvironmentVariables(t, preset, testVal)
	defer unSetTestEnvironmentVariables(t, preset)

	actual, err := Lookup()
	if err != nil {
		t.Errorf("Error looking up environment actual: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(
			"Unexpected results:\nGot: %+v\nExpected: %+v",
			actual,
			expected,
		)
	}
}

func expectedResults() (*Variables, error) {
	jsonBytes, err := ioutil.ReadFile(testVariablesFile)
	if err != nil {
		return nil, err
	}

	variables := &Variables{}

	err = json.Unmarshal(jsonBytes, variables)
	if err != nil {
		return nil, err
	}

	return variables, nil
}

func setTestEnvironmentVariables(t *testing.T, envVars []string, testVal string) {
	for _, envVar := range envVars {
		err := os.Setenv(envVar, testVal)
		if err != nil {
			t.Errorf("Error setting test environment variable: %s", err)
		}
	}
}

func unSetTestEnvironmentVariables(t *testing.T, envVars []string) {
	for _, envVar := range envVars {
		err := os.Unsetenv(envVar)
		if err != nil {
			t.Errorf("Error unsetting test environment variable: %s", err)
		}
	}
}

func checkDefault() {

}
