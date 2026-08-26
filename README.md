# 🌐 Neutaro – Validator Setup & Governance Tools

Welcome to the official repository for **Neutaro**, the blockchain that powers governance, incentives, and decentralization for the **Timpi** search engine.

This repository is the primary entry point for anyone who wants to:

* Run a **Neutaro node or validator**
* Stake or delegate **NTMPI**
* Participate in **on-chain governance**
* Contribute infrastructure to the **Timpi ecosystem**

---

## 📋 Quick Facts

| | |
|---|---|
| Chain ID | `Neutaro-1` |
| Binary | `Neutaro` (Cosmos SDK v0.47, CometBFT v0.37) |
| Token | `NTMPI` — `1 NTMPI = 1,000,000 uneutaro` |
| Seed | `84ae242b0c4c14af59a61438ba2eca4573b91c95@109.199.106.233:36656` |
| Public RPC | `https://rpc2.neutaro.io` |
| Public API (LCD) | `https://api2.neutaro.io` |
| Genesis sha256 | `78724fe90e5bd1f2abd0186bd0c33325e3c190d8e417741a42ff0c7fbdf2fc2d` |
| Explorer | `https://explorer.neutaro.io` |
| P2P port | `26656` (open this one) |

## 📚 Documentation Index

| Guide | When you need it |
|---|---|
| [Installation](Instructions/NeutaroInstallation.md) | New node or validator, from clean Ubuntu to running service |
| [Docker Node](Instructions/NeutaroDocker.md) | A full node in two commands — state-synced in minutes, cannot sign for any validator |
| [State Sync & Snapshot Provider](statesync.md) | Fast sync (~10 min), **key backup + verification**, serving snapshots, validator-safe maintenance, troubleshooting |
| [Validator Commands](Instructions/NeutaroValidatorCommands.md) | Day-to-day: edit, delegate, vote, unjail, transfer |
| [Security Guide](SecurityGuide.md) | Hardening the server and the keys |
| [Mandatory Update (Feb 2026)](MandatoryValidatorUpdate.md) | Historical — already included when building from `main` |

## 📦 What This Repository Provides

* ✅ Step-by-step validator and node installation guides
* ⚡ Fast sync options (State Sync & snapshots)
* 🔐 Validator security best practices
* 🗳️ Governance, staking, and proposal voting tools
* 📁 Source build instructions with `cosmovisor` and `systemd`

---

## 💡 Why Neutaro?

Neutaro is the **governance and reward chain** backing **Timpi** — the first decentralized search engine.

By participating in Neutaro, you can:

* 🛠 Run a validator or node to support the network
* 💸 Stake NTMPI and earn rewards
* 🗳 Vote on governance proposals, including **ethical and operational decisions** affecting Timpi
* 🌍 Help secure and decentralize critical Timpi infrastructure

---

## 🛡️ Security First (Read Before Running a Node)

Running blockchain infrastructure comes with real responsibility.

Before proceeding, **read the official Neutaro Security Guide** to understand how to properly secure your system and keys:

👉 **[Validator Security Guide →](SecurityGuide.md)**

Topics covered include:

* Linux system hardening
* Firewall and port management
* Backup and recovery strategies
* Validator and key management best practices

---

## 🔧 Validator Overview

Validators are responsible for:

* Producing and validating blocks
* Securing the Neutaro network
* Participating in governance
* Earning staking rewards for themselves and delegators

A validator runs a full Neutaro node and maintains high uptime, security, and operational discipline.

---

## 🐧 Validator & Node Setup (Linux – Ubuntu 22.04.03+)

### Standard Validator Installation

👉 **[Neutaro Validator Setup Guide →](Instructions/NeutaroInstallation.md)**

This guide covers:

* Building Neutaro from source
* `cosmovisor` layout and upgrades
* Pruning configuration
* Snapshot usage
* Firewall configuration
* `systemd` service setup

### Validator & Wallet Commands

🛠️ **[Validator Command Reference →](Instructions/NeutaroValidatorCommands.md)**

Includes commands for:

* Creating and editing validators
* Delegating, redelegating, and unbonding
* Voting on proposals
* Unjailing and validator maintenance
* Token transfers

---

## ⚡ Fast Node Sync (Recommended)

Syncing a Neutaro node does **not** require downloading the full blockchain history.

Neutaro supports **State Sync**, allowing new nodes to securely sync to the latest chain state in **minutes instead of days**.
This is the **recommended method** for new validators and node operators.

### 📘 Fast Sync & State Sync Guide

👉 **[Neutaro State Sync & Fast Node Guide →](statesync.md)**

This guide provides a **complete, tested setup** — every command was executed against
`Neutaro-1` and every log line quoted is real captured output — including:

* 🔐 **Key backup with a verification step** (start here even if you sync another way)
* Go + Neutaro source build, `cosmovisor` + `systemd`
* **State Sync (recommended path)** with expected logs for every stage
* **Serving snapshots back to the network** — how to become a state-sync source
* **Validator-safe maintenance** — reclaiming a full disk without risking the validator
* Troubleshooting keyed by the exact log line you are seeing

> After syncing, your node behaves exactly like a normal full node.
> You may safely create or start a validator once `catching_up = false`.
> Snapshot download as an alternative sync path is covered in the
> [Installation guide](Instructions/NeutaroInstallation.md), step 7.

---

## 💡 Delegate Instead of Validating?

Not ready to operate a full validator?

You can **delegate your NTMPI** to an existing validator and still earn staking rewards, without running any infrastructure.

Example:

```bash
Neutaro tx staking delegate <validator_address> 100000000uneutaro \
  --from YOURWALLET \
  --chain-id Neutaro-1
```
