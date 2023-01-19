//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"

	"testing"

	"knative.dev/func/builders"
)

const (
	arrowDown = "\033[B"
	enter     = "\r"
)

// PrepareInteractiveCommand generates a generic func that can be used to test interactive `func config` commands with user input
func PrepareInteractiveCommand(knFunc *TestShellInteractiveCmdRunner, args ...string) func(userInput ...string) {
	fn := knFunc.PrepareRun(args...)
	return func(userInput ...string) {
		result := fn(userInput...)
		if result.Error != nil {
			knFunc.T.Fatal(result.Error)
		}
	}
}

// ConfigLabelsAdd generate sa go function to test `func config labels add` with user input
func ConfigLabelsAdd(knFunc *TestShellInteractiveCmdRunner, project *FunctionTestProject) func(userInput ...string) {
	return PrepareInteractiveCommand(knFunc, "config", "labels", "add", "--path", project.ProjectPath)
}

// ConfigLabelsRemove generates a go function to test `func config labels remove` with user input
func ConfigLabelsRemove(knFunc *TestShellInteractiveCmdRunner, project *FunctionTestProject) func(userInput ...string) {
	return PrepareInteractiveCommand(knFunc, "config", "labels", "remove", "--path", project.ProjectPath)
}

// TestConfigLabel verifies function labels are properly set on the deployed functions.
// It uses "add" and "remove" sub commands with labels with specified value and labels value from environment variable.
// Test adds 3 labels and removes one.
func TestConfigLabel(t *testing.T) {

	// Given...
	labelKey1 := "l1"
	labelValue1 := "v1"
	labelKey2 := "l2"
	labelKey3 := "l3"
	testEnvName := "TEST_ENV"
	testEnvValue := "TEST_VALUE"

	knFunc := NewTestShellInteractiveCmdRunner(t)

	// On When...
	project := FunctionTestProject{}
	project.Runtime = "go"
	project.Template = "http"
	project.FunctionName = "test-config-labels"
	project.ProjectPath = filepath.Join(os.TempDir(), project.FunctionName)
	project.Builder = builders.Pack

	Create(t, knFunc.TestShell, project)
	defer func() { _ = project.RemoveProjectFolder() }()

	// Config labels add
	// Add 2 labels with specified key/value
	// Add 1 label with env
	configLabelsAdd := ConfigLabelsAdd(knFunc, &project)
	configLabelsAdd(enter, labelKey1, enter, labelValue1, enter)                   // Add first label with specified key/value
	configLabelsAdd(enter, enter, labelKey2, enter, "any", enter)                  // Add second label with specified key/value
	configLabelsAdd(enter, arrowDown, enter, labelKey3, enter, testEnvName, enter) // Add third label using value from local environment variable

	// Delete second label
	configLabelsRemove := ConfigLabelsRemove(knFunc, &project)
	configLabelsRemove(arrowDown, enter)

	// Deploy
	knFunc.TestShell.WithEnv(testEnvName, testEnvValue)
	Build(t, knFunc.TestShell, &project)
	Deploy(t, knFunc.TestShell, &project)
	defer Delete(t, knFunc.TestShell, &project)

	// Then assert that...
	// label1 exists and matches value2
	// label2 does not exists
	// label3 exists and matches value3
	resource := RetrieveKnativeServiceResource(t, project.FunctionName)
	metadataMap := resource.UnstructuredContent()["metadata"].(map[string]interface{})
	labelsMap := metadataMap["labels"].(map[string]interface{})
	if labelsMap[labelKey1] != labelValue1 {
		t.Errorf("Expected label with name %v and value %v not found", labelKey1, labelValue1)
	}
	if labelsMap[labelKey2] != nil {
		t.Errorf("Unexpected label with name %v", labelKey2)
	}
	if labelsMap[labelKey3] != testEnvValue {
		t.Errorf("Expected label with name %v and value %v not found", labelKey3, testEnvValue)
	}
}
