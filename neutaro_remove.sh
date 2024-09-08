#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Function to show progress with success or failure message
show_progress() {
    local -r msg=$1
    local -r cmd=$2

    echo -ne "${GREEN}${msg}...${NC}\n"
    sleep 0.5 # Simulate progress

    if eval "$cmd"; then
        echo -e "${GREEN}✔ ${msg} completed successfully.${NC}"
    else
        echo -e "${RED}✖ ${msg} failed to complete.${NC}"
    fi
}

# Stopping and disabling the Neutaro service
show_progress "Stopping and disabling the Neutaro service" "sudo systemctl stop Neutaro > /dev/null 2>&1 && sudo systemctl disable Neutaro > /dev/null 2>&1"

# Removing Neutaro service file
show_progress "Removing Neutaro service file" "sudo rm -f /etc/systemd/system/Neutaro.service"

# Reloading systemd daemon
show_progress "Reloading systemd daemon" "sudo systemctl daemon-reload > /dev/null 2>&1"

# Removing Neutaro and Cosmovisor binaries and configurations
show_progress "Removing Neutaro and Cosmovisor binaries and configurations" "sudo rm -rf $HOME/.Neutaro > /dev/null 2>&1 && sudo rm -rf $HOME/Neutaro > /dev/null 2>&1 && sudo rm -rf /usr/local/bin/Neutaro > /dev/null 2>&1 && sudo rm -rf $HOME/go/bin/cosmovisor > /dev/null 2>&1 && sudo rm -rf $HOME/.bash_profile > /dev/null 2>&1"

# Cleaning up Go installation
show_progress "Cleaning up Go installation" "sudo rm -rf /usr/local/go > /dev/null 2>&1"

# Cleaning up any residual files
show_progress "Cleaning up any residual files" "sudo rm -rf /root/.Neutaro > /dev/null 2>&1 && sudo rm -rf $HOME/.cache/go-build > /dev/null 2>&1 && sudo rm -rf /var/log/Neutaro* > /dev/null 2>&1"

echo -e "${GREEN}Neutaro setup has been completely removed from your system.${NC}"
