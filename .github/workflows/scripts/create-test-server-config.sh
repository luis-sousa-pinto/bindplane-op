#!/usr/bin/env bash
# Copyright  observIQ, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

cat <<EOF | sudo tee /etc/bindplane/config.yaml
apiVersion: bindplane.observiq.com/v1
network:
    host: 127.0.0.1
    port: "3001"
auth:
    username: admin
    password: admin
    secretKey: $(uuidgen)
    sessionSecret: $(uuidgen)
store:
    type: bbolt
    bbolt:
        path: /var/lib/bindplane/storage/bindplane.db
logging:
    output: file
    filePath: /var/log/bindplane/bindplane.log
EOF
