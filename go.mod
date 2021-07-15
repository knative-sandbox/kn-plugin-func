module knative.dev/kn-plugin-func

go 1.15

require (
	github.com/AlecAivazis/survey/v2 v2.2.12
	github.com/buildpacks/pack v0.18.0
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/containers/image/v5 v5.10.5
	github.com/docker/docker v20.10.2+incompatible
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/docker/go-connections v0.4.0
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/markbates/pkger v0.17.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6
	github.com/ory/viper v1.7.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	knative.dev/client v0.22.0
	knative.dev/eventing v0.22.0
	knative.dev/pkg v0.0.0-20210331065221-952fdd90dbb0
	knative.dev/serving v0.22.0
)

// knative.dev/serving@v0.21.0 and knative.dev/pkg@v0.0.0-20210331065221-952fdd90dbb0 require different versions of go-openapi/spec
replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.6
