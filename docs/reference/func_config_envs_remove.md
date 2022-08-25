## func config envs remove

Remove environment variable from the function configuration

### Synopsis

Remove environment variable from the function configuration

Interactive prompt to remove Environment variables from the function project
in the current directory or from the directory specified with --path.


```
func config envs remove [flags]
```

### Options

```
  -h, --help          help for remove
  -p, --path string   Path to the project directory (Env: $FUNC_PATH) (default ".")
```

### Options inherited from parent commands

```
  -n, --namespace string   The namespace on the cluster used for remote commands. By default, the namespace func.yaml is used or the currently active namespace if not set in the configuration. (Env: $FUNC_NAMESPACE)
  -v, --verbose            Print verbose logs ($FUNC_VERBOSE)
```

### SEE ALSO

* [func config envs](func_config_envs.md)	 - List and manage configured environment variable for a function

