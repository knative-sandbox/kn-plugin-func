## func build

Build a Function

### Synopsis


NAME
	func build - Build a Function

SYNOPSIS
	func build [-r|--registry] [--builder] [--builder-image] [--push]
	             [--palatform] [-p|--path] [-c|--confirm] [-v|--verbose]

DESCRIPTION

	Builds a function's container image and optionally pushes it to the
	configured container registry.

	By default building is handled automatically when deploying (see the deploy
	subcommand). However, sometimes it is useful to build a function container
	outside of this normal deployment process, for example for testing or during
	composition when integrationg with other systems. Additionally, the container
	can be pushed to the configured registry using the --push option.

	When building a function for the first time, either a registry or explicit
	image name is required.  Subsequent builds will reuse these option values.

EXAMPLES

	o Build a function container using the given registry.
	  The full image name will be calculated using the registry and function name.
	  $ func build --registry registry.example.com/alice

	o Build a function container using an explicit image name, ignoring registry
	  and function name.
		$ func build --image registry.example.com/alice/f:latest

	o Rebuild a function using prior values to determine container name.
	  $ func build

	o Build a function specifying the Source-to-Image (S2I) builder
	  $ func build --builder=s2i

	o Build a function specifying the Pack builder with a custom Buildpack
	  builder image.
		$ func build --builder=pack --builder-image=cnbs/sample-builder:bionic



```
func build
```

### Options

```
  -b, --builder string         build strategy to use when creating the underlying image. Currently supported build strategies are "pack" and "s2i". (default "pack")
      --builder-image string   builder image, either an as a an image name or a mapping name.
                               Specified value is stored in func.yaml (as 'builder' field) for subsequent builds. ($FUNC_BUILDER_IMAGE)
  -c, --confirm                Prompt to confirm all configuration options (Env: $FUNC_CONFIRM)
  -h, --help                   help for build
  -i, --image string           Full image name in the form [registry]/[namespace]/[name]:[tag] (optional). This option takes precedence over --registry (Env: $FUNC_IMAGE)
  -p, --path string            Path to the project directory.  Default is current working directory (Env: $FUNC_PATH)
      --platform string        Target platform to build (e.g. linux/amd64).
  -u, --push                   Attempt to push the function image after being successfully built
  -r, --registry string        Registry + namespace part of the image to build, ex 'quay.io/myuser'.  The full image name is automatically determined (Env: $FUNC_REGISTRY)
```

### Options inherited from parent commands

```
  -v, --verbose   Print verbose logs ($FUNC_VERBOSE)
```

### SEE ALSO

* [func](func.md)	 - Serverless functions

