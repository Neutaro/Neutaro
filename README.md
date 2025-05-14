# 🌐 Neutaro – Validator Setup & Governance Tools

Welcome to the official repository for **Neutaro**, the blockchain that powers governance and decentralization for the Timpi search engine.  
This repo includes everything you need to run a **Neutaro validator**, **delegate tokens**, and **contribute to the Timpi ecosystem**.

We provide:

✅ Step-by-step installation guide  
🔐 Validator security best practices  
🗳️ Governance, staking & proposal voting  
📁 Source build instructions and systemd setup

---

## 💡 Why Neutaro?

Neutaro is the governance and reward chain backing **Timpi** — the first decentralized search engine.  
As a Neutaro participant, you can:

- 🛠 Run a validator or node to support the network
- 💸 Stake tokens and earn rewards
- 🗳 Vote on proposals — including **ethical decisions** that affect the Timpi search engine’s direction

---

## 🔧 Validator Overview

Validators play a critical role by securing the Neutaro chain and processing transactions.  
By running a validator, you:

- Ensure network integrity
- Participate in governance
- Earn staking rewards

---

## 🐧 Validator for Linux (Ubuntu 22.04.03+)

👉 **[Neutaro Validator Setup Guide](https://github.com/Neutaro/Neutaro/blob/main/Instructions/NeutaroInstallation.md)**  
Includes full source build, `cosmovisor` setup, pruning, snapshots, firewall rules, and systemd instructions.

---

## 💡 Delegate Instead of Validating?

Not ready to run a full validator?  
You can **delegate your tokens** to an active validator and still earn rewards:

```shell
Neutaro tx staking delegate <validator_address> 100000000uneutaro --from YOURWALLET --chain-id Neutaro-1
