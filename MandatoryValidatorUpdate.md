# Action required: Mandatory ONLY for running Validators Update
### ALL Validators need to update the binary of their already installed/running nodes.

### If you plan to setup a full Neutaro validator go for full guide under README.MD

We have some outdated modules in the current deployment which need updating as soon as possible across all validators.

### Preparation
Before starting please make sure you have a backup in place. As a minimum ensure you have a backup: 

⁠validator-sdk-update⁠

Even better - have another Validator node synced and ready for failover in case of issues.

## Ensure Go is accessible for both user and sudo
```shell
echo "export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH" | tee -a ~/.bash_profile ~/.bashrc
source ~/.bash_profile
```

## Fix sudo not recognizing Go
```shell
sudo ln -sf /usr/local/go/bin/go /usr/bin/go
```

## Verify Go installation:
```shell
go version
sudo go version
```

## Download & Build the New Neutaro Binary

Create an Upgrade Directory
```shell
mkdir -p ~/Upgrade
cd ~/Upgrade
```

## Clone the Velocent Neutaro Repository
```shell
git clone https://github.com/Neutaro/Neutaro
cd Neutaro
```

##Build the New Neutaro Binary
```shell
make build
```

##Verify the New Binary
```shell
./build/Neutaro version
```

## :white_check_mark: Check that it says
`cosmos_sdk_version: v0.47.15
go: go version go1.22.2 linux/amd64
name: '"neutaro"'
server_name: '"Neutaro"'
version: 2.0.1 in the bottom
`

## Stop the Validator
```shell
sudo systemctl stop Neutaro
```

## :white_check_mark: Ensure the service is fully stopped before proceeding.

Move the New Binary to Cosmovisor
```shell
mv build/Neutaro ~/.Neutaro/cosmovisor/current/bin/Neutaro
```

## Check That the Binary Was Moved Correctly
```shell
ls -lh ~/.Neutaro/cosmovisor/current/bin/Neutaro
```

## :white_check_mark: Ensure the file exists and has the correct permissions.

### Start the Validator
```shell
sudo systemctl start Neutaro
```

## Verify That the Validator is Running


```shell
sudo systemctl status Neutaro
```

## :white_check_mark: You should see "active (running)" in the output.

### Check Logs in Real-Time
```shell
sudo journalctl -fu Neutaro -o cat
```

## :white_check_mark: No errors should appear.

 ### Verify That the Node is Syncing Properly
```shell
Neutaro status 2>&1 | jq .SyncInfo
```

## :white_check_mark: Ensure that "catching_up": false appears, meaning the node is fully synced.

### Final Steps: Ensure Everything is Working

If any issues arise, check logs again:
```shell
sudo journalctl -fu Neutaro -o cat
```

## If Cosmovisor doesn’t start automatically:
```shell
sudo systemctl restart Neutaro
```
