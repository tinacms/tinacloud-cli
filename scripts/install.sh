#!/usr/bin/env bash

# Some helpful functions
yell() { echo -e "${RED}FAILED> $* ${NC}" >&2; }
die() { yell "$*"; exit 1; }
try() { "$@" || die "failed executing: $*"; }
log() { echo -e "$*"; }

# Colors for colorizing
RED='\033[0;31m'
GREEN='\033[0;32m'
PURPLE='\033[0;35m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

INSTALL_PATH=${INSTALL_PATH:-"/usr/local/bin"}
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

DEFAULT_VERSION="v0.1.1"
version=${VERSION:-$DEFAULT_VERSION}

if [ ! $version ]; then
    log "${YELLOW}"
    log "Failed while attempting to install tinacloud cli. Please manually install:"
    log ""
    log "1. Open your web browser and go to https://github.com/tinacms/tinacloud-cli/releases"
    log "2. Download the cli from the latest release for your platform. Make sure rename it tinacloud"
    log "3. chmod +x ./tinacloud"
    log "4. mv ./tinacloud /usr/local/bin"
    log "${NC}"
    die "exiting..."
fi

log "Selected version: $version"

log "${YELLOW}"
log NOTE: Install a specific version of the CLI by using VERSION variable
log 'curl -L https://github.com/tinacms/tinacloud-cli/raw/scripts/install.sh | bash'
log "${NC}"

# check for existing tinacloud cli installation
hasCli=$(which tinacloud)
if [ "$?" = "0" ]; then
    log ""
    log "${GREEN}You already have the TinaCloud cli at '${hasCli}'${NC}"
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
    die "prebuilt binaries for ${arch} architecture not available, please try building from source https://github.com/tinacms/tinacloud-cli"
fi

# some variables
suffix="-${platform}-${arch}.tar.gz"
targetFile="/tmp/tinacloud$suffix"
targetExe="/tmp/tinacloud"

if [ -e "$targetFile" ]; then
    rm "$targetFile"
fi

if [ -e "$targetExe" ]; then
  rm "$targetExe"
fi

log "${PURPLE}Downloading TinaCloud cli for $platform-$arch to ${targetFile}${NC}"
url=https://github.com/tinacms/tinacloud-cli/releases/download/$version/tinacloud$suffix

try curl -L# -f -o "$targetFile" "$url"
try tar -xzf "$targetFile" -C "/tmp"
try chmod +x "$targetExe"

log "${GREEN}Download complete!${NC}"

# check for sudo
needSudo=$(mkdir -p "${INSTALL_PATH}" && touch "${INSTALL_PATH}/.tinacloudinstall" &> /dev/null)
if [[ "$?" != "0" ]]; then
    NEED_SUDO=1
fi
rm "${INSTALL_PATH}/.tinacloudinstall" &> /dev/null

if [[ "$NEED_SUDO" == '1' ]]; then
    log
    log "${YELLOW}Path '$INSTALL_PATH' requires root access to write."
    log "${YELLOW}This script will attempt to execute the move command with sudo.${NC}"
    log "${YELLOW}Are you ok with that? (y/N)${NC}"
    read a
    if [[ $a == "Y" || $a == "y" || $a = "" ]]; then
        log
    else
        log
        log "  ${BLUE}sudo mv $targetFile ${INSTALL_PATH}/tinacloud${NC}"
        log
        die "Please move the binary manually using the command above."
    fi
fi

log "Moving TinaCloud cli from $targetExe to ${INSTALL_PATH}"

try maybe_sudo mv "$targetExe" "${INSTALL_PATH}/tinacloud"

log
log "${GREEN}TinaCloud cli installed to ${INSTALL_PATH}${NC}"
log

if [ -e "$targetFile" ]; then
    rm "$targetFile"
fi

if [ -e "$targetExe" ]; then
  rm "$targetExe"
fi

log "Running ${BLUE}tinacloud -v${NC}"
tinacloud -v

if ! echo "$PATH" | grep -q "$INSTALL_PATH"; then
    log
    log "${YELLOW}$INSTALL_PATH not found in \$PATH, you might need to add it${NC}"
    log
fi
