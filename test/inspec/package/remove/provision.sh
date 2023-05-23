#!/bin/bash
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

install() {
    if command -v "dpkg" > /dev/null ; then
        sudo bash /tmp/data/install-linux.sh --file /tmp/data/bindplane_*_linux_amd64.deb
    elif command -v "rpm" > /dev/null ; then
        sudo bash /tmp/data/install-linux.sh --file /tmp/data/bindplane_*_linux_amd64.rpm
    else
        echo "failed to detect platform type"
        exit 1
    fi
}

# Modify the config to prevent package manager from revmoving it.
# Inspec will test to ensure the file still exists.
config() {
    cat <<CONFIG | sudo tee /etc/bindplane/config.yaml
host: 127.0.0.1
port: "3001"
serverURL: http://127.0.0.1:3001
username: admin
password: admin
logFilePath: /var/log/bindplane/bindplane.log
server:
    storageFilePath: /var/lib/bindplane/storage/bindplane.db
    secretKey: 3cbca8e5-2ca2-4cd3-8e58-a4ca248470b0
    remoteURL: ws://127.0.0.1:3001
    sessionsSecret: 33914edc-a3f8-41cf-949f-4a16a40bcd03
CONFIG
}

uninstall() {
    if command -v "dpkg" > /dev/null ; then
        sudo apt-get remove -y bindplane
    elif command -v "zypper" > /dev/null ; then
        sudo zypper remove -y bindplane
    elif command -v "rpm" > /dev/null ; then
        sudo yum remove -y bindplane
    else
        echo "failed to detect platform type"
        exit 1
    fi
}

install
config
uninstall
