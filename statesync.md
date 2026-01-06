# WHY NEUTARO

Neutaro works closely with **Timpi** to help build the **first truly decentralized search engine**.

On Neutaro, you can:

* Earn rewards by contributing to Timpi
* Stake and secure the network
* Vote on governance proposals affecting Timpi, including ethical and operational decisions for the Timpi search engine

---

## SECURITY (READ FIRST)

Before running infrastructure, review the official security guide:
[https://github.com/Neutaro/Neutaro/blob/main/Security%20Guide.md](https://github.com/Neutaro/Neutaro/blob/main/Security%20Guide.md)

---

## HOW YOU CAN PARTICIPATE

You can participate in Neutaro by:

* Holding NTMPI
* Delegating tokens
* Running a full node
* Becoming a validator

---

## TOKEN UNITS

```
1 NTMPI = 1,000,000 uneutaro
```

---

## DELEGATING TOKENS (OPTIONAL)

Example: delegate **1 NTMPI** to a validator:

```bash
Neutaro tx staking delegate <VALIDATOR_ADDRESS> 1000000uneutaro \
  --from YOUR_WALLET \
  --chain-id Neutaro-1
```

---

# BEFORE RUNNING A NODE

## OPEN REQUIRED PORT

Port **26656/TCP** must be open to allow inbound peers.

### Linux (UFW)

```bash
sudo ufw allow 26656/tcp
sudo ufw reload
```

### Router

Forward **TCP 26656** to your node’s local IP.

---

## SYSTEM REQUIREMENTS

Recommended minimum:

* Ubuntu 22.04 LTS
* 4 CPU cores
* 8 GB RAM
* 250–500 GB SSD

---

# FULL, CLEAN, WORKING NEUTARO NODE GUIDE

*(root AND normal user – reinstall-safe)*

This guide:

* Uses **ONE `$HOME` variable**
* Uses **Cosmovisor + systemd**
* Supports **State Sync (recommended)** or **Snapshot**
* Is safe to re-run from scratch

> Run everything in the **same terminal session**

---

## OPTIONAL: FULL CLEANUP (SAFE REINSTALL)

Run **ONLY** if reinstalling or fixing a broken setup.

```bash
sudo systemctl stop Neutaro 2>/dev/null || true
sudo systemctl disable Neutaro 2>/dev/null || true
sudo rm -f /etc/systemd/system/Neutaro.service
sudo systemctl daemon-reload
sudo systemctl reset-failed

rm -rf ~/.Neutaro ~/Neutaro
sudo rm -f /usr/local/bin/Neutaro /usr/local/bin/cosmovisor
rm -f ~/go/bin/Neutaro ~/go/bin/cosmovisor
rm -r go
```

---

## 0) SANITY CHECK

```bash
set -euo pipefail
echo "User: $(whoami)"
echo "HOME: $HOME"
```

---

## 1) INSTALL DEPENDENCIES

```bash
sudo apt update && sudo apt install -y \
  curl wget git make jq build-essential \
  clang pkg-config libssl-dev \
  chrony lz4 pv
```

---

## 2) INSTALL GO (SYSTEM-WIDE)

```bash
GO_VERSION="1.22.2"
cd /tmp
wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
rm -f go${GO_VERSION}.linux-amd64.tar.gz

grep -q '/usr/local/go/bin' ~/.bashrc || \
  echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' >> ~/.bashrc

export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH
go version
```

---

## 3) BUILD NEUTARO

```bash
cd "$HOME"
rm -rf Neutaro
git clone https://github.com/Neutaro/Neutaro
cd Neutaro
make build
```

Verify:

```bash
ls -lah ./build/Neutaro
./build/Neutaro version --long
```

❌ If the binary does not exist — STOP.

---

## 4) INSTALL COSMOVISOR

```bash
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.4.0
sudo ln -sf "$HOME/go/bin/cosmovisor" /usr/local/bin/cosmovisor
```

---

## 5) COSMOVISOR LAYOUT

```bash
mkdir -p "$HOME/.Neutaro/cosmovisor/genesis/bin"
mkdir -p "$HOME/.Neutaro/data-backup"

cp "$HOME/Neutaro/build/Neutaro" \
   "$HOME/.Neutaro/cosmovisor/genesis/bin/Neutaro"

chmod +x "$HOME/.Neutaro/cosmovisor/genesis/bin/Neutaro"

ln -sfn "$HOME/.Neutaro/cosmovisor/genesis" \
        "$HOME/.Neutaro/cosmovisor/current"

sudo ln -sf "$HOME/.Neutaro/cosmovisor/current/bin/Neutaro" \
            /usr/local/bin/Neutaro
```

---

## 6) INIT NODE

```bash
MONIKER="YourMoniker"
Neutaro init "$MONIKER" --chain-id Neutaro-1
```

### GENESIS

```bash
curl -fsSL http://154.26.153.186/genesis.json \
  > "$HOME/.Neutaro/config/genesis.json"
```

### SEEDS + PRUNING

```bash
CONFIG="$HOME/.Neutaro/config/config.toml"
APP="$HOME/.Neutaro/config/app.toml"

sed -i 's|^seeds *=.*|seeds = "84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:36656"|' "$CONFIG"

sed -i \
  -e 's/^pruning *=.*/pruning = "custom"/' \
  -e 's/^pruning-keep-recent *=.*/pruning-keep-recent = "100"/' \
  -e 's/^pruning-interval *=.*/pruning-interval = "19"/' \
  "$APP"
```

---

## 7) SYNC THE NODE

### OPTION A — STATE SYNC (RECOMMENDED)

```bash
sed -i 's|^persistent_peers *=.*|persistent_peers = "ee64e5d0c3549fe807149f5f29a2913074e08a62@147.93.4.184:26656"|' "$CONFIG"
```

```bash
cat > "$HOME/state_sync.sh" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
CONFIG="$HOME/.Neutaro/config/config.toml"
RPC1="https://rpc2.neutaro.io:443"
RPC2="https://rpc3.neutaro.io:443"
RPC="$RPC1"
curl -fsS "$RPC/status" >/dev/null || RPC="$RPC2"
HEIGHT=$(curl -s "$RPC/block" | jq -r .result.block.header.height)
TRUST_HEIGHT=$((HEIGHT-2000))
TRUST_HASH=$(curl -s "$RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)
sed -i \
  -e 's|^enable *=.*|enable = true|' \
  -e "s|^rpc_servers *=.*|rpc_servers = \"$RPC1,$RPC2\"|" \
  -e "s|^trust_height *=.*|trust_height = $TRUST_HEIGHT|" \
  -e "s|^trust_hash *=.*|trust_hash = \"$TRUST_HASH\"|" \
  "$CONFIG"
EOF

chmod +x "$HOME/state_sync.sh"
"$HOME/state_sync.sh"

Neutaro tendermint unsafe-reset-all --home "$HOME/.Neutaro" --keep-addr-book
```

---

### OPTION B — SNAPSHOT (If you did State sync step skip this step)

```bash
cd "$HOME/.Neutaro"
wget -O latest.tar.lz4 http://173.212.198.246/snapshot-neutaro/latest.tar.lz4
lz4 -d latest.tar.lz4 | tar -xvf -
rm -f latest.tar.lz4
```

---

## 8) SYSTEMD SERVICE

```bash
sudo tee /etc/systemd/system/Neutaro.service > /dev/null << EOF
[Unit]
Description=Neutaro Node Service
After=network-online.target

[Service]
User=$(whoami)
ExecStart=/usr/local/bin/cosmovisor run start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment=DAEMON_HOME=$HOME/.Neutaro
Environment=DAEMON_NAME=Neutaro
Environment=DAEMON_DATA_BACKUP_DIR=$HOME/.Neutaro/data-backup
Environment=UNSAFE_SKIP_BACKUP=true

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable Neutaro
sudo systemctl restart Neutaro
sudo journalctl -fu Neutaro -o cat
```

---

## 9) VERIFY SYNC

```bash
Neutaro status 2>&1 | jq .SyncInfo
```

Wait for:

```json
"catching_up": false
```

---

## 10) DISABLE STATE SYNC (AFTER SYNC)

```bash
sed -i \
  -e 's|^enable *=.*|enable = false|' \
  -e 's|^rpc_servers *=.*|rpc_servers = ""|' \
  -e 's|^trust_height *=.*|trust_height = 0|' \
  -e 's|^trust_hash *=.*|trust_hash = ""|' \
  "$HOME/.Neutaro/config/config.toml"

sudo systemctl restart Neutaro
```

---

## 11) CREATE VALIDATOR (FINAL STEP)

### CREATE / RECOVER WALLET

```bash
Neutaro keys add WALLET --keyring-backend os --recover
```

### CREATE VALIDATOR

(ONLY after fully synced)

```bash
Neutaro tx staking create-validator \
  --amount=1000000uneutaro \
  --pubkey=$(Neutaro tendermint show-validator) \
  --moniker="YourMoniker" \
  --chain-id Neutaro-1 \
  --from=WALLET \
  --keyring-backend=os \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000" \
  --details="Your validator description"
```

---

## QUICK TROUBLESHOOTING

```bash
which Neutaro
which cosmovisor
systemctl status Neutaro --no-pager
ls -lah ~/.Neutaro/config
```

---

## ✅ DONE
