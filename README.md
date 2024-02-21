# Why you should care about Neutaro
Neutaro is closely working with Timpi to help them create the first truly decentralized search engine! On Neutaro you can claim your rewards for contributing to Timpi and you can vote on different proposals affecting Timpi. These proposals will for example be about the ethical standpoint of the Timpi search engine.

# Introduction
You can be a part of Neutaro by having tokens and delegating those, running a node or by becoming a validator.

# Delegating
By delegating your tokens to a validator you increase the amount of staked tokens they have and by that their voting power. In return you get a cut of their rewards based on the amount you have delegated. You can delegate to a Validator using the command. 1.000.000 uneutaro is 1 NTMPI. if you want to stake 100 NTMPI tokens you put **100000000uneutaro** into the command. The example is 1 token.

```shell
Neutaro tx staking delegate ValidatorAddress 1000000uneutaro --from YOURWALLET --chain-id Neutaro-1
```

# Running a node
Running a node means that you run the chains binary. Follow these steps to create a Validator that runs as a service on linux.

### Preperation
We suggest using Ubuntu 22.04.03, 4 cores and 100gb of free storage. This will increase overtime, but with the suggested pruning and current state of the chain it's fine and it will be fine for a few more months. <br>
<br>
Make sure your system is up to date. <br>
```shell
sudo apt update && sudo apt upgrade -y && sudo apt install curl tar wget clang pkg-config libssl-dev jq build-essential bsdmainutils git make ncdu gcc git jq chrony liblz4-tool -y
```

### Installing Go v1.20.4
```shell
ver="1.20.4"

cd $HOME
wget "https://golang.org/dl/go$ver.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go$ver.linux-amd64.tar.gz"
rm "go$ver.linux-amd64.tar.gz"
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> $HOME/.bash_profile
source $HOME/.bash_profile
```
<br>
Make sure version 1.20.4 is installed <br>

```shell
go version
```

### Then we install the Neutaro Binary
```shell
cd $HOME
git clone https://github.com/Neutaro/Neutaro
cd Neutaro/cmd/Neutaro/
go build
```

### Installing cosmovisor
```shell
mkdir -p $HOME/.Neutaro/cosmovisor/genesis/bin
mv $HOME/Neutaro/cmd/Neutaro/Neutaro $HOME/.Neutaro/cosmovisor/genesis/bin/
sudo ln -s $HOME/.Neutaro/cosmovisor/genesis $HOME/.Neutaro/cosmovisor/current
sudo ln -s $HOME/.Neutaro/cosmovisor/current/bin/Neutaro /usr/local/bin/Neutaro
cd  $HOME/Neutaro/
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.4.0
```

### Now we configure the node
Make sure to replace "YOURMONIKER" with your actual one.
```shell
MONIKER=YOURMONIKER
```
```shell
Neutaro init $MONIKER --chain-id Neutaro-1
Neutaro config chain-id Neutaro-1
Neutaro config keyring-backend os
curl http://154.26.153.186/genesis.json > $HOME/.Neutaro/config/genesis.json
```
### Editing the app.toml file
You can edit the file $HOME/.Neutaro/config/app.toml to change the minimum-gas-prices ( Currently most use 0 ) and the pruning. <br>
minimum-gas-prices = "0uneutaro" <br>
<br>
For pruning you could add whatever you like. These options mainly decide how much storage the Node will use. An example would be using<br>
pruning="custom" <br>
pruning-keep-recent="100" <br>
pruning-interval="19" <br>

### Configuring the config.toml
The file is in $HOME/.Neutaro/config/config.toml. Here you edit the "seeds" to seeds = "0e24a596dc34e7063ec2938baf05d09b374709e6@109.199.106.233:26656". <br>

### Downloading the snapshot
```shell
cd $Home/root/.Neutaro/
mv data data-old
wget http://109.199.106.233/snapshotNeutaro.tar.lz4
lz4 -d snapshotNeutaro.tar.lz4
tar -xf snapshotNeutaro.tar
```
Once the node is running you can delete unnecessary files using
```shell
rm -r snapshotNeutaro.tar
rm -r snapshotNeutaro.tar.lz4
rm -r data-old
```

### Create Neutaro service.
```shell
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
```

### Enabling the Service
```shell
sudo systemctl daemon-reload
```
```shell
sudo systemctl enable Neutaro
```

### Starting the Service/ the Node
```shell
sudo systemctl restart Neutaro
```
### To view the service use
```shell
sudo journalctl -fu Neutaro -o cat
```
<br>
use ctrl + c to exit the log

### Check the sync status
```shell
Neutaro status 2>&1 | jq .SyncInfo
```

### **Proceed once it's synced**
you will be asked for your memonic on this step. You can also remove the --recover flag and create a new wallet and send funds to this new wallet from your main wallet. You can now delete the files using the commands from the snapshot section. <br>
<br>
```shell
Neutaro keys add WALLET --keyring-backend os --recover
```

### Sending the becomming a validator transaction
once you have a funded wallet on the node send this, **__but make sure to check all the parameters to see if they are fine for you!__** <br>
<br>
```shell
Neutaro tx staking create-validator --amount=1000000uneutaro --pubkey=$(Neutaro tendermint show-validator) --moniker=$MONIKER --chain-id=Neutaro-1 --from WALLET --keyring-backend os --commission-rate="0.10" --commission-max-rate="0.20" --commission-max-change-rate="0.01" --min-self-delegation="1000000" --gas="auto" --gas-prices="0.0025uneutaro" --gas-adjustment="1.5"
```
