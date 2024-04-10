# Neutaro interchaintest

These tests were seeded primarily by copying and modifying the excellent test suite they have at Juno: https://github.com/CosmosContracts/juno/tree/main/interchaintest (version used: https://github.com/CosmosContracts/juno/tree/e98863bf7112f4b117a2114e22f7482367362764/interchaintest)
That includes (for now) the contracts in the contracts folder as well.

The tests can be run like any go test, with a few requirements:
* Docker
* neutaro:local built (`docker build -t neutaro:local .` at root of the repo)