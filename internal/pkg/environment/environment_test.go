package environment_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/environment"
)

const testVariablesFile = "testdata/variables.json"

func TestLookup(t *testing.T) {
	const testVal = "testVal"

	_, err := environment.Lookup()
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	expected, err := expectedResults()
	if err != nil {
		t.Fatalf("Unable to obtain expected test results: %s", err)
	}

	preset := []string{
		environment.BotToken,
		environment.LogLevel,
	}

	setTestEnvironmentVariables(t, preset, testVal)
	defer unSetTestEnvironmentVariables(t, preset)

	actual, err := environment.Lookup()
	if err != nil {
		t.Fatalf("Error looking up environment variables: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(
			"Unexpected results:\nGot: %+v\nExpected: %+v",
			actual,
			expected,
		)
	}
}

func expectedResults() (*environment.Variables, error) {
	jsonBytes, err := ioutil.ReadFile(testVariablesFile)
	if err != nil {
		return nil, err
	}

	variables := &environment.Variables{}

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
