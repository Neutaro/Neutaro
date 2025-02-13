# Why Neutaro

Neutaro is closely working with Timpi to help create the first truly decentralized search engine! On Neutaro, you can claim rewards for contributing to Timpi and vote on different proposals affecting Timpi, such as those regarding the ethical standpoint of the Timpi search engine.<br>

## Neutaro Validator Security Guide
For more details on security, please check the [Security Guide](https://github.com/Neutaro/Neutaro/blob/main/Security%20Guide.md).


## Introduction
Running a node means that you run the chains binary. Follow these steps to create a Validator that runs as a service on linux. <br>
You can be a part of Neutaro by holding tokens and delegating them, running a node, or becoming a validator.

## Delegating Tokens

By delegating your tokens to a validator, you increase the amount of staked tokens they have, thereby increasing their voting power. In return, you receive a share of their rewards based on the amount you have delegated.

To delegate your tokens to a Validator, use the command below. **1,000,000 uneutaro** is equal to **1 NTMPI**. If you want to stake 100 NTMPI tokens, you input `100000000uneutaro` into the command.

**Example Command:**

```shell
Neutaro tx staking delegate ValidatorAddress 1000000uneutaro --from YOURWALLET --chain-id Neutaro-1
```

# Before Running a Node

### **Open Port 26656 for Better Connectivity**

To improve network connectivity and allow your node to share seeds with others, you need to open port `26656` on both your Linux firewall and router.

#### **1. Open Port 26656 on Linux Firewall (UFW):**

Run the following commands to allow traffic on port `26656`:

```shell
sudo ufw allow 26656/tcp
sudo ufw reload
```

#### **2. Open Port 26656 on Your Router:**

Ensure port `26656` is open for **TCP** traffic in your router's Port Forwarding settings.

Opening port `26656` helps with better network performance and faster synchronization.


# Running a node
We suggest using **Ubuntu 22.04.03**, **4 cores, 8gb RAM** and **250-500gb of free storage**. The storage will increase overtime, but with the suggested pruning and current state of the chain it's fine and it will be fine for a few more months. <br>
<br>


#### **Automated Removal of Neutaro Validator**

To automatically remove the Neutaro installation, use this command:

```shell
bash <(wget -qO- https://raw.githubusercontent.com/Neutaro/Neutaro/main/neutaro_remove.sh)
```

## Validator's Best Friend: The Neutaro Help Command
Before diving into specific commands, remember that the most valuable tool for any validator is the Neutaro help command. This command provides a comprehensive list of all available commands and their descriptions, making it easier for validators to find exactly what they need. Use it regularly to stay updated:

```shell
Neutaro help
```
This will display all the commands, their descriptions, and available flags for customization.

# Step 1: Manual Setup - Update System and Install Dependencies
We suggest using **Ubuntu 22.04.03**, **4 cores, 8gb RAM** and **250-500gb of free storage**. The storage will increase overtime, but with the suggested pruning and current state of the chain it's fine and it will be fine for a few more months. <br>
<br>

## Step 1: Install Required Dependencies
Run the following command to install essential dependencies: <br>
```shell
sudo apt update && sudo apt install -y \
    curl tar wget clang pkg-config libssl-dev jq build-essential \
    bsdmainutils git make ncdu gcc chrony liblz4-tool pv
```


# Step 2: Install Go
```shell
GO_VERSION="1.22.2"
wget "https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go$GO_VERSION.linux-amd64.tar.gz"
rm "go$GO_VERSION.linux-amd64.tar.gz"
```

# Ensure Go is accessible for both user and sudo
```shell
echo "export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH" | tee -a ~/.bash_profile ~/.bashrc
source ~/.bash_profile
```

# Fix sudo not recognizing Go
```shell
sudo ln -sf /usr/local/go/bin/go /usr/bin/go
```

## Verify Go installation:
```shell
go version
sudo go version
```


## Step 3: Clone and Build Neutaro
```shell
cd $HOME
git clone https://github.com/Neutaro/Neutaro
cd Neutaro
make build
```

Check the Neutaro version:
```shell
./build/Neutaro version
```


## Step 4: Install Cosmovisor
```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.4.0
mkdir -p $HOME/.Neutaro/cosmovisor/genesis/bin
cp build/Neutaro $HOME/.Neutaro/cosmovisor/genesis/bin
ln -s $HOME/.Neutaro/cosmovisor/genesis $HOME/.Neutaro/cosmovisor/current
sudo ln -s $HOME/.Neutaro/cosmovisor/current/bin/Neutaro /usr/local/bin/Neutaro
```


## Step 5: Initialize the Node
Replace `YourMonikerName` with your desired moniker:
```shell
MONIKER="YourMonikerName"
Neutaro init $MONIKER --chain-id Neutaro-1
```


## Step 6: Configure the Node
Edit configuration files using `sed` commands:
```shell
sed -i "s/^seeds *=.*/seeds = \"84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:26656,0e24a596dc34e7063ec2938baf05d09b374709e6@109.199.106.233:26656\"/" $HOME/.Neutaro/config/config.toml
sed -i "s/^pruning *=.*/pruning = \"custom\"/" $HOME/.Neutaro/config/app.toml
sed -i "s/^pruning-keep-recent *=.*/pruning-keep-recent = \"100\"/" $HOME/.Neutaro/config/app.toml
sed -i "s/^pruning-interval *=.*/pruning-interval = \"19\"/" $HOME/.Neutaro/config/app.toml
```

Download the genesis file:
```shell
curl -f http://154.26.153.186/genesis.json > ~/.Neutaro/config/genesis.json
```


## Step 7: Apply Snapshot
Download and extract the snapshot:
```shell
SNAPSHOT_URL="http://173.212.198.246/snapshot-neutaro/latest.tar.lz4"
cd $HOME/.Neutaro
wget $SNAPSHOT_URL -O latest.tar.lz4
lz4 -d latest.tar.lz4 | tar -xvf - -C $HOME/.Neutaro
rm -f latest.tar.lz4
```

Check the extracted files:
```shell
ls -l $HOME/.Neutaro
```
## Step 8: Upgrade to the Latest Version
```shell
cd $HOME/Neutaro
git fetch --all --tags
git checkout v2.0.0
make build
```

Move the new binary to the Cosmovisor upgrade directory:
```shell
mkdir -p $HOME/.Neutaro/cosmovisor/upgrades/v2/bin
cp build/Neutaro $HOME/.Neutaro/cosmovisor/upgrades/v2/bin
```

Verify the upgrade:
```shell
$HOME/.Neutaro/cosmovisor/upgrades/v2/bin/Neutaro version
```


## Step 9: Configure the Systemd Service
Create the `Neutaro.service` file using `nano`:
```shell
sudo nano /etc/systemd/system/Neutaro.service
```

To get your `<your-username>` use this command in terminal:
`whoami`

Replace all occurrences of `<your-username>` in the following parts of the command:

User=`<your-username>`
[e.g. User=BobSmith ]

/home/`<your-username>`/go/bin/cosmovisor
[e.g. /home/BobSmith/go/bin/cosmovisor ]

Environment="DAEMON_HOME=/home/`<your-username>`/.Neutaro"
[e.g.  Environment="DAEMON_HOME=/home/BobSmith/.Neutaro ]


### Replace `<your-username>` in 3 places:
```shell
sudo tee /etc/systemd/system/Neutaro.service > /dev/null << EOF
[Unit]
Description=Neutaro Node Service
After=network-online.target

[Service]
User=<your-username>
ExecStart=/home/<your-username>/go/bin/cosmovisor run start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment="DAEMON_HOME=/home/<your-username>/.Neutaro"
Environment="DAEMON_NAME=Neutaro"
Environment="UNSAFE_SKIP_BACKUP=true"

[Install]
WantedBy=multi-user.target
EOF
```


## Step 10: Enable and Start the Service
Reload the systemd daemon and enable the service:
```shell
sudo systemctl daemon-reload
sudo systemctl enable Neutaro
sudo systemctl restart Neutaro
```

## Note
It can take up to 3h before it starts to sync.

Check the logs:
```shell
sudo journalctl -fu Neutaro -o cat
```


## Step 11: Verify Sync Status
Run the following command to verify the node's sync status:
```shell
Neutaro status 2>&1 | jq .SyncInfo
```


## Step 12: Become a Validator
Create or recover a wallet:
```shell
sudo Neutaro keys add WALLET --keyring-backend os --recover
```

Send the validator transaction. **Make sure you are fully synced before you send the command. (update parameters as needed in example below):**
```shell
Neutaro tx staking create-validator \
--amount=1000000uneutaro \
--pubkey=$(Neutaro tendermint show-validator) \
--moniker="YourName" \
--chain-id=Neutaro-1 \
--from=WALLET \
--keyring-backend=os \
--commission-rate="0.10" \
--commission-max-rate="0.20" \
--commission-max-change-rate="0.01" \
--min-self-delegation="1000000" \
--details="Your validator details"
```


Congratulations! Your Neutaro validator should now be set up and running.
