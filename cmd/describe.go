package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/ory/viper"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	fn "knative.dev/kn-plugin-func"
	"knative.dev/kn-plugin-func/knative"
)

func init() {
	root.AddCommand(NewDescribeCmd(newDescribeClient))
}

func newDescribeClient(cfg describeConfig) (*fn.Client, error) {
	describer, err := knative.NewDescriber(cfg.Namespace)
	if err != nil {
		return nil, err
	}

	describer.Verbose = cfg.Verbose

	return fn.New(
		fn.WithDescriber(describer),
		fn.WithVerbose(cfg.Verbose),
	), nil
}

type describeClientFn func(describeConfig) (*fn.Client, error)

func NewDescribeCmd(clientFn describeClientFn) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Show details of a function",
		Long: `Show details of a function

Prints the name, route and any event subscriptions for a deployed function in
the current directory or from the directory specified with --path.
`,
		Example: `
# Show the details of a function as declared in the local func.yaml
kn func describe

# Show the details of the function in the myotherfunc directory with yaml output
kn func describe --output yaml --path myotherfunc
`,
		SuggestFor:        []string{"desc", "get"},
		ValidArgsFunction: CompleteFunctionList,
		PreRunE:           bindEnv("namespace", "output", "path"),
	}

	cmd.Flags().StringP("namespace", "n", "", "Namespace of the function. By default, the namespace in func.yaml is used or the actual active namespace if not set in the configuration. (Env: $FUNC_NAMESPACE)")
	cmd.Flags().StringP("output", "o", "human", "Output format (human|plain|json|xml|yaml|url) (Env: $FUNC_OUTPUT)")
	cmd.Flags().StringP("path", "p", cwd(), "Path to the project directory (Env: $FUNC_PATH)")

	if err := cmd.RegisterFlagCompletionFunc("output", CompleteOutputFormatList); err != nil {
		fmt.Println("internal: error while calling RegisterFlagCompletionFunc: ", err)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runDescribe(cmd, args, clientFn)
	}

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string, clientFn describeClientFn) (err error) {
	config := newDescribeConfig(args)

	function, err := fn.NewFunction(config.Path)
	if err != nil {
		return
	}

	// Check if the Function has been initialized
	if !function.Initialized() {
		return fmt.Errorf("the given path '%v' does not contain an initialized function", config.Path)
	}

	// Create a client
	client, err := clientFn(config)
	if err != nil {
		return err
	}

	// Get the description
	d, err := client.Describe(cmd.Context(), config.Name, config.Path)
	if err != nil {
		return
	}
	d.Image = function.Image

	write(os.Stdout, description(d), config.Output)
	return
}

// CLI Configuration (parameters)
// ------------------------------

type describeConfig struct {
	Name      string
	Namespace string
	Output    string
	Path      string
	Verbose   bool
}

func newDescribeConfig(args []string) describeConfig {
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	return describeConfig{
		Name:      deriveName(name, viper.GetString("path")),
		Namespace: viper.GetString("namespace"),
		Output:    viper.GetString("output"),
		Path:      viper.GetString("path"),
		Verbose:   viper.GetBool("verbose"),
	}
}

// Output Formatting (serializers)
// -------------------------------

type description fn.Description

func (d description) Human(w io.Writer) error {
	fmt.Fprintln(w, "Function name:")
	fmt.Fprintf(w, "  %v\n", d.Name)
	fmt.Fprintln(w, "Function is built in image:")
	fmt.Fprintf(w, "  %v\n", d.Image)
	fmt.Fprintln(w, "Function is deployed in namespace:")
	fmt.Fprintf(w, "  %v\n", d.Namespace)
	fmt.Fprintln(w, "Routes:")

	for _, route := range d.Routes {
		fmt.Fprintf(w, "  %v\n", route)
	}

	if len(d.Subscriptions) > 0 {
		fmt.Fprintln(w, "Subscriptions (Source, Type, Broker):")
		for _, s := range d.Subscriptions {
			fmt.Fprintf(w, "  %v %v %v\n", s.Source, s.Type, s.Broker)
		}
	}
	return nil
}

func (d description) Plain(w io.Writer) error {
	fmt.Fprintf(w, "Name %v\n", d.Name)
	fmt.Fprintf(w, "Image %v\n", d.Image)
	fmt.Fprintf(w, "Namespace %v\n", d.Namespace)

	for _, route := range d.Routes {
		fmt.Fprintf(w, "Route %v\n", route)
	}

	if len(d.Subscriptions) > 0 {
		for _, s := range d.Subscriptions {
			fmt.Fprintf(w, "Subscription %v %v %v\n", s.Source, s.Type, s.Broker)
		}
	}
	return nil
}

func (d description) JSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(d)
}

func (d description) XML(w io.Writer) error {
	return xml.NewEncoder(w).Encode(d)
}

func (d description) YAML(w io.Writer) error {
	return yaml.NewEncoder(w).Encode(d)
}

func (d description) URL(w io.Writer) error {
	if len(d.Routes) > 0 {
		fmt.Fprintf(w, "%s\n", d.Routes[0])
	}
	return nil
}
