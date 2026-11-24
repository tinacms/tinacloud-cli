#!/usr/bin/env bash

# Some helpful functions
yell() { echo -e "${RED}FAILED> $* ${NC}" >&2; }
die() { yell "$*"; exit 1; }
try() { "$@" || die "failed executing: $*"; }
log() { echo -e "$*"; }

# Use colour only when it will be displayed in a terminal.
if [[ -t 1 && -z "${NO_COLOR:-}" && "${TERM:-}" != "dumb" ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    PURPLE='\033[0;35m'
    BLUE='\033[0;34m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    PURPLE=''
    BLUE=''
    YELLOW=''
    NC=''
fi

INSTALL_PATH=${INSTALL_PATH:-"/usr/local/bin"}
MAN_PATH=${MAN_PATH:-"/usr/local/share/man/man1"}
NEED_SUDO=0

function maybe_sudo() {
    if [[ "$NEED_SUDO" == '1' ]]; then
        sudo "$@"
    else
        "$@"
    fi
}

# check for curl
hasCurl=$(which curl)
if [ "$?" = "1" ]; then
    die "You need to install curl to use this script."
fi

log "Selecting version..."

DEFAULT_VERSION="v0.2.0"
version=${VERSION:-$DEFAULT_VERSION}

log "Selected version: $version"

log "${YELLOW}"
log NOTE: Install a specific version of the CLI by using VERSION variable
log 'curl -L https://github.com/tinacms/tinacloud-cli/raw/main/scripts/install.sh | bash'
log "${NC}"

# check for existing tinacloud cli installation
hasCli=$(which tinacloud)
if [ "$?" = "0" ]; then
    log ""
    log "${GREEN}You already have the TinaCloud CLI at '${hasCli}'${NC}"
    export n=3
    log "${YELLOW}Downloading again in $n seconds... Press Ctrl+C to cancel.${NC}"
    log ""
    sleep $n
fi

# get platform and arch
platform='unknown'
unamestr=`uname`
if [[ "$unamestr" == 'Linux' ]]; then
    platform='linux'
elif [[ "$unamestr" == 'Darwin' ]]; then
    platform='darwin'
fi

if [[ "$platform" == 'unknown' ]]; then
    die "Unknown OS platform"
fi

arch='unknown'
archstr=`uname -m`
if [[ "$archstr" == 'x86_64' ]]; then
    arch='x86_64'
elif [[ "$archstr" == 'arm64' ]] || [[ "$archstr" == 'aarch64' ]]; then
    arch='aarch64'
else
    die "prebuilt binaries for ${archstr} architecture not available, please try building from source https://github.com/tinacms/tinacloud-cli"
fi

# some variables
suffix="-${platform}-${arch}.tar.gz"
targetFile="/tmp/tinacloud$suffix"
targetExe="/tmp/tinacloud"
targetMan="/tmp/man"

if [ -e "$targetFile" ]; then
    rm "$targetFile"
fi

if [ -e "$targetExe" ]; then
  rm "$targetExe"
fi

if [ -e "$targetMan" ]; then
  rm -rf "$targetMan"
fi

log "${PURPLE}Downloading TinaCloud CLI for $platform-$arch to ${targetFile}${NC}"
url=https://github.com/tinacms/tinacloud-cli/releases/download/$version/tinacloud$suffix

try curl -L# -f -o "$targetFile" "$url"
try tar -xzf "$targetFile" -C "/tmp"
try chmod +x "$targetExe"

log "${GREEN}Download complete!${NC}"

# Check whether either install destination needs sudo.
can_write() {
    local install_dir="$1"
    local marker="${install_dir}/.tinacloudinstall"
    mkdir -p "${install_dir}" &> /dev/null &&
        touch "${marker}" &> /dev/null &&
        rm -f "${marker}"
}

if ! can_write "${INSTALL_PATH}" || ! can_write "${MAN_PATH}"; then
    NEED_SUDO=1
fi

if [[ "$NEED_SUDO" == '1' ]]; then
    log
    log "${YELLOW}An installation path requires root access to write."
    log "${YELLOW}This script will attempt to execute the move command with sudo.${NC}"
    if [ -t 0 ]; then
        log "${YELLOW}Are you ok with that? (y/N)${NC}"
        read a
    elif { exec 3< /dev/tty; } 2>/dev/null; then
        log "${YELLOW}Are you ok with that? (y/N)${NC}"
        read a <&3
        exec 3<&-
    fi
    if [[ $a == "Y" || $a == "y" ]]; then
        log
    else
        log
        log "  ${BLUE}sudo mv $targetExe ${INSTALL_PATH}/tinacloud${NC}"
        log "  ${BLUE}sudo mv $targetMan/* ${MAN_PATH}/"
        log
        die "Please move the binary manually using the command above."
    fi
fi

log "Moving TinaCloud CLI from $targetExe to ${INSTALL_PATH}"

try maybe_sudo mkdir -p "${INSTALL_PATH}"
try maybe_sudo mv "$targetExe" "${INSTALL_PATH}/tinacloud"

log "Moving TinaCloud CLI MAN pages from $targetMan/* to ${MAN_PATH}"

try maybe_sudo mkdir -p "${MAN_PATH}"
try maybe_sudo mv ${targetMan}/* "${MAN_PATH}/"

log
log "${GREEN}TinaCloud CLI installed to ${INSTALL_PATH}${NC}"
log

if [ -e "$targetFile" ]; then
    rm "$targetFile"
fi

if [ -e "$targetExe" ]; then
  rm "$targetExe"
fi

if [ -e "$targetMan" ]; then
  rm -rf "$targetMan"
fi

log "Running ${BLUE}tinacloud -v${NC}"
tinacloud -v

if ! echo "$PATH" | grep -q "$INSTALL_PATH"; then
    log
    log "${YELLOW}$INSTALL_PATH not found in \$PATH, you might need to add it${NC}"
    log
fi
