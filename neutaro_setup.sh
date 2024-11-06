#!/bin/bash 

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Neutaro Validator Auto-Setup Script${NC}"

# Prompt for Moniker Name
read -p "Enter your moniker name: " MONIKER

# Set default seeds and pruning options
DEFAULT_SEEDS="0e24a596dc34e7063ec2938baf05d09b374709e6@109.199.106.233:26656,84ae242b0c4c14af59a61438ba2eca4573b91c95@seed0.neutaro.tech:36656"
read -p "Enter seeds (default: $DEFAULT_SEEDS): " SEEDS
SEEDS=${SEEDS:-$DEFAULT_SEEDS} # Use default if input is empty

# Set pruning options to custom
PRUNING="custom"
PRUNING_KEEP_RECENT="100"
PRUNING_INTERVAL="19"

echo -e "${GREEN}Pruning is set to custom with the following settings:${NC}"
echo -e "${GREEN}Keep recent: $PRUNING_KEEP_RECENT, Interval: $PRUNING_INTERVAL${NC}"

# Update and install dependencies
echo -e "${GREEN}Updating and installing dependencies...${NC}"
sudo apt update && sudo apt upgrade -y
sudo apt install curl tar wget clang pkg-config libssl-dev jq build-essential bsdmainutils git make ncdu gcc chrony liblz4-tool pv -y
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to install dependencies. Exiting.${NC}"
  exit 1
fi

# Install Go with fallback download links
GO_VERSION="1.22.2"
echo -e "${GREEN}Installing Go $GO_VERSION...${NC}"
cd $HOME
if ! curl -LO "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz"; then
  echo -e "${YELLOW}Primary download link failed. Trying alternative link...${NC}"
  if ! curl -LO "https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz"; then
    echo -e "${RED}Failed to download Go. Exiting.${NC}"
    exit 1
  fi
fi

# Continue with Go installation
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go$GO_VERSION.linux-amd64.tar.gz"
rm "go$GO_VERSION.linux-amd64.tar.gz"
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> $HOME/.bash_profile
source $HOME/.bash_profile
go version >/dev/null 2>&1

if ! command -v go &> /dev/null || [[ "$(go version)" != *"$GO_VERSION"* ]]; then
  echo -e "${RED}Go installation failed or incorrect version installed. Exiting.${NC}"
  exit 1
fi

# Clone and build Neutaro
echo -e "${GREEN}Cloning and building Neutaro...${NC}"
cd $HOME
git clone https://github.com/Neutaro/Neutaro || { echo -e "${RED}Failed to clone Neutaro repository. Exiting.${NC}"; exit 1; }
cd Neutaro
make build || { echo -e "${RED}Failed to build Neutaro. Exiting.${NC}"; exit 1; }

# Initialize and configure the node
echo -e "${GREEN}Configuring your Neutaro node...${NC}"
Neutaro init $MONIKER --chain-id Neutaro-1 || { echo -e "${RED}Failed to initialize the node. Exiting.${NC}"; exit 1; }
Neutaro config chain-id Neutaro-1
Neutaro config keyring-backend os
curl http://154.26.153.186/genesis.json > $HOME/.Neutaro/config/genesis.json

# Configure app.toml
echo -e "${GREEN}Setting gas prices and pruning options...${NC}"
sed -i "s/^minimum-gas-prices *=.*/minimum-gas-prices = \"0uneutaro\"/" $HOME/.Neutaro/config/app.toml
sed -i "s/^pruning *=.*/pruning = \"$PRUNING\"/" $HOME/.Neutaro/config/app.toml
sed -i "s/^pruning-keep-recent *=.*/pruning-keep-recent = \"$PRUNING_KEEP_RECENT\"/" $HOME/.Neutaro/config/app.toml
sed -i "s/^pruning-interval *=.*/pruning-interval = \"$PRUNING_INTERVAL\"/" $HOME/.Neutaro/config/app.toml

# Configure config.toml
sed -i "s/^seeds *=.*/seeds = \"$SEEDS\"/" $HOME/.Neutaro/config/config.toml
if [ -n "$PERSISTENT_PEERS" ]; then
  sed -i "s/^persistent_peers *=.*/persistent_peers = \"$PERSISTENT_PEERS\"/" $HOME/.Neutaro/config/config.toml
fi

# Open port 26656 (recommended)
sudo ufw allow 26656/tcp

# Download and apply the new snapshot with progress
echo -e "${GREEN}Downloading and applying the blockchain snapshot...${NC}"
cd $HOME
mv $HOME/.Neutaro/data $HOME/.Neutaro/data-old || echo "data-old directory does not exist"
mv $HOME/.Neutaro/wasm $HOME/.Neutaro/wasm-old || echo "wasm-old directory does not exist"
wget https://snapshot.neutaro.tech/latest.tar.lz4

# Unpacking the snapshot with progress
echo -e "${GREEN}Unpacking the snapshot...${NC}"
pv latest.tar.lz4 | lz4 -d - | tar xvf - -C $HOME/.Neutaro --strip-components=1 > /dev/null 2>&1

# Clean up the downloaded snapshot file
rm -r latest.tar.lz4

# Function to fetch the latest block height from Neutaro API endpoint
get_latest_block_height() {
  curl -s https://api1.neutaro.tech/cosmos/base/tendermint/v1beta1/blocks/latest | jq -r '.block.header.height'
}

# Function to fetch the current block height of the local node
get_current_block_height() {
  Neutaro status 2>&1 | jq -r '.SyncInfo.latest_block_height'
}

# Function to check if the node is catching up
is_node_catching_up() {
  Neutaro status 2>&1 | jq -r '.SyncInfo.catching_up'
}

# Wait for a few seconds before starting the sync check
echo -e "${YELLOW}Waiting for the node to initialize...${NC}"
sleep 60  # Adjust the sleep duration as needed

# Main script to check sync status
echo -e "${GREEN}Checking Neutaro Node Sync Status...${NC}"

while true; do
  # Get the latest block height from the network
  latest_block_height=$(get_latest_block_height)
  if [ -z "$latest_block_height" ]; then
    echo -ne "Failed to fetch the latest block height. Retrying...\r"
    sleep 10
    continue
  fi

  # Get the current block height of the node
  current_block_height=$(get_current_block_height)
  if [ -z "$current_block_height" ]; then
    echo -ne "Failed to fetch the current block height of the node. Retrying...\r"
    sleep 10
    continue
  fi

  # Check if the node is catching up
  if [ "$(is_node_catching_up)" == "false" ]; then
    echo -ne "Node is fully synced! Current Block Height: $current_block_height\n"
    break
  fi

  # Calculate progress percentage
  progress=$(( (current_block_height * 100) / latest_block_height ))

  # Display sync progress on the same line
  echo -ne "Syncing: $progress% - Current Height: $current_block_height / Latest Height: $latest_block_height\r"

  # Wait for some time before the next check
  sleep 10
done

echo -e "${GREEN}Setup complete! Check sync status or monitor logs if needed.${NC}"
