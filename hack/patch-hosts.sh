#!/usr/bin/env bash

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Create DNS A records for '127.0.0.1.sslip.io' and '*.127.0.0.1.sslip.io' pointing to the cluster node.
#

source "$(dirname "$(realpath "$0")")/common.sh"

function patch_hosts() {
  echo "${blue}Configuring Magic DNS${reset}"

  local cluster_node_addr

  cluster_node_addr="$(docker container inspect func-control-plane | jq ".[0].NetworkSettings.Networks.kind.IPAddress" -r)"

  $KUBECTL patch cm/coredns -n kube-system --patch-file /dev/stdin <<EOF
{
  "data": {
    "Corefile": ".:53 {\n    errors\n    health {\n       lameduck 5s\n    }\n    ready\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n       pods insecure\n       fallthrough in-addr.arpa ip6.arpa\n       ttl 30\n    }\n    file /etc/coredns/example.db 127.0.0.1.sslip.io\n    prometheus :9153\n    forward . /etc/resolv.conf {\n       max_concurrent 1000\n    }\n    cache 30\n    loop\n    reload\n    loadbalance\n}\n",
    "example.db": "; 127.0.0.1.sslip.io test file\n127.0.0.1.sslip.io.            IN      SOA     sns.dns.icann.org. noc.dns.icann.org. 2015082541 7200 3600 1209600 3600\n127.0.0.1.sslip.io.            IN      A       ${cluster_node_addr}\n*.127.0.0.1.sslip.io.          IN      A       ${cluster_node_addr}\n"
  }
}
EOF

  $KUBECTL patch deploy/coredns -n kube-system --patch-file /dev/stdin <<EOF
{
  "spec": {
    "template": {
      "spec": {
        "\$setElementOrder/volumes": [
          {
            "name": "config-volume"
          }
        ],
        "volumes": [
          {
            "\$retainKeys": [
              "configMap",
              "name"
            ],
            "configMap": {
              "items": [
                {
                  "key": "Corefile",
                  "path": "Corefile"
                },
                {
                  "key": "example.db",
                  "path": "example.db"
                }
              ]
            },
            "name": "config-volume"
          }
        ]
      }
    }
  }
}
EOF
  sleep 1
  $KUBECTL wait pod --for=condition=Ready -l '!job-name' -n kube-system --timeout=15s

  echo "${green}✅ Magic DNS${reset}"
}

if [ "$0" = "${BASH_SOURCE[0]}" ]; then
  set -o errexit
  set -o nounset
  set -o pipefail

  function main() {
    patch_hosts
  }
  main "$@"
fi
