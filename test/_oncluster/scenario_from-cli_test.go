//go:build oncluster
// +build oncluster

package oncluster

/*
Tests on this file covers the following scenarios:

A) Default Branch Test
func deploy --remote --git-url=http://gitserver/myfunc.git

b) Feature Branch Test
func deploy --remote --git-url=http://gitserver/myfunc.git --git-branch=feature/my-branch

c) Context Dir test
func deploy --remote --git-url=http://gitserver/myfunc.git --git-dir=functions/myfunc
*/

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	common "knative.dev/func/test/_common"
	e2e "knative.dev/func/test/_e2e"
)

// TestFromCliDefaultBranch triggers a default branch test by using CLI flags
func TestFromCliDefaultBranch(t *testing.T) {

	var gitProjectName = "test-func-yaml-build-local"
	var gitProjectPath = filepath.Join(os.TempDir(), gitProjectName)
	var funcName = gitProjectName
	var funcPath = gitProjectPath

	gitServer := common.GitTestServerProvider{}
	gitServer.Init(t)
	remoteRepo := gitServer.CreateRepository(gitProjectName)
	defer gitServer.DeleteRepository(gitProjectName)

	knFunc := common.NewKnFuncShellCli(t)
	knFunc.Exec("create", "-l", "node", funcPath)
	defer os.RemoveAll(gitProjectPath)

	GitInitialCommitAndPush(t, gitProjectPath, remoteRepo.ExternalCloneURL)

	knFunc.Exec("deploy",
		"-r", e2e.GetRegistry(),
		"-p", funcPath,
		"--remote",
		"--git-url", remoteRepo.ClusterCloneURL)

	defer knFunc.Exec("delete", "-p", funcPath)

	// ## ASSERTIONS
	result := knFunc.Exec("invoke", "-p", funcPath)
	assert.Assert(t, strings.Contains(result.Stdout, "Hello"), "Func body does not contain 'Hello'")
	AssertThatTektonPipelineRunSucceed(t, funcName)

}

// TestFromCliFeatureBranch trigger a feature branch test by using CLI flags
func TestFromCliFeatureBranch(t *testing.T) {

	var funcName = "test-func-cli-feature-branch"
	var funcPath = filepath.Join(os.TempDir(), funcName)

	gitServer := common.GitTestServerProvider{}
	gitServer.Init(t)
	remoteRepo := gitServer.CreateRepository(funcName)
	defer gitServer.DeleteRepository(funcName)

	knFunc := common.NewKnFuncShellCli(t)
	knFunc.Exec("create", "-l", "node", funcPath)
	defer os.RemoveAll(funcPath)

	GitInitialCommitAndPush(t, funcPath, remoteRepo.ExternalCloneURL)

	WriteNewSimpleIndexJS(t, funcPath, "hello branch")

	sh := common.NewShellCmd(t, funcPath)
	sh.ShouldFailOnError = true
	sh.Exec("git checkout -b feature/branch")
	sh.Exec("git add index.js")
	sh.Exec(`git commit -m "feature branch change"`)
	sh.Exec("git push -u origin feature/branch")

	knFunc.Exec("deploy",
		"-r", e2e.GetRegistry(),
		"-p", funcPath,
		"--remote",
		"--git-url", remoteRepo.ClusterCloneURL,
		"--git-branch", "feature/branch")

	defer knFunc.Exec("delete", "-p", funcPath)

	// ## ASSERTIONS
	result := knFunc.Exec("invoke", "-p", funcPath)
	assert.Assert(t, strings.Contains(result.Stdout, "hello branch"), "Func body does not contain 'hello branch'")
	AssertThatTektonPipelineRunSucceed(t, funcName)

}

// TestFromCliContextDirFunc triggers a context dir test by using CLI flags
func TestFromCliContextDirFunc(t *testing.T) {

	var gitProjectName = "test-project"
	var gitProjectPath = filepath.Join(os.TempDir(), gitProjectName)
	var funcName = "test-func-context-dir"
	var funcContextDir = filepath.Join("functions", funcName)
	var funcPath = filepath.Join(gitProjectPath, funcContextDir)

	gitServer := common.GitTestServerProvider{}
	gitServer.Init(t)
	remoteRepo := gitServer.CreateRepository(gitProjectName)
	defer gitServer.DeleteRepository(gitProjectName)

	knFunc := common.NewKnFuncShellCli(t)
	knFunc.Exec("create", "-l", "node", funcPath)
	defer os.RemoveAll(gitProjectPath)

	WriteNewSimpleIndexJS(t, funcPath, "hello dir")

	GitInitialCommitAndPush(t, gitProjectPath, remoteRepo.ExternalCloneURL)

	knFunc.Exec("deploy",
		"-r", e2e.GetRegistry(),
		"-p", funcPath,
		"--remote",
		"--git-url", remoteRepo.ClusterCloneURL,
		"--git-dir", funcContextDir)

	defer knFunc.Exec("delete", "-p", funcPath)

	// -- Assertions --
	result := knFunc.Exec("invoke", "-p", funcPath)
	assert.Assert(t, strings.Contains(result.Stdout, "hello dir"), "Func body does not contain 'hello dir'")
	AssertThatTektonPipelineRunSucceed(t, funcName)

}
