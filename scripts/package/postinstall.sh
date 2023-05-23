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

# Install handles systemd service management for debian and rhel based
# platforms. This function can be called more than once as it is idempotent.
install() {
    systemctl daemon-reload

    # Debian based platforms should enable and start the bindplane service.
    if command -v dpkg >/dev/null; then
        systemctl enable --now bindplane
    fi
}

# Upgrade performs the same steps as install.
upgrade() {
    install
}

action="$1"

case "$action" in
  "0" | "install")
    install
    ;;
  "1" | "upgrade")
    upgrade
    ;;
  *)
    install
    ;;
esac
