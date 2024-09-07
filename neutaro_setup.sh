#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Neutaro Validator Auto-Setup Script${NC}"

# Prompt for Moniker Name
read -p "Enter your moniker name: " MONIKER

# Set default seeds and pruning options
DEFAULT_SEEDS="84ae242b0c4c14af59a61438ba2eca4573b91c95@seed0.neutaro.tech:36656"
read -p "Enter seeds (default: $DEFAULT_SEEDS): " SEEDS
SEEDS=${SEEDS:-$DEFAULT_SEEDS} # Use default if input is empty

echo "Pruning options (choose one):"
echo "1) custom (recommended)"
echo "2) default"
echo "3) nothing"
read -p "Enter pruning option (1-3): " PRUNING_CHOICE

case $PRUNING_CHOICE in
  1)
    PRUNING="custom"
    PRUNING_KEEP_RECENT="100"
    PRUNING_INTERVAL="19"
    ;;
  2)
    PRUNING="default"
    ;;
  3)
    PRUNING="nothing"
    ;;
  *)
    echo -e "${RED}Invalid choice, using default pruning settings.${NC}"
    PRUNING="default"
    ;;
esac

# Inform about persistent peers
echo -e "${GREEN}Important: If you face sync issues due to limited seeds, you may need to add more persistent peers.${NC}"
echo -e "${GREEN}Also, opening port 26656 for TCP connections is recommended to share your node with others and improve the network.${NC}"

# Prompt for persistent peers (optional)
read -p "Enter persistent peers (optional, leave blank to skip): " PERSISTENT_PEERS

# Update and install dependencies
echo -e "${GREEN}Updating and installing dependencies...${NC}"
sudo apt update && sudo apt upgrade -y
sudo apt install curl tar wget clang pkg-config libssl-dev jq build-essential bsdmainutils git make ncdu gcc chrony liblz4-tool -y

# Install Go
GO_VERSION="1.22.2"
echo -e "${GREEN}Installing Go $GO_VERSION...${NC}"
cd $HOME
wget "https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go$GO_VERSION.linux-amd64.tar.gz"
rm "go$GO_VERSION.linux-amd64.tar.gz"
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> $HOME/.bash_profile
source $HOME/.bash_profile
go version

# Clone and build Neutaro
echo -e "${GREEN}Cloning and building Neutaro...${NC}"
cd $HOME
git clone https://github.com/Neutaro/Neutaro
cd Neutaro
make build

# Install Cosmovisor
echo -e "${GREEN}Setting up Cosmovisor...${NC}"
mkdir -p $HOME/.Neutaro/cosmovisor/genesis/bin
mv $HOME/Neutaro/build/Neutaro $HOME/.Neutaro/cosmovisor/genesis/bin/
sudo ln -s $HOME/.Neutaro/cosmovisor/genesis $HOME/.Neutaro/cosmovisor/current
sudo ln -s $HOME/.Neutaro/cosmovisor/current/bin/Neutaro /usr/local/bin/Neutaro
cd $HOME/Neutaro/
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.4.0

# Initialize and configure the node
echo -e "${GREEN}Configuring your Neutaro node...${NC}"
Neutaro init $MONIKER --chain-id Neutaro-1
Neutaro config chain-id Neutaro-1
Neutaro config keyring-backend os
curl http://154.26.153.186/genesis.json > $HOME/.Neutaro/config/genesis.json

# Configure app.toml
echo -e "${GREEN}Setting gas prices and pruning options...${NC}"
sed -i "s/^minimum-gas-prices *=.*/minimum-gas-prices = \"0uneutaro\"/" $HOME/.Neutaro/config/app.toml
sed -i "s/^pruning *=.*/pruning = \"$PRUNING\"/" $HOME/.Neutaro/config/app.toml
if [ "$PRUNING" == "custom" ]; then
  sed -i "s/^pruning-keep-recent *=.*/pruning-keep-recent = \"$PRUNING_KEEP_RECENT\"/" $HOME/.Neutaro/config/app.toml
  sed -i "s/^pruning-interval *=.*/pruning-interval = \"$PRUNING_INTERVAL\"/" $HOME/.Neutaro/config/app.toml
fi

# Configure config.toml
sed -i "s/^seeds *=.*/seeds = \"$SEEDS\"/" $HOME/.Neutaro/config/config.toml
if [ -n "$PERSISTENT_PEERS" ]; then
  sed -i "s/^persistent_peers *=.*/persistent_peers = \"$PERSISTENT_PEERS\"/" $HOME/.Neutaro/config/config.toml
fi

# Open port 26656 (recommended)
sudo ufw allow 26656/tcp

# Download and apply the new snapshot
echo -e "${GREEN}Downloading and applying the blockchain snapshot...${NC}"
cd $HOME
mv $HOME/.Neutaro/data $HOME/.Neutaro/data-old || echo "data-old directory does not exist"
mv $HOME/.Neutaro/wasm $HOME/.Neutaro/wasm-old || echo "wasm-old directory does not exist"
wget https://poker.neutaro.tech/snapshot010924.tar.lz4
lz4 -d snapshot010924.tar.lz4 -c | tar xvf -
rm -r snapshot010924.tar.lz4

# Create systemd service for Neutaro
echo -e "${GREEN}Creating systemd service for Neutaro...${NC}"
sudo tee /etc/systemd/system/Neutaro.service > /dev/null << EOF
[Unit]
Description=Neutaro Node Service
After=network-online.target

[Service]
User=$USER
ExecStart=$(which cosmovisor) run start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment="DAEMON_HOME=$HOME/.Neutaro"
Environment="DAEMON_NAME=Neutaro"
Environment="UNSAFE_SKIP_BACKUP=true"

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
echo -e "${GREEN}Starting Neutaro service...${NC}"
sudo systemctl daemon-reload
sudo systemctl enable Neutaro
sudo systemctl restart Neutaro

# Fix potential issues with missing binaries
mkdir -p $HOME/.Neutaro/cosmovisor/upgrades/v2/bin
cp $HOME/.Neutaro/cosmovisor/genesis/bin/Neutaro $HOME/.Neutaro/cosmovisor/upgrades/v2/bin/

# Display sync status
echo -e "${GREEN}Setup complete! Check the sync status with:${NC}"
echo "Neutaro status 2>&1 | jq .SyncInfo"

# Monitor the logs
echo -e "${GREEN}To monitor the logs, use:${NC}"
echo "sudo journalctl -fu Neutaro -o cat"

# Additional Instructions for After the Node is Synced
echo -e "${GREEN}Once the node is synced, follow the steps below to configure your validator and begin participating in the network:${NC}"
echo "1. If needed, recover your wallet using:"
echo "   Neutaro keys add WALLET --keyring-backend os --recover"
echo "2. Once your wallet is funded, create your validator with:"
echo "   Neutaro tx staking create-validator --amount=1000000uneutaro --pubkey=\$(Neutaro tendermint show-validator) --moniker=$MONIKER --chain-id=Neutaro-1 --from WALLET --keyring-backend os --commission-rate=\"0.10\" --details=\"About_Your_Validator\" --commission-max-rate=\"0.20\" --commission-max-change-rate=\"0.01\" --min-self-delegation=\"1000000\" --gas=\"auto\" --gas-prices=\"0.0025uneutaro\" --gas-adjustment=\"1.5\""

# End of script
