#!/usr/bin/env bash
#
# Installs binaries on linux systems.
#
# Note that there are multiple 'yq's out there.  The one we want is kislyuk/yq,
# which is a thin wrapper around jq.

set -o errexit
set -o nounset
set -o pipefail

export TERM="${TERM:-dumb}"

main() {
  local kubectl_version=v1.21.1
  local kind_version=v0.11.1

  local em=$(tput bold)$(tput setaf 2)
  local me=$(tput sgr0)

  echo "${em}Fetching Binaries...${me}"

  install_kubectl
  install_kind
  install_yq

  echo "${em}DONE${me}"

}

install_kubectl() {
    echo 'Installing kubectl...'
    curl -sSLO "https://storage.googleapis.com/kubernetes-release/release/$kubectl_version/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/kubectl
    kubectl version --client=true
}

install_kind() {
    echo 'Installing kind...'
    curl -sSLo kind "https://github.com/kubernetes-sigs/kind/releases/download/$kind_version/kind-linux-amd64"
    chmod +x kind
    sudo mv kind /usr/local/bin/kind
    kind version
}

install_yq() {
    echo 'Installing yq...'
    sudo pip3 install yq
    yq --version
}


main "$@"
