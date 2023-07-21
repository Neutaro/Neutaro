# This is the Source code for the TimpiChain on the Testnet.

# To get started you have to
- use Linux
- copy this repo
- install go (https://go.dev/doc/install) version 1.18.1 works.
- go to /cmd/TimpiChain
- execute "go build"
- now you can start setting up the node.

# To run a standalone node to play around with

- ./TimpiChain init TimpiChain --chain-id WhatYouWant
- ./TimpiChain keys add MainValidator --keyring-backend test
- ./TimpiChain add-genesis-account MainValidator 4000000000stake --keyring-backend test
- ./TimpiChain gentx MainValidator 1000000stake --keyring-backend test --chain-id WhatYouWant
- ./TimpiChain collect-gentxs
- ./TimpiChain start

# To run a linked node you do
- ./TimpiChain init TimpiChain

copy the genesis.json file from http://173.249.54.208/genesis.json into root/.TimpiChain/config and replace the old one.
edit the config.toml in the same directory and change 

[rpc]

TCP or UNIX socket address for the RPC server to listen on
laddr = "tcp://127.0.0.1:26657"

to

[rpc]

TCP or UNIX socket address for the RPC server to listen on
laddr = "tcp://0.0.0.0:26657"

and add "9d63a46f1af8eaffb66831701cc1b22fab0429d7@173.249.54.208:26656" to
persistent_peers = "" like so persistent_peers = "9d63a46f1af8eaffb66831701cc1b22fab0429d7@173.249.54.208:26656"

- ./TimpiChain start

# To run a validator node you
- you have to stake at least 1 timpiTN ( 1.000.000utimpiTN ), but also be in the top 120 on all staking validators.
- ./TimpiChain tx staking create-validator --amount=1500000utimpiTN --pubkey=$(./TimpiChain tendermint show-validator)  --moniker=WhatYouWant --chain-id=TimpiChainTN --from YourWallet --keyring-backend test --commission-rate="0.10" --commission-max-rate="0.20" --commission-max-change-rate="0.01" --min-self-delegation="1000000" --gas="auto" --gas-prices="0.0025utimpiTN" --gas-adjustment="1.5"

you can check the validators via 
- ./TimpiChain query staking validators

and the active ( top 100 ) ones with 
- ./TimpiChain q tendermint-validator-set

if you want / have to increase your staked tokens you can use
- ./TimpiChain tx staking delegate ValidatorAddress 1000000utimpiTN --from YourWallet --chain-id TimpiChainTN --keyring-backend test
1000000 being however much you want to delegate.

# To use the faucet
- add a key to your node and use this link http://173.249.54.208:1337/YOURWALLET. Replace YOURWALLET with your address.

if you use the faucet too quickly you will be timed out before getting your token limit. ./TimpiChain q bank balances YOURWALLET check after every call if you got the tokens and then call the API again when you did.

