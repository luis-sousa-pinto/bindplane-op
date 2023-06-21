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

# TODO don't strip password

set -e

PREREQS="printf sed tar gzip bindplane tr uname xargs basename"
INDENT_WIDTH='  '
indent=""

bindplane_log_dir=/var/log/bindplane/

# Colors
num_colors=$(tput colors 2>/dev/null)
if test -n "$num_colors" && test "$num_colors" -ge 8; then
  reset="$(tput sgr0)"
  fg_cyan="$(tput setaf 6)"
  fg_green="$(tput setaf 2)"
  fg_red="$(tput setaf 1)"
  fg_yellow="$(tput setaf 3)"
fi

if [ -z "$reset" ]; then
  sed_ignore=''
else
  sed_ignore="/^[$reset]+$/!"
fi

printf() {
  if command -v sed >/dev/null; then
    command printf -- "$@" | sed -E "$sed_ignore s/^/$indent/g"  # Ignore sole reset characters if defined
  else
    # Ignore $* suggestion as this breaks the output
    # shellcheck disable=SC2145
    command printf -- "$indent$@"
  fi
}

increase_indent() { indent="$INDENT_WIDTH$indent" ; }
decrease_indent() { indent="${indent#*"$INDENT_WIDTH"}" ; }

# Color functions reset only when given an argument
# Ignore "parameters are never passed"
# shellcheck disable=SC2120
reset() { command printf "$reset$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_cyan() { command printf "$fg_cyan$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_green() { command printf "$fg_green$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_red() { command printf "$fg_red$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_yellow() { command printf "$fg_yellow$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }

# Intentionally using variables in format string
# shellcheck disable=SC2059
info() { printf "$*\\n" ; }

# Intentionally using variables in format string
# shellcheck disable=SC2059
error() {
  increase_indent
  printf "$fg_red$*$reset\\n"
  decrease_indent
}

# Intentionally using variables in format string
# shellcheck disable=SC2059
success() { printf "$fg_green$*$reset\\n" ; }

observiq_banner()
{
  fg_cyan "           888                                        8888888 .d88888b.\\n"
  fg_cyan "           888                                          888  d88P\" \"Y88b\\n"
  fg_cyan "           888                                          888  888     888\\n"
  fg_cyan "   .d88b.  88888b.  .d8888b   .d88b.  888d888 888  888  888  888     888\\n"
  fg_cyan "  d88\"\"88b 888 \"88b 88K      d8P  Y8b 888P\"   888  888  888  888     888\\n"
  fg_cyan "  888  888 888  888 \"Y8888b. 88888888 888     Y88  88P  888  888 Y8b 888\\n"
  fg_cyan "  Y88..88P 888 d88P      X88 Y8b.     888      Y8bd8P   888  Y88b.Y8b88P\\n"
  fg_cyan "   \"Y88P\"  88888P\"   88888P'  \"Y8888  888       Y88P  8888888 \"Y888888\"\\n"
  fg_cyan "                                                                   Y8b  \\n"

  reset
}

separator() { printf "===================================================\\n" ; }

banner() {
  printf "\\n"
  separator
  printf "| %s\\n" "$*" ;
  separator
}

usage() {
  increase_indent
  USAGE=$(cat <<EOF
Usage:
  Collects support bundle for BindPlane OP
  $(fg_yellow '-h, --help')
      Prints this help message

  $(fg_yellow '-a, --agent')
      An optional agent ID to collect info about.
EOF
  )
  info "$USAGE"
  decrease_indent
  return 0
}

force_exit() {
  # Exit regardless of subshell level with no "Terminated" message
  kill -PIPE $$
  # Call exit to handle special circumstances (like running script during docker container build)
  exit 1
}

error_exit() {
  line_num=$(if [ -n "$1" ]; then command printf ":$1"; fi)
  error "ERROR ($SCRIPT_NAME$line_num): ${2:-Unknown Error}" >&2
  if [ -n "$0" ]; then
    increase_indent
    error "$*"
    decrease_indent
  fi
  force_exit
}

succeeded() {
  increase_indent
  success "Succeeded!"
  decrease_indent
}

failed() {
  error "Failed!"
}

root_check() {
  system_user_name=$(id -un)
  if [[ "${system_user_name}" != 'root' || $EUID -ne 0 ]]; then
    failed
    error_exit "$LINENO" "Script needs to be run as root or with sudo"
  fi
}

# This will check if the current environment has
# all required shell dependencies to run the installation.
dependencies_check() {
  info "Checking for script dependencies..."
  FAILED_PREREQS=''
  for prerequisite in $PREREQS; do
    if command -v "$prerequisite" >/dev/null; then
      continue
    else
      if [ -z "$FAILED_PREREQS" ]; then
        FAILED_PREREQS="${fg_red}$prerequisite${reset}"
      else
        FAILED_PREREQS="$FAILED_PREREQS, ${fg_red}$prerequisite${reset}"
      fi
    fi
  done

  if [ -n "$FAILED_PREREQS" ]; then
    failed
    error_exit "$LINENO" "The following dependencies are required by this script: [$FAILED_PREREQS]"
  fi
  succeeded
}

check_prereqs() {
  banner "Checking Prerequisites"
  increase_indent
  root_check
  dependencies_check
  success "Prerequisite check complete!"
  decrease_indent
}

function bundle_files() {
    banner "Collecting files for support bundle"
    increase_indent

    tar_filename="bindplane_support_bundle_$(date +%Y%m%d_%H%M%S).tar"
    
    read -r -p "Would you like to include the BindPlane config file? (y or n) " include_bindplane_config

    bindplane_cmd="bindplane"    
    bindplane_config_file="/etc/bindplane/config.yaml"
    set +e
    ($bindplane_cmd profile get --current > bindplane_config.yaml 2>&1)
    exit_status=$?    
    # Collect the BindPlane config
    if [ $exit_status -ne 0 ]; then
        info "Failed to to get bindplane profile"
        info "Trying /etc/bindplane/config.yaml"
        if [ -f "$bindplane_config_file" ]; then
            bindplane_cmd="bindplane --config $bindplane_config_file"
        else
            error "Failed to find BindPlane config file"
            # prompt for config location
            read -r -p "Please enter the full path to the BindPlane config file. This is required for collecting information even if you're not including it in the support bundle: " bindplane_config_file
            if [ -f "$bindplane_config_file" ]; then
                bindplane_cmd="bindplane --config $bindplane_config_file"
            else
                info "Failed to find $fg_red $bindplane_config_file($reset) BindPlane config file"
                failed
                error_exit "$LINENO" "Unable to find BindPlane config file"
            fi
        fi
        if [ "$include_bindplane_config" != "n" ]; then
          cp "$bindplane_config_file" bindplane_config.yaml
        fi
    fi
    if [ "$include_bindplane_config" != "n" ]; then
      info "Collecting BindPlane config"
      # remove password field
      if [[ "$OSTYPE" == "darwin"* ]]; then
          sed -i "" '/\s*password:.*/d' bindplane_config.yaml  # macOS
      else
          sed -i '/\s*password:.*/d' bindplane_config.yaml      # Linux
      fi
      tar -cf "$tar_filename" bindplane_config.yaml
    fi
    rm -f bindplane_config.yaml
    
    # BindPlane version info
    info "Collecting version info"
    ($bindplane_cmd version > version_info.txt)
    tar --append --file="$tar_filename" version_info.txt
    rm version_info.txt

    # Check if BindPlane logs directory exists
    if [ ! -d "$bindplane_log_dir" ]; then
        info "Log file directory $fg_red $bindplane_log_dir$(reset) does not exist."
        read -r -p "Please enter the full path to the BindPlane log directory: " bindplane_log_dir
        if [ ! -d "$bindplane_log_dir" ]; then
            info "Directory $fg_red $bindplane_log_dir$(reset) does not exist."
            info "Skipping BindPlane log collection"
        fi
    fi
    
    # Ask user if they want just the most recent log or all logs
    if [ -d "$bindplane_log_dir" ]; then
        info "BindPlane log directory: $fg_cyan $bindplane_log_dir$(reset)"
        read -r -p "Would you like to collect all logs? (y or n): " collect_all_logs
        if [ "$collect_all_logs" != "n" ]; then
            info "Collecting all BindPlane logs"
            # only grab logs
            while IFS= read -r -d '' file; do
              tar --append --file="$tar_filename" -C "$bindplane_log_dir" "$(basename "$file")"
            done < <(find "$bindplane_log_dir" -name "*log*" -type f -print0)
        else
            info "Collecting most recent BindPlane log"
            tar --append --file="$tar_filename" -C "$bindplane_log_dir" bindplane.log
        fi
    fi
    
    # If agent argument is set, collect agent info
    if [ ! -z "$agent_id" ] ; then
        info "Collecting agent info"
        ("$bindplane_cmd" get agent "$agent_id" -o yaml > agent_info.yaml)
        tar --append --file="$tar_filename" agent_info.yaml
        rm agent_info.yaml

        info "Collecting agent configuration"
        config_name=$("$bindplane_cmd" get agent "$agent_id" |  sed -n -e 's/^.*configuration=\([^ ]*\).*$/\1/p' | tr -d '[:space:]')

        info "Agent is labeled with configuration: $fg_cyan $config_name$(reset)"
        # Get the current, pending, and latest versions of the agent config

        ("$bindplane_cmd" get configuration "$config_name" -o yaml > agent_config_latest.yaml)
        ("$bindplane_cmd" get configuration "$config_name":current -o yaml > agent_config_current.yaml)
        ("$bindplane_cmd" get configuration "$config_name":pending -o yaml > agent_config_pending.yaml)

        tar --append --file="$tar_filename" agent_config_latest.yaml agent_config_current.yaml agent_config_pending.yaml
        rm agent_config_latest.yaml agent_config_current.yaml agent_config_pending.yaml
    fi

    # Check if the files exist, if yes append them to the tar file
    for file in issue os-release
    do
        if [ -f "/etc/$file" ]; then
            # These might be symlinks, so cat them to real files
            cat "/etc/$file" > "$file"
            tar --append --file="$tar_filename" $file
            rm $file
            info "Added file $(fg_cyan "/etc/$file")$(reset) to the tar file."
        else
            info "File $(fg_red "/etc/$file")$(reset) does not exist."
        fi
    done

    # Compress the tar file
    info "Compressing the tar file..."
    gzip "$tar_filename"

    info "Files have been added to the file $(realpath "$tar_filename.gz") successfully."
    decrease_indent
}


main() {
  if [ $# -ge 1 ]; then
    while [ -n "$1" ]; do
      case "$1" in  
        -a|--agent)
          agent_id=$2 ; shift 2
          ;;              
        -h|--help)
          usage
          force_exit
          ;;
      --)
        shift; break ;;
      *)
        error "Invalid argument: $1"
        usage
        force_exit
        ;;
      esac
    done
  fi

  observiq_banner
  check_prereqs
  bundle_files
}

main "$@"

