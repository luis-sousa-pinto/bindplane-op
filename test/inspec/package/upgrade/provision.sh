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

# install installs the 1.12.2 official release of BindPlane OP.
install() {
    curl -fsSlL https://github.com/observiq/bindplane-op/releases/latest/download/install-linux.sh | bash -s -- --version 1.12.2
    systemctl enable --now bindplane
}

# upgrade uses the system's package manager and upgrade the package directly.
# We expect the service to continue running after the upgrade. We do not want
# to use the install script to perform the upgrade, because that would mask any
# issues related to service management.
upgrade() {
    if command -v "dpkg" > /dev/null ; then
        # Run with '--allow-downgrades' because apt will consider 
        # snapshot builds a "downgrade" sometimes but the behavior
        # is the same as an actual upgrade.
        sudo apt-get install -y --allow-downgrades /tmp/data/bindplane_*_linux_amd64.deb
    elif command -v "zypper" > /dev/null ; then
        sudo rpm -i /tmp/data/bindplane_*_linux_amd64.rpm
    elif command -v "rpm" > /dev/null ; then
        sudo yum install -y /tmp/data/bindplane_*_linux_amd64.rpm
    else
        echo "failed to detect platform type"
        exit 1
    fi
}

install
upgrade
