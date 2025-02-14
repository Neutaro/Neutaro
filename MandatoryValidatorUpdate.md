**Mandatory Update for All Neutaro Validators**  

All validators **must** update their binary to ensure the stability and security of their nodes. If you are setting up a **new** Neutaro validator, refer to the **full setup guide** in `README.md`.

### **Why This Update is Required**
Some outdated modules in the current deployment need to be updated **immediately** to maintain network integrity.  

### **Preparation**
Before proceeding, ensure you have a **backup** of your validator setup. At a minimum, back up:  

:warning: 1. Preparation (Backup Important Files)
Before updating, make sure you back up critical files to avoid data loss.

Backup the Validator Keys

```bash
cp ~/.Neutaro/config/priv_validator_key.json ~/priv_validator_key_backup.json
cp ~/.Neutaro/config/node_key.json ~/node_key_backup.json
```

Verify backup:
```bash
ls -lh ~ | grep priv_validator_key_backup.json
ls -lh ~ | grep node_key_backup.json
```
:white_check_mark: If both files appear, your backup is successful.

For added security, we recommend having a **secondary, fully synced validator node** ready as a failover in case of issues.

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

## Download & Build the New Neutaro Binary ✅ **`validator-sdk-update`**  

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

## Build the New Neutaro Binary
```shell
make build
```

## Verify the New Binary
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
`-rwxrwxr-x 1 <your-user-name> 68M feb 13 19:42 /home/<your-user-name>/.Neutaro/cosmovisor/current/bin/Neutaro`
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
