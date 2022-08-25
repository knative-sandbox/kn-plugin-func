## func config envs add

Add environment variable to the function configuration

### Synopsis

Add environment variable to the function configuration

Interactive prompt to add Environment variables to the function project
in the current directory or from the directory specified with --path.

The environment variable can be set directly from a value,
from an environment variable on the local machine or from Secrets and ConfigMaps.


```
func config envs add [flags]
```

### Options

```
  -h, --help          help for add
  -p, --path string   Path to the project directory (Env: $FUNC_PATH) (default ".")
```

### Options inherited from parent commands

```
  -n, --namespace string   The namespace on the cluster used for remote commands. By default, the namespace func.yaml is used or the currently active namespace if not set in the configuration. (Env: $FUNC_NAMESPACE)
  -v, --verbose            Print verbose logs ($FUNC_VERBOSE)
```

### SEE ALSO

* [func config envs](func_config_envs.md)	 - List and manage configured environment variable for a function

