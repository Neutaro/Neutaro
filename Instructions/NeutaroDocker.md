# Neutaro in Docker — a full node in two commands

Run a Neutaro node in a container: state-synced in minutes, self-configuring, and **incapable of
signing for any validator** unless you deliberately give it a key. Ideal for RPC/API access,
development, monitoring, or just following the chain — anywhere Docker runs.

Everything below was executed end-to-end against `Neutaro-1`; the quickstart's expected log lines
are captured output.

## Contents

- [1. Why a container cannot double-sign](#1-why-a-container-cannot-double-sign)
- [2. Quickstart](#2-quickstart)
- [3. docker-compose](#3-docker-compose)
- [4. Configuration (environment variables)](#4-configuration-environment-variables)
- [5. ⚠️ Running two nodes behind one IP](#5-️-running-two-nodes-behind-one-ip)
- [6. Logs, upgrades, disk](#6-logs-upgrades-disk)
- [7. 🔴 Validators and Docker](#7--validators-and-docker)

---

## 1. Why a container cannot double-sign

On first boot the entrypoint runs `Neutaro init`, which generates a **fresh consensus key** inside
the container's volume. That key has never existed before and is not in the validator set — its
voting power is zero, so nothing it could ever sign counts for consensus. The chain does not know
it exists.

This is the safe default and the entire point: **a node, not a validator.** You get the full chain
— RPC, queries, tx broadcasting, event subscriptions — with no signing risk whatsoever. The only
way a container can endanger a validator is if you copy a real `priv_validator_key.json` into it
(§7).

## 2. Quickstart

```bash
git clone https://github.com/Neutaro/Neutaro && cd Neutaro
docker build -t neutaro:local .          # static binary (the repo's existing Dockerfile)
docker build -t neutaro:node docker/     # node runtime + entrypoint

docker run -d --name neutaro-node \
  -v neutaro-data:/neutaro \
  -p 26656:26656 -p 127.0.0.1:26657:26657 \
  --log-opt max-size=50m --log-opt max-file=3 \
  neutaro:node

docker logs -f neutaro-node
```

First boot, expected sequence:

```
[entrypoint] first boot: init docker-node on Neutaro-1
[entrypoint] fetching genesis
[entrypoint] genesis verified: 78724fe90e5bd1f2abd0186bd0c33325e3c190d8e417741a42ff0c7fbdf2fc2d
[entrypoint] priming state sync (TRUST_LAG=10000)
[entrypoint] primed: tip=<height> trust_height=<height-10000>
[entrypoint] starting Neutaro
...
INF Discovered new snapshot ... module=statesync
INF Snapshot accepted, restoring ...
INF Applied snapshot chunk to ABCI app chunk=0 ...
INF Snapshot restored ...
```

The entrypoint **verifies the genesis checksum and refuses to start on a mismatch**, and fails
loudly rather than writing an empty trust hash. Check sync state:

```bash
curl -s localhost:26657/status | jq .result.sync_info
# "catching_up": false  →  done. Measured: 2–10 minutes depending on chunk peers.
```

Restarts resume from local data — state sync only primes when the data directory is empty
(`STATESYNC=auto`).

## 3. docker-compose

`docker/docker-compose.yml` does the same with restart policy and log rotation included:

```bash
cd docker && docker compose up -d && docker compose logs -f
```

## 4. Configuration (environment variables)

| Variable | Default | Meaning |
|---|---|---|
| `MONIKER` | `docker-node` | node name |
| `SEEDS` | the official seed (`…@109.199.106.233:36656`) | comma-separated |
| `PERSISTENT_PEERS` | a verified-live set | comma-separated `id@host:port` |
| `STATESYNC` | `auto` | `auto` = prime only when data is empty · `on` · `off` |
| `TRUST_LAG` | `10000` | blocks behind tip for the trust height (see `statesync.md` §11.2) |
| `GENESIS_URL` / `GENESIS_SHA256` | official | change both together or not at all |
| `RPC1` / `RPC2` | both rpc2.neutaro.io | light-client verification endpoints (rpc3 retired 2026-08-26) |

Example — custom moniker and your own peers:

```bash
docker run -d --name neutaro-node \
  -e MONIKER=my-node \
  -e PERSISTENT_PEERS="id1@host1:26656,id2@host2:26656" \
  -v neutaro-data:/neutaro -p 26656:26656 -p 127.0.0.1:26657:26657 \
  neutaro:node
```

## 5. ⚠️ Running two nodes behind one IP

If **another Neutaro node already runs behind your public IP** — the host itself, another machine
on your LAN, a validator in the house — the container will be rejected by every peer that node is
connected to: peers filter duplicate IPs, and the symptom is
`auth failure: secret conn failed: connection reset by peer` from everyone, forever
(`statesync.md` §11.2a, found running exactly this setup).

Fix: point `PERSISTENT_PEERS` at nodes the other machine is **not** connected to. List its peers
with `curl -s localhost:26657/net_info | jq -r '.result.peers[].remote_ip'` on that machine.

## 6. Logs, upgrades, disk

**Logs:** Docker's default json-file driver grows **without bound**. The quickstart's
`--log-opt max-size=50m --log-opt max-file=3` (or the compose file's logging block) caps it at
150 MB. Do not skip this on a long-lived node.

**Upgrades:** rebuild and recreate — the volume keeps all state:

```bash
cd Neutaro && git pull
docker build -t neutaro:local . && docker build -t neutaro:node docker/
docker stop neutaro-node && docker rm neutaro-node
# ...then the same docker run as in the quickstart
```

There is no cosmovisor inside the container — Docker's restart policy plays that role, and chain
upgrades arrive as image rebuilds instead of on-disk binary swaps.

**Disk:** a pruned state-synced node stabilises around ~10 GB
(`docker system df -v` shows the volume). `tx_index.db` grows ~40 MB/day and is pruned by nothing;
set `indexer = "null"` in the volume's `config.toml` if you never query historical txs locally.

## 7. 🔴 Validators and Docker

Short version: **don't, unless you have a specific reason.** A validator's one unforgivable
failure is the same key signing from two places (`statesync.md` §10.1) — and containers make
duplication *easy*: images are copied, volumes are cloned, `docker run` is cheap, orchestrators
restart things in surprising places. The safety margin that systemd + one machine gives you
exists precisely because starting a second copy is *hard*.

If you accept that and proceed anyway:

1. The key goes in via a **read-only bind mount** — never baked into an image, never in a volume
   that might be cloned:
   ```bash
   -v /secure/priv_validator_key.json:/neutaro/.Neutaro/config/priv_validator_key.json:ro
   ```
2. `priv_validator_state.json` (the double-sign guard) must live on a **persistent volume** that
   exactly one container can ever use. Losing it or cloning it = tombstone risk.
3. `restart: unless-stopped`, **never** an orchestrator that reschedules onto other hosts
   (no Swarm/Kubernetes replicas, ever — `replicas: 2` on a validator is a self-tombstone button).
4. Read `statesync.md` §1 (backup + verify) and §10 before the first start.

The sane middle ground: run the **validator** on bare metal with systemd, and use Docker nodes
(this guide) for everything else — RPC, monitoring, indexing, development.
