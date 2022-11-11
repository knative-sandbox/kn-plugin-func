package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ory/viper"
	"knative.dev/client/pkg/util"

	fn "knative.dev/func"
	. "knative.dev/func/testing"
)

func TestRoot_PersistentFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "not provided",
			args:     []string{"list"},
			expected: false,
		},
		{
			name:     "provided as root flags",
			args:     []string{"--verbose", "list"},
			expected: true,
		},
		{
			name:     "provided as sub-command flags",
			args:     []string{"list", "--verbose"},
			expected: true,
		},
		{
			name:     "provided as sub-sub-command flags",
			args:     []string{"repositories", "list", "--verbose"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = fromTempDirectory(t)

			cmd := NewCreateCmd(NewClient)                      // Create a function
			cmd.SetArgs([]string{"--language", "go", "myfunc"}) // providing language
			if err := cmd.Execute(); err != nil {               // fail on any errors
				t.Fatal(err)
			}

			// Assert the persistent variables were propagated to the Client constructor
			// when the command is actually invoked.
			cmd = NewRootCmd(RootCommandConfig{NewClient: func(cfg ClientConfig, _ ...fn.Option) (*fn.Client, func()) {
				if cfg.Verbose != tt.expected {
					t.Fatal("verbose persistent flag not propagated correctly")
				}
				return fn.New(), func() {}
			}})
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRoot_mergeEnvMaps(t *testing.T) {

	a := "A"
	b := "B"
	v1 := "x"
	v2 := "y"

	type args struct {
		envs     []fn.Env
		toUpdate *util.OrderedMap
		toRemove []string
	}
	tests := []struct {
		name string
		args args
		want []fn.Env
	}{
		{
			"add new var to empty list",
			args{
				[]fn.Env{},
				util.NewOrderedMapWithKVStrings([][]string{{a, v1}}),
				[]string{},
			},
			[]fn.Env{{Name: &a, Value: &v1}},
		},
		{
			"add new var",
			args{
				[]fn.Env{{Name: &b, Value: &v2}},
				util.NewOrderedMapWithKVStrings([][]string{{a, v1}}),
				[]string{},
			},
			[]fn.Env{{Name: &b, Value: &v2}, {Name: &a, Value: &v1}},
		},
		{
			"update var",
			args{
				[]fn.Env{{Name: &a, Value: &v1}},
				util.NewOrderedMapWithKVStrings([][]string{{a, v2}}),
				[]string{},
			},
			[]fn.Env{{Name: &a, Value: &v2}},
		},
		{
			"update multiple vars",
			args{
				[]fn.Env{{Name: &a, Value: &v1}, {Name: &b, Value: &v2}},
				util.NewOrderedMapWithKVStrings([][]string{{a, v2}, {b, v1}}),
				[]string{},
			},
			[]fn.Env{{Name: &a, Value: &v2}, {Name: &b, Value: &v1}},
		},
		{
			"remove var",
			args{
				[]fn.Env{{Name: &a, Value: &v1}},
				util.NewOrderedMap(),
				[]string{a},
			},
			[]fn.Env{},
		},
		{
			"remove multiple vars",
			args{
				[]fn.Env{{Name: &a, Value: &v1}, {Name: &b, Value: &v2}},
				util.NewOrderedMap(),
				[]string{a, b},
			},
			[]fn.Env{},
		},
		{
			"update and remove vars",
			args{
				[]fn.Env{{Name: &a, Value: &v1}, {Name: &b, Value: &v2}},
				util.NewOrderedMapWithKVStrings([][]string{{a, v2}}),
				[]string{b},
			},
			[]fn.Env{{Name: &a, Value: &v2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := mergeEnvs(tt.args.envs, tt.args.toUpdate, tt.args.toRemove)
			if err != nil {
				t.Errorf("mergeEnvs() for initial vars %v and toUpdate %v and toRemove %v got error %v",
					tt.args.envs, tt.args.toUpdate, tt.args.toRemove, err)
			}
			if !reflect.DeepEqual(got, tt.want) {

				gotString := "{ "
				for _, e := range got {
					gotString += fmt.Sprintf("{ %s: %s } ", *e.Name, *e.Value)
				}
				gotString += "}"

				wantString := "{ "
				for _, e := range tt.want {
					wantString += fmt.Sprintf("{ %s: %s } ", *e.Name, *e.Value)
				}
				wantString += "}"

				t.Errorf("mergeEnvs() = got: %s, want %s", gotString, wantString)
			}
		})
	}
}

// TestRoot_CommandNameParameterized confirmst that the command name, as
// printed in help text, is parameterized based on the constructor parameters
// of the root command.  This allows, for example, to have help text correct
// when both embedded as a plugin or standalone.
func TestRoot_CommandNameParameterized(t *testing.T) {
	expectedSynopsis := "%v [-v|--verbose] <command> [args]"

	tests := []string{
		"func",    // standalone
		"kn func", // kn plugin
	}

	for _, testName := range tests {
		var (
			cmd = NewRootCmd(RootCommandConfig{Name: testName})
			out = strings.Builder{}
		)
		cmd.SetArgs([]string{}) // Do not use test command args
		cmd.SetOut(&out)
		if err := cmd.Help(); err != nil {
			t.Fatal(err)
		}
		if cmd.Use != testName {
			t.Fatalf("expected command Use '%v', got '%v'", testName, cmd.Use)
		}
		if !strings.Contains(out.String(), fmt.Sprintf(expectedSynopsis, testName)) {
			t.Logf("Testing '%v'\n", testName)
			t.Log(out.String())
			t.Fatalf("Help text does not include substituted name '%v'", testName)
		}
	}
}

func TestVerbose(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "verbose as version's flag",
			args: []string{"version", "-v"},
			want: "v0.42.0-cafe-1970-01-01\n",
		},
		{
			name: "no verbose",
			args: []string{"version"},
			want: "v0.42.0\n",
		},
		{
			name: "verbose as root's flag",
			args: []string{"--verbose", "version"},
			want: "v0.42.0-cafe-1970-01-01\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			var out bytes.Buffer

			cmd := NewRootCmd(RootCommandConfig{
				Name: "func",
				Version: Version{
					Date: "1970-01-01",
					Vers: "v0.42.0",
					Hash: "cafe",
				}})

			cmd.SetArgs(tt.args)
			cmd.SetOut(&out)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			if out.String() != tt.want {
				t.Errorf("expected output: %q but got: %q", tt.want, out.String())
			}
		})
	}
}

// TestRoot_effectivePath ensures that the path method returns the effective path
// to use with the following precedence:  empty by default, then FUNC_PATH
// environment variable, -p flag, or finally --path with the highest precedence.
func TestRoot_effectivePath(t *testing.T) {

	args := os.Args
	t.Cleanup(func() { os.Args = args })

	t.Run("default", func(t *testing.T) {
		if effectivePath() != "" {
			t.Fatalf("the default path should be '.', got '%v'", effectivePath())
		}
	})

	t.Run("FUNC_PATH", func(t *testing.T) {
		t.Setenv("FUNC_PATH", "p1")
		if effectivePath() != "p1" {
			t.Fatalf("the effetive path did not load the environment variable.  Expected 'p1', got '%v'", effectivePath())
		}
	})

	t.Run("--path", func(t *testing.T) {
		os.Args = []string{"test", "--path=p2"}
		if effectivePath() != "p2" {
			t.Fatalf("the effective path did not load the --path flag.  Expected 'p2', got '%v'", effectivePath())
		}
	})

	t.Run("-p", func(t *testing.T) {
		os.Args = []string{"test", "-p=p3"}
		if effectivePath() != "p3" {
			t.Fatalf("the effective path did not load the -p flag.  Expected 'p3', got '%v'", effectivePath())
		}
	})

	t.Run("short flag precedence", func(t *testing.T) {
		t.Setenv("FUNC_PATH", "p1")
		os.Args = []string{"test", "-p=p3"}
		if effectivePath() != "p3" {
			t.Fatalf("the effective path did not load the -p flag with precedence over FUNC_PATH.  Expected 'p3', got '%v'", effectivePath())
		}
	})

	t.Run("--path highest precedence", func(t *testing.T) {
		t.Setenv("FUNC_PATH", "p1")
		os.Args = []string{"test", "--path=p2", "-p=p3"}
		if effectivePath() != "p2" {
			t.Fatalf("the effective path did not take --path with highest precedence over -p and FUNC_PATH.  Expected 'p2', got '%v'", effectivePath())
		}
	})

}

// Helpers
// -------

// pipe the output of stdout to a buffer whose value is returned
// from the returned function.  Call pipe() to start piping output
// to the buffer, call the returned function to access the data in
// the buffer.
func piped(t *testing.T) func() string {
	t.Helper()
	var (
		o = os.Stdout
		c = make(chan error, 1)
		b strings.Builder
	)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w

	go func() {
		_, err := io.Copy(&b, r)
		r.Close()
		c <- err
	}()

	return func() string {
		os.Stdout = o
		w.Close()
		err := <-c
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(b.String())
	}
}

// fromTempDirectory is a cli-specific test helper which endeavors to create
// an environment clean of developer's settings for use during CLI testing.
func fromTempDirectory(t *testing.T) string {
	t.Helper()
	// We have to define KUBECONFIG, or the file at ~/.kube/config (if extant)
	// will be used (disrupting tests by using the current user's environment).
	// The test kubeconfig set below has the current namespace set to 'func'
	// NOTE: the below settings affect unit tests only, and we do explicitly
	// want all unit tests to start in an empty environment with tests "opting in"
	// to config, not opting out.
	t.Setenv("KUBECONFIG", filepath.Join(cwd(), "testdata", "default_kubeconfig"))

	// By default unit tests presum no config exists unless provided in testdata.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// creates and CDs to a temp directory
	d, done := Mktemp(t)

	// Return to original directory and resets viper.
	t.Cleanup(func() { done(); viper.Reset() })
	return d
}
