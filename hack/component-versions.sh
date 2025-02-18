#!/usr/bin/env bash

# AUTOGENERATED FILE - edit versions in ./component-versions.json.
# If you are adding components, modify this scripts' template in
# ./cmd/update-knative-components/main.go.
# You can regenerate locally with "make generate-kn-components-local".

set_versions() {
	# Note: Kubernetes Version node image per Kind releases (full hash is suggested):
	# https://github.com/kubernetes-sigs/kind/releases
	kind_node_version=v1.32.0@sha256:c48c62eac5da28cdadcf560d1d8616cfa6783b58f0d94cf63ad1bf49600cb027

	# gets updated programatically via workflow -> PR creation
	knative_serving_version="v1.17.0"
	knative_eventing_version="v1.17.0"
	contour_version="v1.17.0"
}
