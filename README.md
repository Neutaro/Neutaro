# 🌐 Neutaro – Validator Setup & Governance Tools

Welcome to the official repository for **Neutaro**, the blockchain that powers governance, incentives, and decentralization for the **Timpi** search engine.

This repository is the primary entry point for anyone who wants to:

* Run a **Neutaro node or validator**
* Stake or delegate **NTMPI**
* Participate in **on-chain governance**
* Contribute infrastructure to the **Timpi ecosystem**

---

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

👉 **[Validator Security Guide →](https://github.com/Neutaro/Neutaro/blob/main/SecurityGuide.md)**

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

👉 **[Neutaro Validator Setup Guide →](https://github.com/Neutaro/Neutaro/blob/main/Instructions/NeutaroInstallation.md)**

This guide covers:

* Building Neutaro from source
* `cosmovisor` layout and upgrades
* Pruning configuration
* Snapshot usage
* Firewall configuration
* `systemd` service setup

### Validator & Wallet Commands

🛠️ **[Validator Command Reference →](https://github.com/Neutaro/Neutaro/blob/main/Instructions/NeutaroValidatorCommands.md)**

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

👉 **[Neutaro State Sync & Fast Node Guide →](https://github.com/johnolofs/timpi/blob/main/Neutaro/statesync.md)**

This guide provides a **complete, reinstall-safe setup**, including:

* Security checklist (read first)
* System requirements and required ports
* Clean reinstall instructions
* Go + Neutaro source build
* `cosmovisor` + `systemd` configuration
* **State Sync (recommended path)**
* Snapshot sync (fallback option)
* Validator creation (final step)
* Sync verification and troubleshooting

> After syncing, your node behaves exactly like a normal full node.
> You may safely create or start a validator once `catching_up = false`.

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
