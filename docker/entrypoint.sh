#!/bin/sh
# Neutaro node entrypoint — non-signing full node by default.
# The consensus key generated here is fresh and unknown to the chain: this node
# CANNOT sign for any validator. Never copy a real validator key into a container
# without reading Instructions/NeutaroDocker.md first.
set -eu

HOME_DIR="${NEUTARO_HOME:-/neutaro/.Neutaro}"
MONIKER="${MONIKER:-docker-node}"
CHAIN_ID="${CHAIN_ID:-Neutaro-1}"
SEEDS="${SEEDS:-84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:36656}"
PERSISTENT_PEERS="${PERSISTENT_PEERS:-d6c8714a14d6f5c99756b22b7fade065b2cae56b@100.42.180.106:26656,10333501f9b6204b240a1581e8fc2fe84f704622@213.199.49.225:26656,2fd06277f46e845ca73df8f81caf68e6579bbe32@86.48.20.122:26656,d891af90afdcf3973f7dc44eef316dfe652ccb1c@38.242.135.246:26656,5225956661fa6bad61d7681fdb70174eb24f35e4@173.212.198.192:26656}"
STATESYNC="${STATESYNC:-auto}"        # auto | on | off
TRUST_LAG="${TRUST_LAG:-10000}"
GENESIS_URL="${GENESIS_URL:-http://154.26.153.186/genesis.json}"
GENESIS_SHA256="${GENESIS_SHA256:-78724fe90e5bd1f2abd0186bd0c33325e3c190d8e417741a42ff0c7fbdf2fc2d}"
RPC1="${RPC1:-https://rpc2.neutaro.io:443}"
RPC2="${RPC2:-https://rpc2.neutaro.io:443}"   # rpc3 retired 2026-08-26; second entry may duplicate the first

CFG="$HOME_DIR/config/config.toml"
APP="$HOME_DIR/config/app.toml"

if [ ! -f "$CFG" ]; then
    echo "[entrypoint] first boot: init $MONIKER on $CHAIN_ID"
    Neutaro init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null 2>&1

    sed -i "s|^seeds *=.*|seeds = \"$SEEDS\"|" "$CFG"
    sed -i "s|^persistent_peers *=.*|persistent_peers = \"$PERSISTENT_PEERS\"|" "$CFG"
    sed -i "s|^laddr *=.*26657.*|laddr = \"tcp://0.0.0.0:26657\"|" "$CFG"
    sed -i \
      -e "s|^pruning *=.*|pruning = \"custom\"|" \
      -e "s|^pruning-keep-recent *=.*|pruning-keep-recent = \"100\"|" \
      -e "s|^pruning-interval *=.*|pruning-interval = \"10\"|" \
      -e "s|^min-retain-blocks *=.*|min-retain-blocks = 500000|" \
      -e "s|^minimum-gas-prices *=.*|minimum-gas-prices = \"0uneutaro\"|" \
      "$APP"

    echo "[entrypoint] fetching genesis"
    curl -fsS "$GENESIS_URL" -o "$HOME_DIR/config/genesis.json"
    GOT=$(sha256sum "$HOME_DIR/config/genesis.json" | cut -d' ' -f1)
    if [ "$GOT" != "$GENESIS_SHA256" ]; then
        echo "[entrypoint] FATAL: genesis sha256 mismatch: got $GOT" >&2
        exit 1
    fi
    echo "[entrypoint] genesis verified: $GOT"
fi

WANT_SYNC=0
if [ "$STATESYNC" = "on" ]; then WANT_SYNC=1; fi
if [ "$STATESYNC" = "auto" ] && [ ! -d "$HOME_DIR/data/state.db" ]; then WANT_SYNC=1; fi

if [ "$WANT_SYNC" = "1" ]; then
    echo "[entrypoint] priming state sync (TRUST_LAG=$TRUST_LAG)"
    RPC="$RPC1"
    curl -fsS -m 10 "$RPC/status" >/dev/null 2>&1 || RPC="$RPC2"
    HEIGHT=$(curl -fsS -m 10 "$RPC/block" | jq -r .result.block.header.height)
    [ -n "$HEIGHT" ] && [ "$HEIGHT" -gt "$TRUST_LAG" ] || { echo "[entrypoint] FATAL: bad height '$HEIGHT' from $RPC" >&2; exit 1; }
    TH=$((HEIGHT-TRUST_LAG))
    HASH=$(curl -fsS -m 10 "$RPC/block?height=$TH" | jq -r .result.block_id.hash)
    case "$HASH" in ""|null) echo "[entrypoint] FATAL: empty trust hash at $TH" >&2; exit 1;; esac
    # section-scoped edits: only touch keys inside [statesync]
    sed -i "/^\[statesync\]/,/^\[/ s|^enable *=.*|enable = true|" "$CFG"
    sed -i "/^\[statesync\]/,/^\[/ s|^rpc_servers *=.*|rpc_servers = \"$RPC1,$RPC2\"|" "$CFG"
    sed -i "/^\[statesync\]/,/^\[/ s|^trust_height *=.*|trust_height = $TH|" "$CFG"
    sed -i "/^\[statesync\]/,/^\[/ s|^trust_hash *=.*|trust_hash = \"$HASH\"|" "$CFG"
    echo "[entrypoint] primed: tip=$HEIGHT trust_height=$TH"
else
    sed -i "/^\[statesync\]/,/^\[/ s|^enable *=.*|enable = false|" "$CFG"
fi

echo "[entrypoint] starting Neutaro"
exec Neutaro start --home "$HOME_DIR"
