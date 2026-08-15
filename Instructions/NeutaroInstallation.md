# Neutaro — Node & Validator Installation

This guide takes you from a clean Ubuntu server to a running Neutaro node, and optionally a
validator. Every command was verified against `Neutaro-1` in August 2026.

**Chain:** `Neutaro-1` · **Binary:** `Neutaro` · **Token:** `1 NTMPI = 1,000,000 uneutaro`

| You want | Read |
|---|---|
| A full node or validator, step by step | this guide, top to bottom |
| The fastest possible sync (~10 minutes) | this guide through step 6, then [`statesync.md`](../statesync.md) §8 |
| Day-to-day validator commands | [`NeutaroValidatorCommands.md`](NeutaroValidatorCommands.md) |
| Security hardening | [`SecurityGuide.md`](../SecurityGuide.md) |
| To back up your keys (do it early) | [`statesync.md`](../statesync.md) §1 — including how to *verify* the backup |

## Contents

- [1. Requirements](#1-requirements)
- [2. Ports and firewall](#2-ports-and-firewall)
- [3. Dependencies](#3-dependencies)
- [4. Go](#4-go)
- [5. Build Neutaro + cosmovisor](#5-build-neutaro--cosmovisor)
- [6. Initialize and configure](#6-initialize-and-configure)
- [7. Sync the node](#7-sync-the-node)
- [8. systemd service](#8-systemd-service)
- [9. Verify](#9-verify)
- [10. Become a validator](#10-become-a-validator)
- [11. Removing an installation](#11-removing-an-installation)

---

## 1. Requirements

* Ubuntu 22.04 LTS, 4+ cores, 8 GB+ RAM
* Disk: **~40 GB** for a state-synced pruned node, **~250 GB+** if you sync from a snapshot and
  keep history
* A user with sudo. **Do not run the node as root**, and — important later — **never run
  `Neutaro keys` commands with `sudo`** (see step 10).

## 2. Ports and firewall

| Port | Purpose | Expose? |
|---|---|---|
| 26656/tcp | P2P — blocks, peers, snapshots | **Yes** — firewall *and* router if behind NAT |
| 26657/tcp | RPC | Only if you deliberately serve RPC |
| 1317, 9090 | REST / gRPC | Localhost unless deliberately published |

```shell
sudo ufw allow 26656/tcp
sudo ufw reload
```

## 3. Dependencies

```shell
sudo apt update && sudo apt install -y \
    curl tar wget clang pkg-config libssl-dev jq build-essential \
    bsdmainutils git make ncdu gcc chrony liblz4-tool pv
```

`chrony` is not optional — peers reject nodes with a skewed clock.

## 4. Go

```shell
GO_VERSION="1.22.2"
cd /tmp
wget "https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go$GO_VERSION.linux-amd64.tar.gz"
rm "go$GO_VERSION.linux-amd64.tar.gz"

grep -q '/usr/local/go/bin' ~/.bashrc || \
  echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' >> ~/.bashrc
export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH
go version   # go version go1.22.2 linux/amd64
```

## 5. Build Neutaro + cosmovisor

```shell
cd $HOME
git clone https://github.com/Neutaro/Neutaro
cd Neutaro
make build
./build/Neutaro version --long | grep -E 'version|commit|cosmos_sdk'
```

Install cosmovisor and lay out the directories:

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.4.0

mkdir -p $HOME/.Neutaro/cosmovisor/genesis/bin
mkdir -p $HOME/.Neutaro/cosmovisor/upgrades/v2/bin
cp build/Neutaro $HOME/.Neutaro/cosmovisor/genesis/bin
cp build/Neutaro $HOME/.Neutaro/cosmovisor/upgrades/v2/bin
ln -sfn $HOME/.Neutaro/cosmovisor/genesis $HOME/.Neutaro/cosmovisor/current
sudo ln -sf $HOME/.Neutaro/cosmovisor/current/bin/Neutaro /usr/local/bin/Neutaro
```

## 6. Initialize and configure

```shell
MONIKER="YourMonikerName"
Neutaro init "$MONIKER" --chain-id Neutaro-1
```

Configure — the seed, pruning, and a real gas denom:

```shell
CONFIG="$HOME/.Neutaro/config/config.toml"
APP="$HOME/.Neutaro/config/app.toml"

sed -i 's|^seeds *=.*|seeds = "84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:36656"|' "$CONFIG"

sed -i \
  -e 's|^pruning *=.*|pruning = "custom"|' \
  -e 's|^pruning-keep-recent *=.*|pruning-keep-recent = "100"|' \
  -e 's|^pruning-interval *=.*|pruning-interval = "10"|' \
  -e 's|^min-retain-blocks *=.*|min-retain-blocks = 500000|' \
  -e 's|^minimum-gas-prices *=.*|minimum-gas-prices = "0uneutaro"|' \
  "$APP"
```

Why these values:

* The seed runs on port **36656** — a dedicated seed process, not the full node on 26656.
* `min-retain-blocks = 500000` keeps ~1 month of blocks. **The default `0` keeps every block
  forever and will eventually fill any disk.**
* `minimum-gas-prices` defaults to `"0stake"` — a denom that does not exist on Neutaro.
* If you later want to *serve* state-sync snapshots, `pruning-keep-recent` must satisfy the
  constraint in [`statesync.md`](../statesync.md) §9.1.

Download the genesis and **verify it** — do not skip the checksum:

```shell
curl -f http://154.26.153.186/genesis.json > ~/.Neutaro/config/genesis.json
sha256sum ~/.Neutaro/config/genesis.json
# must print:
# 78724fe90e5bd1f2abd0186bd0c33325e3c190d8e417741a42ff0c7fbdf2fc2d
```

If the hash differs, stop — do not start a node on an unverified genesis.

## 7. Sync the node

Two options. **State sync is the recommended path** — minutes instead of hours, ~40 GB instead of
250 GB+.

### Option A — State sync (recommended)

Follow [`statesync.md`](../statesync.md) **§8** — prime the trust height, start the service
(step 8 below), and the node restores the latest network snapshot in about 10 minutes. Come back
here for step 8 first, since the service must exist before you start it.

### Option B — Snapshot download

```shell
cd $HOME/.Neutaro
SNAPSHOT_URL="http://173.212.198.246/snapshot-neutaro/latest.tar.lz4"
wget "$SNAPSHOT_URL" -O latest.tar.lz4
lz4 -t latest.tar.lz4 && lz4 -d latest.tar.lz4 | tar -xvf - -C $HOME/.Neutaro
rm -f latest.tar.lz4
```

With a snapshot it can take a while before the node starts syncing — see
[`statesync.md`](../statesync.md) §11.4b: a node with a lot of data looks hung on startup and
is not.

## 8. systemd service

Find your username with `whoami`, then create the service (replace `<your-username>` in **three**
places):

```shell
sudo tee /etc/systemd/system/Neutaro.service > /dev/null << 'EOF'
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

sudo systemctl daemon-reload
sudo systemctl enable Neutaro
sudo systemctl start Neutaro
sudo journalctl -fu Neutaro -o cat
```

## 9. Verify

```shell
Neutaro status | jq .SyncInfo
```

Wait until `"catching_up": false`. With state sync that is minutes; from a snapshot it can be
hours. Expected healthy log lines and every common failure are in
[`statesync.md`](../statesync.md) §8.3 and §11.

## 10. Become a validator

**Back up your keys first** — [`statesync.md`](../statesync.md) §1 shows what to save and how to
*prove* the backup is good. `config/priv_validator_key.json` is your validator identity and can
never be recreated.

Create or recover the wallet. **No `sudo` here** — `sudo` puts the key in root's keyring, and
every later command run without `sudo` will not find it:

```shell
Neutaro keys add WALLET --keyring-backend os            # new wallet
Neutaro keys add WALLET --keyring-backend os --recover  # or restore from seed phrase
```

Once `catching_up` is `false`, send the create-validator transaction:

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
  --gas auto --gas-adjustment 1.4 --gas-prices 0.025uneutaro \
  --details="Your validator details"
```

Day-to-day operations (edit, delegate, vote, unjail):
[`NeutaroValidatorCommands.md`](NeutaroValidatorCommands.md).

Before any maintenance that stops the node, read [`statesync.md`](../statesync.md) §10 — it is the
difference between a recoverable mistake and losing the validator permanently.

## 11. Removing an installation

```shell
bash <(wget -qO- https://raw.githubusercontent.com/Neutaro/Neutaro/main/neutaro_remove.sh)
```

⚠️ This deletes `~/.Neutaro` **including your keys**. Back up
`config/priv_validator_key.json` and your keyring first if there is any chance you will want this
identity again.
