# Neutaro — State Sync & Snapshot Provider Guide

> **What this covers:** installing a Neutaro node, backing up the keys that cannot be replaced,
> catching up in minutes with state sync, serving snapshots back to the network, and reclaiming a
> full disk on a validator without risking it.
>
> Every command here was executed end-to-end against `Neutaro-1` on 2026-08-14, across **two real
> installs**: a clean Ubuntu 22.04 box built into a public snapshot provider (§1–§9), and a **live
> validator with 596,270 voting power** whose full disk was reclaimed by state sync — 183 G → 9.8 G,
> ~9 minutes of downtime, no slashing (§10.3). **Every log line quoted is captured output**, not an
> illustration.

**Chain:** `Neutaro-1` · **Binary:** `Neutaro` · **SDK** v0.47.15 · **CometBFT** v0.37.5 · **Go** 1.22

---

## Contents

- [0. Which guide do you want?](#0-which-guide-do-you-want)
- [1. 🔐 Back up your keys — before anything else](#1--back-up-your-keys--before-anything-else)
  - [1.1 What is actually irreplaceable](#11-what-is-actually-irreplaceable)
  - [1.2 Take the backup](#12-take-the-backup)
  - [1.3 🔴 Now verify it — a copied file is not a proven backup](#13--now-verify-it--a-copied-file-is-not-a-proven-backup)
  - [1.4 ⚠️ `priv_validator_state.json` is the one file you must not restore blindly](#14--priv_validator_statejson-is-the-one-file-you-must-not-restore-blindly)
- [2. Requirements](#2-requirements)
- [3. Optional: clean reinstall](#3-optional-clean-reinstall)
- [4. Dependencies](#4-dependencies)
- [5. Go 1.22](#5-go-122)
- [6. Build + cosmovisor](#6-build--cosmovisor)
- [7. Init + genesis](#7-init--genesis)
  - [7.1 Editing TOML safely](#71-editing-toml-safely)
  - [7.2 Config](#72-config)
- [8. State sync](#8-state-sync)
  - [8.1 Prime trust height and hash](#81-prime-trust-height-and-hash)
  - [8.2 systemd](#82-systemd)
  - [8.3 Expected logs — the happy path](#83-expected-logs--the-happy-path)
  - [8.4 Verify](#84-verify)
  - [8.5 Turn state sync back off](#85-turn-state-sync-back-off)
- [9. 🔴 Serving state sync to others](#9--serving-state-sync-to-others)
  - [9.1 ⚠️ The pruning/snapshot constraint](#91--the-pruningsnapshot-constraint)
  - [9.2 Confirm you are actually serving](#92-confirm-you-are-actually-serving)
  - [9.3 Publish your endpoint](#93-publish-your-endpoint)
- [10. 🔴 Validator safety, and state-syncing a validator that is already running](#10--validator-safety-and-state-syncing-a-validator-that-is-already-running)
  - [10.1 Know what the two penalties actually cost](#101-know-what-the-two-penalties-actually-cost)
  - [10.2 `unsafe-reset-all` vs `reset-state` — the distinction that saves the validator](#102-unsafe-reset-all-vs-reset-state--the-distinction-that-saves-the-validator)
  - [10.3 Procedure — state-syncing a validator that is already live](#103-procedure--state-syncing-a-validator-that-is-already-live)
  - [10.4 What this costs you](#104-what-this-costs-you)
  - [10.5 The rules, short](#105-the-rules-short)
- [11. Troubleshooting](#11-troubleshooting)
  - [11.1 Health check](#111-health-check)
  - [11.2 `Discovered new snapshot` repeats forever, never `Offering snapshot to ABCI app`](#112-discovered-new-snapshot-repeats-forever-never-offering-snapshot-to-abci-app)
  - [11.3 Verify a peer before you trust the config](#113-verify-a-peer-before-you-trust-the-config)
  - [11.4 `no witnesses connected. please reset light client`](#114-no-witnesses-connected-please-reset-light-client)
  - [11.5 Service will not start](#115-service-will-not-start)
  - [11.6 Disk filling up](#116-disk-filling-up)
  - [11.7 Nobody can state-sync from me](#117-nobody-can-state-sync-from-me)
- [12. Reference](#12-reference)
- [Appendix — what changed from the previous version of this file, and why](#appendix--what-changed-from-the-previous-version-of-this-file-and-why)

---

## 0. Which guide do you want?

| You want | Read |
|---|---|
| **To not lose your keys** | 🔐 **§1 — start here regardless of what else you are doing** |
| A node that catches up in ~10 minutes instead of days | §2–§8 (**Consume state sync**) |
| A node that *serves* state sync to everyone else | §2–§8, then **§9 (Provide state sync)** |
| **An existing node whose disk is full** | **§1**, then **§10.3** — you do not need to rebuild the box |
| A validator | **§10 first**, then this guide, then `Instructions/NeutaroValidatorCommands.md` |
| To state-sync a validator that is already live | **§10.3** — proven procedure, real logs, ~9 min downtime |

State sync fetches a *snapshot of application state* over **P2P (26656)** and verifies it against block
headers pulled over **RPC (26657)**. Both matter, and they do different jobs — that distinction is the
source of most failed installs:

- `rpc_servers` = light-client verification only. Does **not** transfer the snapshot.
- `persistent_peers` / PEX peers = where the snapshot bytes actually come from.

A perfectly reachable `rpc_servers` and zero snapshot-serving peers gives you a node that logs
"Discovered new snapshot" forever and never syncs.

---

## 1. 🔐 Back up your keys — before anything else

Do this **first**, before an install, before a state sync, before touching a disk. It takes ten
seconds and it is the difference between an inconvenience and losing a validator permanently.

### 1.1 What is actually irreplaceable

| File | Lives in | Replaceable? |
|---|---|---|
| `config/priv_validator_key.json` | `config/` | ❌ **Never.** This *is* your validator identity. |
| `keyring-*/` | `~/.Neutaro/` | ❌ **Never.** Your wallets. With `--keyring-backend test` these are **unencrypted**. |
| `config/node_key.json` | `config/` | ✅ P2P identity only — regenerable, you just get a new node ID. |
| `config/*.toml` | `config/` | ✅ Config; back it up to save yourself re-editing. |
| `config/genesis.json` | `config/` | ✅ Public, downloadable, checksummable. |
| `data/priv_validator_state.json` | `data/` | ⚠️ Special — see the warning below. |

Note where they live: **the keys are in `config/`, never in `data/`.** Everything the reset commands
and the state-sync procedure destroy lives in `data/`. That is why state-syncing a node cannot cost
you your keys — but back them up anyway, because disks fail for reasons unrelated to you.

### 1.2 Take the backup

Stream it over SSH so nothing is written to a disk that may be full:

```bash
ssh root@YOUR_NODE 'cd ~/.Neutaro && tar czf - \
    config/priv_validator_key.json config/node_key.json \
    config/config.toml config/app.toml config/client.toml config/genesis.json \
    keyring-test/ data/priv_validator_state.json' \
  > neutaro-keys-$(date +%Y%m%d).tar.gz
```

Encrypt it before it lands anywhere shared. `keyring-test/` is plaintext private keys:

```bash
umask 077
openssl rand -base64 33 > passphrase.txt        # keep this OFF the servers
gpg --batch --symmetric --cipher-algo AES256 --pinentry-mode loopback \
    --passphrase-file passphrase.txt neutaro-keys-$(date +%Y%m%d).tar.gz
```

### 1.3 🔴 Now verify it — a copied file is not a proven backup

Almost nobody does this step, and it is the only one that actually tells you the backup is good.
Derive the consensus address from the backed-up public key and compare it to what the chain thinks
your validator is:

```bash
tar xzf neutaro-keys-*.tar.gz -C /tmp/verify
python3 -c "
import json,hashlib,base64
k=json.load(open('/tmp/verify/config/priv_validator_key.json'))
print('from backup :', hashlib.sha256(base64.b64decode(k['pub_key']['value'])).hexdigest()[:40].upper())
print('in the file :', k['address'])
print('has privkey :', bool(k.get('priv_key',{}).get('value')))
"
# compare against the live node:
Neutaro tendermint show-address           # or /status -> validator_info.address
```

Real output from a verified backup:

```
from backup : F48ABC68C122CFBA7C80308AD2B349CADE6085E4
in the file : F48ABC68C122CFBA7C80308AD2B349CADE6085E4
has privkey : True
```

All three must agree. If they do, that archive can rebuild your validator on bare metal.

### 1.4 ⚠️ `priv_validator_state.json` is the one file you must not restore blindly

It records the last height/round/step your key signed, and exists to stop the key signing twice at
the same height. **A backup copy of it goes stale the instant it is taken.** Restoring an old one
onto a running validator lets the key re-sign heights it already signed — that is a double-sign, and
on Neutaro it costs 5 % of your stake **and permanently tombstones the validator**.

Keep it in the archive for reference. The only copy that is ever safe to restore is one taken
**after** the node has been stopped. See §10.

---

## 2. Requirements

* Ubuntu 22.04 LTS, 4+ cores, 8 GB+ RAM
* Disk: **~40 GB** for a pruned state-sync node, **~250 GB+** if you keep full history
* Outbound to `:443` and `:26656`; inbound **26656/tcp** (and **26657/tcp** if you serve RPC)

Token unit: `1 NTMPI = 1,000,000 uneutaro`

---

## 3. Optional: clean reinstall

Run **only** if you are wiping an existing install.

```bash
sudo systemctl stop Neutaro 2>/dev/null || true
sudo systemctl disable Neutaro 2>/dev/null || true
sudo rm -f /etc/systemd/system/Neutaro.service
sudo systemctl daemon-reload
sudo systemctl reset-failed

rm -rf "$HOME/.Neutaro" "$HOME/Neutaro" "$HOME/go"
sudo rm -f /usr/local/bin/Neutaro /usr/local/bin/cosmovisor
```

> ⚠️ The current guide ends this block with `rm -r go`. That is **relative to your current directory**
> and has no `-f`. Run it from the wrong place and you delete an unrelated `go/`; run it where no `go/`
> exists and it aborts. Always use the quoted absolute `"$HOME/go"` form above.

> ⚠️ Do **not** paste `set -euo pipefail` into an interactive shell (the current guide's step 0 does).
> Any later command returning non-zero — a `grep -q` that finds nothing is enough — closes your
> session mid-install. Use it in *scripts*, not at your prompt.

---

## 4. Dependencies

```bash
sudo apt update && sudo apt install -y \
  curl wget git make jq build-essential \
  clang pkg-config libssl-dev chrony lz4 pv
```

`chrony` is not optional. A node with a skewed clock is rejected by peers.

---

## 5. Go 1.22

```bash
GO_VERSION="1.22.2"
cd /tmp
wget -q "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
rm -f "go${GO_VERSION}.linux-amd64.tar.gz"

grep -q '/usr/local/go/bin' ~/.bashrc || \
  echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' >> ~/.bashrc
export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH
go version   # go version go1.22.2 linux/amd64
```

## 6. Build + cosmovisor

```bash
cd "$HOME" && rm -rf Neutaro
git clone https://github.com/Neutaro/Neutaro && cd Neutaro
make build
./build/Neutaro version --long | grep -E 'version|commit|cosmos_sdk_version'
```

Expected (2026-08-14 `main`):

```
version: 2.0.1-30-g2b302b7
commit: 2b302b7295e6380e9585c602a075a1dc8a608872
cosmos_sdk_version: v0.47.15
```

❌ No `./build/Neutaro`? Stop — everything below will fail confusingly.

```bash
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.4.0
sudo ln -sf "$HOME/go/bin/cosmovisor" /usr/local/bin/cosmovisor

mkdir -p "$HOME/.Neutaro/cosmovisor/genesis/bin" "$HOME/.Neutaro/data-backup"
cp "$HOME/Neutaro/build/Neutaro" "$HOME/.Neutaro/cosmovisor/genesis/bin/Neutaro"
chmod +x "$HOME/.Neutaro/cosmovisor/genesis/bin/Neutaro"
ln -sfn "$HOME/.Neutaro/cosmovisor/genesis" "$HOME/.Neutaro/cosmovisor/current"
sudo ln -sf "$HOME/.Neutaro/cosmovisor/current/bin/Neutaro" /usr/local/bin/Neutaro
```

> `cosmovisor version` on its own prints
> `DAEMON_NAME is not set, DAEMON_HOME is not set, DAEMON_DATA_BACKUP_DIR must not be empty`.
> That is expected — cosmovisor v1.4.0 demands all three even for `version`, and even when
> `UNSAFE_SKIP_BACKUP=true`. The systemd unit in §8 sets them. **Do not delete
> `~/.Neutaro/data-backup`**, or the service will refuse to start.

---

## 7. Init + genesis

```bash
MONIKER="YourMoniker"
Neutaro init "$MONIKER" --chain-id Neutaro-1

curl -fsSL http://154.26.153.186/genesis.json > "$HOME/.Neutaro/config/genesis.json"
sha256sum "$HOME/.Neutaro/config/genesis.json"
```

**Verify before you trust it.** Genesis is served over plaintext HTTP from a bare IP with no published
checksum — a tampered genesis gives you a node that silently follows the wrong chain.

Known-good as of 2026-08-14:

```
sha256  78724fe90e5bd1f2abd0186bd0c33325e3c190d8e417741a42ff0c7fbdf2fc2d
size    32358 bytes
chain_id        Neutaro-1
genesis_time    2023-07-25T10:48:11.525747832Z
initial_height  1
```

```bash
# quick sanity check
python3 -c "import json;g=json.load(open('$HOME/.Neutaro/config/genesis.json'));print(g['chain_id'], g['genesis_time'])"
```

### 7.1 Editing TOML safely

`config.toml` and `app.toml` reuse key names across sections — `app.toml` alone has **11** bare
`enable =` lines. A blind `sed -i 's/^enable.*/enable = true/'` is a coin flip. Today it happens to be
safe on `config.toml` (exactly one bare `^enable`, at line 358 under `[statesync]`), but that is luck,
not design, and it breaks the moment upstream adds a key.

Use a section-aware setter that **fails loudly** rather than editing the wrong line:

```bash
cat > "$HOME/tomlset.py" <<'PY'
#!/usr/bin/env python3
"""Set key=value inside a specific [section]. Usage: tomlset.py FILE 'section:key=value' ...
Section '' = the top-level block. Exits non-zero if the key is missing or ambiguous."""
import sys, re
path, *edits = sys.argv[1:]
lines = open(path).read().split("\n")
def section_of(i):
    for j in range(i, -1, -1):
        m = re.match(r'^\s*\[([^\]]+)\]', lines[j])
        if m: return m.group(1)
    return ""
for edit in edits:
    sec, kv = edit.split(":", 1); key, val = kv.split("=", 1)
    hits = [i for i, ln in enumerate(lines)
            if (m := re.match(r'^\s*(#\s*)?([A-Za-z0-9_\-]+)\s*=', ln))
            and m.group(2) == key and section_of(i) == sec]
    if len(hits) != 1: sys.exit(f"FAIL: [{sec}] {key} matched {len(hits)} lines in {path}")
    print(f"  [{sec or 'top'}] {lines[hits[0]].strip()}  ->  {key} = {val}")
    lines[hits[0]] = f"{key} = {val}"
open(path, "w").write("\n".join(lines))
PY
chmod +x "$HOME/tomlset.py"
```

### 7.2 Config

```bash
CONFIG="$HOME/.Neutaro/config/config.toml"
APP="$HOME/.Neutaro/config/app.toml"

SEEDS="84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:36656"
PEERS="0e24a596dc34e7063ec2938baf05d09b374709e6@109.199.106.233:26656,\
726d5975dd11383a175d1b748526257d3749058c@185.182.184.50:26656,\
95f6fc822469efdf868ab6cdfc218fa80f716951@62.84.180.12:26656,\
90dcafcc67687feff6d1b355892a3690c6cb71f3@185.182.184.8:26656,\
d891af90afdcf3973f7dc44eef316dfe652ccb1c@38.242.135.246:26656"

"$HOME/tomlset.py" "$APP" \
  ':minimum-gas-prices="0uneutaro"' \
  ':pruning="custom"' \
  ':pruning-keep-recent="2000"' \
  ':pruning-interval="10"' \
  ':min-retain-blocks=500000'

"$HOME/tomlset.py" "$CONFIG" \
  "p2p:seeds=\"$SEEDS\"" \
  "p2p:persistent_peers=\"$PEERS\"" \
  'p2p:max_num_inbound_peers=80' \
  'p2p:max_num_outbound_peers=30' \
  'tx_index:indexer="kv"'
```

**Why these differ from the current guide:**

| Setting | Current guide | Here | Reason |
|---|---|---|---|
| `minimum-gas-prices` | *(untouched)* | `0uneutaro` | The generated default is **`"0stake"`** — a denom that does not exist on Neutaro. |
| `pruning-keep-recent` | `100` | `2000` | `100` is fine for a *consumer*, but it **breaks snapshot serving** (see §9). |
| `min-retain-blocks` | *(never mentioned)* | `500000` | Without it CometBFT **never prunes blocks**. This is the single most common way Neutaro nodes fill their disk. |
| `persistent_peers` | one dead host | 5 verified live | See §11.3. |

> 💾 **`min-retain-blocks = 0` (the default) means blocks are kept forever.** On a long-running Neutaro
> node that is `blockstore.db` ≈ 76 GB and `state.db` ≈ 40 GB and climbing. `500000` retains ~34 days
> (~5.83 s/block) for under 3 GB.
>
> ⚠️ `tx_index.db` is pruned by **nothing** — not `pruning`, not `min-retain-blocks`. With
> `indexer = "kv"` budget roughly **40 MB/day (~14 GB/year)** and monitor it. Set
> `tx_index:indexer="null"` if you do not need local `/tx_search`.

---

## 8. State sync

### 8.1 Prime trust height and hash

Save as `~/state_sync.sh`:

```bash
#!/usr/bin/env bash
# Prime CometBFT state sync. Fails loudly instead of writing an empty trust_hash.
set -euo pipefail
CONFIG="$HOME/.Neutaro/config/config.toml"
RPC1="https://rpc2.neutaro.io:443"
RPC2="https://rpc3.neutaro.io:443"
LAG="${TRUST_LAG:-10000}"          # blocks behind tip; MUST be > the provider's snapshot interval

pick_rpc() {
  for r in "$RPC1" "$RPC2"; do
    h=$(curl -fsS -m 10 "$r/status" 2>/dev/null | jq -r '.result.sync_info.latest_block_height // empty')
    if [ -n "${h:-}" ] && [ "$h" -gt 0 ] 2>/dev/null; then echo "$r"; return 0; fi
  done
  echo "FATAL: neither $RPC1 nor $RPC2 answered /status" >&2; exit 1
}

RPC=$(pick_rpc)
HEIGHT=$(curl -fsS -m 10 "$RPC/block" | jq -r .result.block.header.height)
[ -n "$HEIGHT" ] && [ "$HEIGHT" -gt "$LAG" ] || { echo "FATAL: bad height '$HEIGHT'" >&2; exit 1; }
TRUST_HEIGHT=$((HEIGHT-LAG))
TRUST_HASH=$(curl -fsS -m 10 "$RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)
case "$TRUST_HASH" in ""|null) echo "FATAL: empty trust hash at $TRUST_HEIGHT" >&2; exit 1;; esac
[ ${#TRUST_HASH} -eq 64 ] || { echo "FATAL: trust hash not 64 hex chars" >&2; exit 1; }

"$HOME/tomlset.py" "$CONFIG" \
  "statesync:enable=true" \
  "statesync:rpc_servers=\"$RPC1,$RPC2\"" \
  "statesync:trust_height=$TRUST_HEIGHT" \
  "statesync:trust_hash=\"$TRUST_HASH\"" \
  "statesync:trust_period=\"168h0m0s\""
echo "primed from $RPC: tip=$HEIGHT trust_height=$TRUST_HEIGHT trust_hash=$TRUST_HASH"
```

```bash
chmod +x "$HOME/state_sync.sh" && "$HOME/state_sync.sh"
Neutaro tendermint unsafe-reset-all --home "$HOME/.Neutaro" --keep-addr-book
```

#### 🔴 Why `TRUST_LAG` is 10000 and not 2000

The current guide uses `TRUST_HEIGHT=$((HEIGHT-2000))`. **This deadlocks against real providers.**

Providers snapshot every 1000–2000 blocks and keep only the last 2, so the newest snapshot on the
network is routinely ~2000 blocks *behind* the tip. `HEIGHT-2000` then lands your trust height *above
every snapshot that exists*, and the light client cannot verify backwards. Captured live on this
install with the guide's own value:

```
INF Discovered new snapshot format=3 height=16616000 module=statesync
INF Discovered new snapshot format=3 height=16614000 module=statesync
INF Discovered new snapshot format=3 height=16616000 module=statesync      <-- forever
```

`trust_height` was `16617286`. **Both offers were below it**, so `Offering snapshot to ABCI app`
never appeared. `TRUST_LAG=10000` (~16 h, well inside the 168 h `trust_period`) fixes it.

The other half of the fix is **peer diversity** — see §11.3.

> ⚠️ `unsafe-reset-all` also **resets `priv_validator_state.json` to genesis**. On a machine with an
> active validator key that is a double-signing (slashing) risk. See §10.

### 8.2 systemd

```bash
sudo tee /etc/systemd/system/Neutaro.service > /dev/null <<EOF
[Unit]
Description=Neutaro Node
After=network-online.target
Wants=network-online.target

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
Environment=DAEMON_RESTART_AFTER_UPGRADE=true
Environment=DAEMON_ALLOW_DOWNLOAD_BINARIES=false

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now Neutaro
sudo journalctl -fu Neutaro -o cat
```

### 8.3 Expected logs — the happy path

In order. If you never reach stage ③, go to §11.

```
① INF Starting state sync module=statesync
   INF Downloading trusted light block using options module=light
   INF sync any module=statesync msg="Discovering snapshots for 15s"

② INF Discovered new snapshot format=3 height=16619000 module=statesync

③ INF Offering snapshot to ABCI app format=3 height=16619000 module=statesync
   INF Snapshot accepted, restoring format=3 height=16619000 module=statesync

④ INF Fetching snapshot chunk chunk=0 format=3 height=16619000 total=3
   INF Applied snapshot chunk to ABCI app chunk=0 format=3 height=16619000 total=3

⑤ INF executed block height=16619419 module=state num_valid_txs=0
   INF indexed block events height=16619419 module=txindex
```

Stage ⑤ = state sync is done and you are replaying live blocks. On this install ①→⑤ took **~90 seconds**.

**Benign noise you will see and should ignore:**

```
INF Connection is closed @ recvRoutine (likely by the other side)
ERR Stopping peer for error err=EOF module=p2p
INF Reconnecting to peer module=p2p
INF Inbound Peer rejected err="filtered CONN<...>: duplicate CONN<...>" numPeers=38
```

Peer churn is normal. It is only a problem if stage ② repeats with **no** stage ③.

Also benign — a snapshot can be rejected and the next one still succeed:

```
ERR error from light block request from primary, removing... error="post failed: context deadline exceeded"
ERR failed to remove witnesses err="no witnesses connected. please reset light client"
INF Snapshot rejected format=3 height=15046000 module=statesync
INF Offering snapshot to ABCI app format=3 height=16619000     <-- recovered on the next candidate
```

### 8.4 Verify

```bash
Neutaro status | jq .SyncInfo
```

```json
{
  "latest_block_height": "16619421",
  "earliest_block_height": "16619001",
  "catching_up": false
}
```

Wait for `"catching_up": false`. `earliest_block_height` being ~the snapshot height (not `1`) is
correct — a state-synced node has no history before its restore point.

> Nit: the current guide writes `Neutaro status 2>&1 | jq .SyncInfo`. On v0.47.15 `status` writes to
> **stdout** (verified: 1065 bytes stdout, 0 bytes stderr), so the `2>&1` is unnecessary — and if
> anything ever *does* hit stderr it corrupts jq's input. Drop it. `.SyncInfo` (capitalised) is
> correct; top-level keys are `NodeInfo, SyncInfo, ValidatorInfo`.

### 8.5 Turn state sync back off

Otherwise a later restart with a stale `trust_height` will try to re-sync.

```bash
"$HOME/tomlset.py" "$HOME/.Neutaro/config/config.toml" \
  'statesync:enable=false' 'statesync:rpc_servers=""' \
  'statesync:trust_height=0' 'statesync:trust_hash=""'
sudo systemctl restart Neutaro
```

---

## 9. 🔴 Serving state sync to others

**Snapshot production is off by default.** `snapshot-interval` defaults to `0`,
and the guide does not touch it — so every node built from that guide can state-sync *from* the network
and can never serve it back. That is why the chain has so few usable state-sync sources.

If you follow only §2–§8 you are a **consumer**. To become a **provider**:

```bash
"$HOME/tomlset.py" "$HOME/.Neutaro/config/app.toml" \
  'state-sync:snapshot-interval=1000' \
  'state-sync:snapshot-keep-recent=2'

"$HOME/tomlset.py" "$HOME/.Neutaro/config/config.toml" \
  'rpc:laddr="tcp://0.0.0.0:26657"' \
  'rpc:cors_allowed_origins=["*"]' \
  "p2p:external_address=\"$(curl -s ifconfig.me):26656\""

sudo ufw allow 26656/tcp comment 'Neutaro p2p (snapshot serving)'
sudo ufw allow 26657/tcp comment 'Neutaro RPC (state-sync clients)'
sudo systemctl restart Neutaro
```

### 9.1 ⚠️ The pruning/snapshot constraint

```
pruning-keep-recent  ≥  snapshot-interval × snapshot-keep-recent
```

This is the Cosmos SDK's documented requirement for snapshot providers: the state versions a snapshot
is built from must survive long enough to be read. The current guide's `pruning-keep-recent = "100"`
does not satisfy it for any snapshot interval ≥ 100, so a provider configured that way risks snapshots
that silently never appear. `1000 × 2 = 2000`, which is why §7.2 sets `pruning-keep-recent = "2000"`.

> Scope note: this install was built to satisfy the constraint, not to reproduce its failure — we did
> not run a node at `pruning-keep-recent = 100` to watch snapshots fail. The rule is upstream SDK
> guidance; §9's "confirm you are actually serving" check is how you verify your own box either way.

Also keep `min-retain-blocks` (500000) far above `snapshot-interval × snapshot-keep-recent`, or you
prune the blocks your consumers' light clients need to verify against.

### 9.2 Confirm you are actually serving

Snapshots are only written at heights divisible by `snapshot-interval`, so allow up to
`interval × 5.83 s` (≈97 min at 1000) before judging.

```bash
ls -la ~/.Neutaro/data/snapshots/          # one directory per snapshot height
journalctl -u Neutaro | grep -i "Creating state snapshot\|Completed state snapshot"
```

```
INF Creating state snapshot  height=16620000 module=server
INF Completed state snapshot height=16620000 format=3 module=server
```

> There is **no `Neutaro snapshots list` subcommand** in this build — the SDK's `snapshots` command
> group is not registered (`Error: unknown command "snapshots" for "Neutaro"`). Inspect the directory.

### 9.3 Publish your endpoint

Give consumers your node ID and P2P address:

```bash
echo "$(Neutaro tendermint show-node-id)@$(curl -s ifconfig.me):26656"
```

If you expose RPC publicly, put TLS in front of it (nginx + certbot) rather than shipping plaintext
`:26657` — consumers put your URL in `rpc_servers`, and a MITM there feeds them a forged trust hash.

---

## 10. 🔴 Validator safety, and state-syncing a validator that is already running

### 10.1 Know what the two penalties actually cost

Query them yourself — do not take anyone's word, including this guide's:

```bash
curl -s https://api2.neutaro.io/cosmos/slashing/v1beta1/params | jq .
```

```json
{ "signed_blocks_window": "51840", "min_signed_per_window": "0.050000000000000000",
  "downtime_jail_duration": "600s",
  "slash_fraction_downtime": "0.010000000000000000",
  "slash_fraction_double_sign": "0.050000000000000000" }
```

Read that carefully, because it drives every decision below:

| | Downtime | Double-sign |
|---|---|---|
| Trigger | sign < 5 % of a 51,840-block window | sign two different blocks at one height |
| Cost | 1 % slash, 600 s jail, then `unjail` | **5 % slash + PERMANENT TOMBSTONE** |
| Recoverable | ✅ yes | ❌ **never** — not with that key, not with a new one |

5 % of 51,840 means you must sign 2,592 blocks per window — so you can miss **~49,248 consecutive
blocks, roughly 80 hours**, before you are even jailed.

**Therefore: never trade double-sign risk for speed.** You have three days of downtime budget. Use
it. Stop the node, think, check, then act.

### 10.2 `unsafe-reset-all` vs `reset-state` — the distinction that saves the validator

Both claim to "remove all the data and WAL". Only one is safe on a validator:

```
reset-state        Remove all the data and WAL
unsafe-reset-all   Remove all the data and WAL, reset this node's validator to genesis state
```

That trailing clause zeroes `priv_validator_state.json`:

```
Reset private validator file to genesis state keyFile=.../priv_validator_key.json stateFile=.../priv_validator_state.json
```

…which deletes the only thing that physically prevents your key from signing a height twice. And
state sync makes that dangerous rather than theoretical: the node restores at a snapshot height
*below* what it last signed (providers snapshot every 1000–2000 blocks and keep two), then
block-syncs **up through heights it has already signed**.

**On a validator, always use `reset-state`.** The previous revision of this file put `unsafe-reset-all` in
the state-sync path and then, four steps later, tells you to create a validator on the same node,
with no warning between them.

### 10.3 Procedure — state-syncing a validator that is already live

Use this when a validator's disk is full. It is far safer than rebuilding the box, because the
consensus key never moves. **Executed on a live validator with 596,270 voting power on 2026-08-14;
every log line below is captured output from that run.**

Result: **183 G → 9.8 G, ~9 minutes of downtime, 87 missed blocks of ~49,248 allowed, voting power
unchanged, `tombstoned: false`.**

**Do all preparation while the node is still running and signing** — config is only read at startup,
so pruning changes, peer changes and script staging cost zero downtime. Only steps 3–8 need the node
down.

```bash
BIN=~/.Neutaro/cosmovisor/upgrades/v2/bin/Neutaro    # the binary cosmovisor actually runs
```

**1. Back up the keys** and verify them — see the backup section above. Do not skip the verify.

**2. Pre-stage config** (node still running). Set `min-retain-blocks`, or the disk simply refills:

```toml
pruning-keep-recent = "2000"
pruning-interval    = "10"
min-retain-blocks   = 444902     # ~30 days. Default 0 = keep every block forever.
snapshot-interval   = 0          # a validator should not spend I/O serving snapshots
```

**3. Stop, and prove it stopped.**

```bash
sudo systemctl stop Neutaro && sleep 6
systemctl is-active Neutaro     # inactive
pgrep -a Neutaro; pgrep -a cosmovisor    # both must print nothing
```

From here until step 8, this key must not be running anywhere on earth. Not for one block.

**4. Take the authoritative state file** — *this* copy, not the one in your backup archive:

```bash
cp -a ~/.Neutaro/data/priv_validator_state.json ~/priv_validator_state.POST-STOP.json
cat ~/priv_validator_state.POST-STOP.json     # note the height — call it H_last
```

**5. `reset-state`** — and confirm the guard survived:

```bash
$BIN tendermint reset-state --home ~/.Neutaro
```

```
I[18:32:39.057] Removed all blockstore.db     dir=/root/.Neutaro/data/blockstore.db
I[18:32:41.179] Removed all state.db          dir=/root/.Neutaro/data/state.db
I[18:32:41.348] Removed all cs.wal            dir=/root/.Neutaro/data/cs.wal
I[18:32:41.349] Removed all evidence.db       dir=/root/.Neutaro/data/evidence.db
I[18:32:42.456] Removed tx_index.db           dir=/root/.Neutaro/data/tx_index.db
```

```bash
cat ~/.Neutaro/data/priv_validator_state.json    # MUST still show H_last, not 0
```

If it reads `0`, you ran `unsafe-reset-all`. Restore the step-4 copy before doing anything else.

**6. 🔴 Remove `application.db` as well — every guide misses this.**

Look at what step 5 left behind:

```
drwxr-xr-x 2 root root 516096 Aug 14 18:30 application.db      <- 13 G, still there
-rw------- 1 root root    401 Aug 14 18:31 priv_validator_state.json
```

`reset-state` and `unsafe-reset-all` are **CometBFT** commands. `application.db` is the **Cosmos
SDK** application database and neither touches it. Leave it and the node will not start: the app
reports its old last-block height while CometBFT sits at genesis, and the handshake fails with
*"app block height is higher than core"*. State sync never runs.

```bash
cd ~/.Neutaro/data
cp -a priv_validator_state.json ~/priv_validator_state.GUARD.json   # guard first
rm -rf application.db snapshots
ls -la          # only priv_validator_state.json + upgrade-info.json should remain
```

**7. Prime state sync and start** — §8.1, with a generous lag:

```bash
TRUST_LAG=10000 ~/state_sync.sh
sudo systemctl start Neutaro
```

```
INF Offering snapshot to ABCI app    format=3 height=16624000 module=statesync
INF Snapshot accepted, restoring     format=3 height=16624000 module=statesync
INF Fetching snapshot chunk  chunk=1 format=3 height=16624000 total=3
INF Fetching snapshot chunk  chunk=2 format=3 height=16624000 total=3
INF Fetching snapshot chunk  chunk=0 format=3 height=16624000 total=3
INF Applied snapshot chunk to ABCI app  chunk=0 ...
INF Applied snapshot chunk to ABCI app  chunk=1 ...
```

`/status` reports `height: 0, catching_up: true` for several minutes during the restore, and the
final chunk consistently lands well after the others. **Neither is a hang.** Wait.

**8. Confirm it is signing again — and that the chain agrees.**

```bash
curl -s "https://api2.neutaro.io/cosmos/slashing/v1beta1/signing_infos/YOUR_VALCONS" | jq .
cat ~/.Neutaro/data/priv_validator_state.json     # height must now be ABOVE H_last
```

```
tombstoned    : False        jailed_until : 1970-01-01T00:00:00Z
missed_blocks : 89           (87 missed during the operation)
status        : BOND_STATUS_BONDED
height        : 16624575     catching_up : false
earliest      : 16624001     <- state-synced, correct
```

`H_last` was 16624479 and the node resumed at 16624569 — above it, never below. That is the guard
doing its job. Finally set `[statesync] enable = false` (§8.5).

### 10.4 What this costs you

`tx_index.db` is wiped and rebuilds only from the sync height forward, so **the node loses its own
transaction history**. Any local LCD/RPC query for an older transaction returns nothing. Verify
historical transactions against a public endpoint such as `api2.neutaro.io` instead. Files on disk
(exports, receipts, CSVs) are unaffected.

### 10.5 The rules, short

* Never run two nodes with the same `priv_validator_key.json`. This is the only unrecoverable mistake.
* On a validator: `reset-state`, never `unsafe-reset-all`.
* `application.db` must be removed separately, or state sync will not run at all.
* The only safe `priv_validator_state.json` is one copied **after** the node stopped.
* Prefer downtime over risk, always. You have ~80 hours of it.

---

## 11. Troubleshooting

### 11.1 Health check

```bash
systemctl status Neutaro --no-pager
Neutaro status | jq .SyncInfo
curl -s localhost:26657/net_info | jq -r '.result.n_peers'          # want > 5
curl -s localhost:26657/status | jq -r '.result.sync_info.catching_up'
journalctl -u Neutaro --since "-5min" -o cat | grep -iE 'ERR|panic' | head
df -h / && du -sh ~/.Neutaro/data/*
```

### 11.2 `Discovered new snapshot` repeats forever, never `Offering snapshot to ABCI app`

The #1 failure. Four causes, in order of likelihood:

1. **Trust height is above every available snapshot.** Compare:
   ```bash
   grep -E '^trust_height' ~/.Neutaro/config/config.toml
   journalctl -u Neutaro -o cat | grep -o 'Discovered new snapshot.*height=[0-9]*' | grep -o 'height=[0-9]*' | sort -u
   ```
   Every discovered height below `trust_height` → re-run with a bigger lag:
   ```bash
   sudo systemctl stop Neutaro
   TRUST_LAG=20000 "$HOME/state_sync.sh"
   Neutaro tendermint unsafe-reset-all --home "$HOME/.Neutaro" --keep-addr-book
   rm -rf "$HOME/.Neutaro/data/application.db"     # or state sync will not run — §10.3 step 6
   sudo systemctl restart Neutaro
   ```
   ⚠️ **On a validator, swap `unsafe-reset-all` for `reset-state`** — see §10.2. `unsafe-reset-all`
   zeroes your double-sign guard.
2. **Too few snapshot-serving peers.** Check `n_peers`. If you are stuck on 1–2, the seed has not
   populated your address book yet; wait, or add more `persistent_peers` (§11.3).
3. **The offering peer keeps dropping you** (`Stopping peer for error err=EOF` on a loop against a
   single peer) — its inbound slots are full. Add peers.

4. **The node you are re-syncing is itself one of the `rpc_servers`.** Found the hard way while
   re-syncing `rpc3.neutaro.io`: the light client uses the second entry as its *witness*, and that
   witness was the node being wiped — sitting at height 0, unable to confirm any header, so
   verification never completes even though snapshots *above* the trust height are being discovered.
   The symptom is identical to cause 1, but the trust height is fine. Fix: point both entries at the
   *other* RPC — duplicates are allowed:
   ```toml
   rpc_servers = "https://rpc2.neutaro.io:443,https://rpc2.neutaro.io:443"
   ```
   This applies any time you rebuild one of the chain's public RPC endpoints.

### 11.3 Verify a peer before you trust the config

The value shipped in the current guide as the state-sync `persistent_peers` —
`ee64e5d0c3549fe807149f5f29a2913074e08a62@147.93.4.184:26656` — **is dead**
(TCP connect times out, verified twice on 2026-08-14). Following the guide literally leaves you with
exactly one, unreachable, peer.

Always test before committing a peer list:

```bash
nc -vz 147.93.4.184 26656          # -> times out (dead)
nc -vz 109.199.106.233 26656       # -> succeeded
```

Harvest live peers from a public RPC instead of trusting a static list:

```bash
curl -s https://rpc2.neutaro.io/net_info \
  | jq -r '.result.peers[] | "\(.node_info.id)@\(.remote_ip):26656  # \(.node_info.moniker)"'
```

Verified live 2026-08-14:

| Peer | Moniker |
|---|---|
| `0e24a596dc34e7063ec2938baf05d09b374709e6@109.199.106.233:26656` | NeutaroRPC (rpc2) |
| `726d5975dd11383a175d1b748526257d3749058c@185.182.184.50:26656` | Rock (rpc3) |
| `95f6fc822469efdf868ab6cdfc218fa80f716951@62.84.180.12:26656` | archivenode |
| `90dcafcc67687feff6d1b355892a3690c6cb71f3@185.182.184.8:26656` | Jinnx |
| `d891af90afdcf3973f7dc44eef316dfe652ccb1c@38.242.135.246:26656` | TimpiTap |

Seed (port **36656**, not 26656): `84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:36656`

Note `rpc.neutaro.io` does **not** resolve; use `rpc2` / `rpc3`.

### 11.4 `no witnesses connected. please reset light client`

```
ERR error from light block request from primary, removing... error="post failed: context deadline exceeded"
ERR failed to remove witnesses err="no witnesses connected. please reset light client" witnessesToRemove=[0]
```

Both `rpc_servers` timed out. Usually transient — CometBFT retries with the next snapshot and
recovers on its own (it did here). If it persists, the RPCs are down or you are rate-limited; re-run
`state_sync.sh` to pick a fresh pair. You need **two distinct** entries in `rpc_servers`; one is
rejected.

### 11.5 Service will not start

| Symptom | Cause |
|---|---|
| `DAEMON_DATA_BACKUP_DIR must not be empty` | `~/.Neutaro/data-backup` deleted, or unit missing the env var. Recreate it. |
| `cannot open shared object file: libwasmvm.x86_64.so` | The binary links libwasmvm out of the **Go module cache**. Never `rm -rf ~/go/pkg` on a Neutaro host. Restore from `github.com/CosmWasm/wasmvm/releases/download/v1.5.0/libwasmvm.x86_64.so` into `/usr/lib` + `ldconfig`. Silent until the next restart. |
| Starts, then exits at genesis height | Wrong/corrupt `genesis.json` — re-check the sha256 in §7. |
| `error in app.toml: set minimum gas price` | `minimum-gas-prices` empty. §7.2. |
| **`app block height is higher than core`** | You reset the data but left **`application.db`**. Neither reset command removes it — it is the Cosmos SDK's database, not CometBFT's. `rm -rf ~/.Neutaro/data/application.db` and restart. **§10.3 step 6.** |
| State sync configured, but the node block-syncs from genesis instead | Same cause as above, or `data/` was not actually emptied. State sync only triggers on a node with no existing state. |

### 11.6 Disk filling up

```bash
du -sh ~/.Neutaro/data/*
```

| Directory | Pruned by | If it is huge |
|---|---|---|
| `blockstore.db` | `min-retain-blocks` | It is `0`. Set it (§7.2) and restart. |
| `state.db` | `min-retain-blocks` | Same. |
| `application.db` | `pruning-keep-recent` | Lower it — but respect the §9 constraint if you serve snapshots. |
| `tx_index.db` | **nothing** | Only fix is `indexer = "null"` + delete the directory. You lose local tx history. |
| `snapshots/` | `snapshot-keep-recent` | Working as intended; lower `snapshot-keep-recent`. |

### 11.7 Nobody can state-sync from me

1. `snapshot-interval` still `0`? (§9 — the default, and the current guide never changes it.)
2. `pruning-keep-recent < snapshot-interval × snapshot-keep-recent`? (§9.)
3. `26656` closed, or `external_address` unset/wrong?
4. Not yet reached a height divisible by `snapshot-interval`.

---

## 12. Reference

```bash
# identity
Neutaro tendermint show-node-id
Neutaro status | jq .NodeInfo.moniker

# peers
curl -s localhost:26657/net_info | jq -r '.result.n_peers'

# config
grep -A6 '^\[statesync\]' ~/.Neutaro/config/config.toml
grep -E '^(pruning|min-retain-blocks|minimum-gas-prices)' ~/.Neutaro/config/app.toml
grep -A4 '^\[state-sync\]' ~/.Neutaro/config/app.toml

# service
sudo systemctl restart Neutaro && sudo journalctl -fu Neutaro -o cat

# backup (see the backup section — verify it, do not just copy it)
tar czf keys-$(date +%F).tar.gz -C ~/.Neutaro \
    config/priv_validator_key.json config/node_key.json keyring-test/

# validator identity, three ways — they must all agree
Neutaro tendermint show-validator                     # consensus pubkey
Neutaro tendermint show-address                       # consensus address
curl -s localhost:26657/status | jq -r '.result.validator_info.address'

# wiping data: pick the right one
Neutaro tendermint reset-state                        # validators — keeps the double-sign guard
Neutaro tendermint unsafe-reset-all --keep-addr-book  # non-validators ONLY
rm -rf ~/.Neutaro/data/application.db                 # required by BOTH of the above (§10.3 step 6)

# am I in trouble? (public endpoint, not your own node)
curl -s https://api2.neutaro.io/cosmos/slashing/v1beta1/params | jq .
curl -s https://api2.neutaro.io/cosmos/slashing/v1beta1/signing_infos/YOUR_VALCONS | jq .
```

| Port | Purpose | Expose |
|---|---|---|
| 26656 | P2P — **snapshot transfer** | Yes, always |
| 26657 | RPC — light-client verification | Only if you serve state sync (prefer TLS) |
| 1317 | REST/LCD | Localhost unless deliberately published |
| 9090 | gRPC | Localhost unless deliberately published |
| 6060 | pprof | Localhost only |

---

## Appendix — what changed from the previous version of this file, and why

The previous revision of this file (commit `2b302b72`, 2026-01-06) was checked line by line;
every row below was verified against it rather than recalled. Line numbers refer to that revision.

### 🔴 Breaks the thing the file is named after, or risks a validator

| # | Issue | Where |
|---|---|---|
| 1 | **`snapshot-interval` is never set.** It defaults to `0`, and the previous revision never set it — so no node built from it could ever *serve* state sync. That is why the chain has so few sources. | absent |
| 2 | **`pruning-keep-recent = "100"`** violates the SDK's `≥ snapshot-interval × snapshot-keep-recent` rule, so a node that *does* enable snapshots still serves nothing. | L229 |
| 3 | **The only `persistent_peers` entry is dead.** `147.93.4.184:26656` times out (verified twice). Follow the guide literally and you get exactly one unreachable peer. | L241 |
| 4 | **`TRUST_HEIGHT=$((HEIGHT-2000))` deadlocks.** Providers snapshot every 1000–2000 blocks and keep two, so the newest snapshot is routinely ~2000 blocks behind tip — putting trust height *above every snapshot that exists*. Observed verbatim: offers at 16616000/16614000, trust height 16617286, `Offering snapshot` never fires. | L254 |
| 5 | **`unsafe-reset-all` sits in the state-sync path with no double-sign warning** — and §11 tells you to create a validator on that same node. It resets `priv_validator_state.json` to genesis. | L267 vs L343 |
| 6 | **`reset-state` is never mentioned**, though it is the safe alternative on a validator: same wipe, keeps the double-sign guard. | absent |
| 7 | **`application.db` survives both reset commands** — it is the Cosmos SDK's database, not CometBFT's. Leave it and the node dies on *"app block height is higher than core"* and state sync silently never runs. | absent |
| 8 | **No backup section at all.** Nothing says which files are irreplaceable, where they live, or how to *verify* a backup rather than merely copy one. | absent |

### 🟠 Costs you a disk, a restart, or a security property

| # | Issue | Where |
|---|---|---|
| 9 | `min-retain-blocks` never mentioned → `blockstore.db` and `state.db` grow forever. This is what took a production validator to 183 G on a 194 G disk. | absent |
| 10 | `tx_index.db` growth never mentioned; it is pruned by nothing (~42 MB/day observed). | absent |
| 11 | `state_sync.sh` writes an unvalidated `trust_hash` — only the RPC probe on L252 uses `-f`; L253 and L255 are bare `curl -s`, so a failed request silently writes an empty hash. | L252–L255 |
| 12 | `minimum-gas-prices` never mentioned — the binary's `init` default `"0stake"` names a denom Neutaro does not have, and the node refuses to start until it is changed. | absent |
| 13 | Genesis and snapshot are fetched from **bare IPs over plaintext HTTP** (`154.26.153.186`, `173.212.198.246`) with no TLS, no hostname and no checksum. | L215, L276 |
| 14 | `rm -r go` — relative to the current directory, unquoted, no `-f`. | L110 |
| 15 | No procedure for state-syncing an **existing** node; only fresh installs are covered, yet a full disk is the usual reason anyone needs state sync. | absent |
| 16 | Slashing parameters never quoted, so a reader cannot see that downtime is ~80 h of budget while double-signing is permanent. | absent |

### 🟡 Friction and polish

| # | Issue | Where |
|---|---|---|
| 17 | `set -euo pipefail` recommended for an *interactive* shell — one non-zero exit closes your session mid-install. | L118 |
| 18 | `external_address`, the RPC bind address, and `26657` in UFW are never covered. | absent |
| 19 | `trust_period` never set. | absent |
| 20 | No troubleshooting section and no expected-log reference — the file goes from "run this" straight to "create a validator". | absent |
| 21 | `Neutaro status 2>&1 \| jq .SyncInfo` — the `2>&1` pipes stderr into `jq`, which turns a clear error into a parse failure. | L317 |

Items 6, 7, 8, 15 and 16 come from state-syncing a live 596,270-power validator on 2026-08-14;
§10.3 is that run, logs and all.
