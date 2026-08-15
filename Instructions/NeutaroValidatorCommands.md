# ⚙️ Neutaro Validator CLI Command Guide

This page contains essential CLI commands for running and managing a Neutaro validator, delegating, voting, and handling unbonding/transfer of tokens.

> 🧠 All commands assume you are running on **Ubuntu 22.04.4+** with the `Neutaro` binary installed and configured. Replace placeholders (`YourWallet`, `ValidatorAddress`, etc.) with your actual data.
>
> ⚠️ **Never run `Neutaro keys` or `tx` commands with `sudo`** — sudo uses *root's* keyring, and
> your wallet will seem to vanish for every command run without it. Use the same
> `--keyring-backend` you created the key with, consistently.

---

## 🟢 Start, Restart, and Monitor Your Validator

```shell
# Start the validator service
sudo systemctl start Neutaro

# Restart the validator service
sudo systemctl restart Neutaro

# Monitor validator logs in real time
sudo journalctl -fu Neutaro -o cat
```

---

## ✏️ Edit Validator Settings

```shell
Neutaro tx staking edit-validator \
  --from YourWallet \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  [flags...]
```

### Common Flags

| Action                   | Example                                                        |
| ------------------------ | -------------------------------------------------------------- |
| Change Moniker           | `--new-moniker "JohnOlofs"`                                    |
| Change Commission Rate   | `--commission-rate 0.10`                                       |
| Set Keybase Identity     | `--identity "EAFA3628665BE045"`                                |
| Set Security Contact     | `--security-contact "johnolof@timpi.se"`                       |
| Set Website              | `--website "https://nordicnodes.blogspot.com/"`                |
| Set Description          | `--details "A reliable and trusted validator."`                |
| Minimum Self Delegation  | `--min-self-delegation 1000000`                                |
| Set Gas | `--gas auto --gas-adjustment 1.4 --gas-prices 0.025uneutaro` |
| Optional Memo            | `--memo "Updating validator settings"`                         |

### ✅ Full Example

```shell
Neutaro tx staking edit-validator \
  --from YourWallet \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  --new-moniker "JohnOlofs" \
  --commission-rate 0.10 \
  --identity "EAFA3628665BE045" \
  --security-contact "johnolof@timpi.se" \
  --website "https://nordicnodes.blogspot.com/" \
  --details "A reliable and trusted validator." \
  --min-self-delegation 1000000 \
  --gas auto \
  --gas-adjustment 1.4 \
  --gas-prices 0.025uneutaro \
  --memo "Updating validator settings"
```

> ℹ️ `--broadcast-mode block` was **removed in Cosmos SDK 0.47** — commands using it fail. The
> default (`sync`) returns a tx hash immediately; confirm with
> `Neutaro query tx <hash>` a few seconds later.

---

## 🚫 Unjail a Validator

```shell
Neutaro tx slashing unjail \
  --from YourWallet \
  --gas-adjustment 1.4 \
  --gas auto \
  --gas-prices 0.025uneutaro \
  --chain-id Neutaro-1 \
  --keyring-backend os
```

---

## 💸 Delegation Commands

### Delegate Tokens

```shell
Neutaro tx staking delegate <ValidatorAddress> <Amount>uneutaro \
  --from YourWallet \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  --gas auto \
  --gas-prices 0.025uneutaro
```

### Redelegate Tokens

```shell
Neutaro tx staking redelegate <FromValidatorAddress> <ToValidatorAddress> <Amount>uneutaro \
  --from YourWallet \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  --gas auto \
  --gas-prices 0.025uneutaro
```

---

## 🗳 Vote on Proposals

```shell
# Vote Yes
Neutaro tx gov vote <ProposalID> yes \
  --from YourWallet \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  --gas auto \
  --gas-prices 0.025uneutaro

# Vote No
Neutaro tx gov vote <ProposalID> no ...
```

---

## 🔓 Unbond Tokens

```shell
Neutaro tx staking unbond <ValidatorAddress> <Amount>uneutaro \
  --from YourWallet \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  --gas auto \
  --gas-prices 0.025uneutaro
```

🔎 Then wait for the unbonding period (typically 21 days).
Check balance afterwards:

```shell
Neutaro query bank balances <YourWalletAddress>
```

---

## 🔁 Send Tokens (After Unbonding)

```shell
Neutaro tx bank send <FromWallet> <ToWallet> <Amount>uneutaro \
  --chain-id Neutaro-1 \
  --keyring-backend os \
  --gas auto \
  --gas-prices 0.025uneutaro
```

---

## ⚠️ Notes on `--gas` and `--gas-prices`

> Some commands may fail depending on network or local configurations when using:

```shell
--gas auto --gas-prices 0.025uneutaro
```

> If that happens, remove those flags and try again manually with `--gas 200000`.

---

## ✅ Helpful Utilities

```shell
# View all keys in keyring
Neutaro keys list --keyring-backend os

# Query sync status
Neutaro status 2>&1 | jq .SyncInfo
```

---

📌 Feel free to bookmark this page or print it out as your **Validator Command Cheat Sheet**.

```
