## func

Serverless functions

### Synopsis

Serverless functions v0.0.0-source-2022-09-20T08:13:19&#43;02:00

	Create, build and deploy Knative functions

SYNOPSIS
	func [-v|--verbose] <command> [args]

EXAMPLES

	o Create a Node function in the current directory
	  $ func create --language node .

	o Deploy the function defined in the current working directory to the
	  currently connected cluster, specifying a container registry in place of
	  quay.io/user for the function's container.
	  $ func deploy --registry quay.io.user

	o Invoke the function defined in the current working directory with an example
	  request.
	  $ func invoke

	For more examples, see 'func [command] --help'.

### Options

```
  -h, --help               help for func
  -n, --namespace string   The namespace on the cluster used for remote commands. By default, the namespace func.yaml is used or the currently active namespace if not set in the configuration. (Env: $FUNC_NAMESPACE)
  -v, --verbose            Print verbose logs ($FUNC_VERBOSE)
```

### SEE ALSO

* [func build](func_build.md)	 - Build a function project as a container image
* [func completion](func_completion.md)	 - Generate completion scripts for bash, fish and zsh
* [func config](func_config.md)	 - Configure a function
* [func create](func_create.md)	 - Create a function project
* [func delete](func_delete.md)	 - Undeploy a function
* [func deploy](func_deploy.md)	 - Deploy a Function
* [func info](func_info.md)	 - Show details of a function
* [func invoke](func_invoke.md)	 - Invoke a function
* [func languages](func_languages.md)	 - List available function language runtimes
* [func list](func_list.md)	 - List functions
* [func repository](func_repository.md)	 - Manage installed template repositories
* [func run](func_run.md)	 - Run the function locally
* [func templates](func_templates.md)	 - Templates
* [func version](func_version.md)	 - Show the version

