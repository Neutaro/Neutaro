#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# ------------------------------------------------------------------
# SAFETY GATE: this script deletes ~/.Neutaro INCLUDING YOUR VALIDATOR
# KEYS and wallet keyring. That identity can never be recreated.
# ------------------------------------------------------------------
echo -e "${RED}WARNING:${NC} this will permanently delete:"
echo "  - \$HOME/.Neutaro   (INCLUDING priv_validator_key.json and your keyring)"
echo "  - \$HOME/Neutaro    (the source checkout)"
echo "  - /usr/local/go     (the system-wide Go install, also used by other software)"
echo
echo "If there is ANY chance you want this validator identity again, back up"
echo "\$HOME/.Neutaro/config/priv_validator_key.json and your keyring first."
echo
read -r -p "Type 'delete' to proceed: " CONFIRM
if [ "$CONFIRM" != "delete" ]; then
    echo "Aborted. Nothing was removed."
    exit 1
fi

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
show_progress "Removing Neutaro and Cosmovisor binaries and configurations" "sudo rm -rf $HOME/.Neutaro > /dev/null 2>&1 && sudo rm -rf $HOME/Neutaro > /dev/null 2>&1 && sudo rm -rf /usr/local/bin/Neutaro > /dev/null 2>&1 && sudo rm -rf $HOME/go/bin/cosmovisor > /dev/null 2>&1"

# Cleaning up Go installation
show_progress "Cleaning up Go installation" "sudo rm -rf /usr/local/go > /dev/null 2>&1"

# Cleaning up any residual files
show_progress "Cleaning up any residual files" "sudo rm -rf /root/.Neutaro > /dev/null 2>&1 && sudo rm -rf $HOME/.cache/go-build > /dev/null 2>&1 && sudo rm -rf /var/log/Neutaro* > /dev/null 2>&1"

echo -e "${GREEN}Neutaro setup has been completely removed from your system.${NC}"
