# Used for interchaintest. Build with docker build -t neutaro:local .
FROM golang:1.20-bookworm

RUN apt-get update -yq \
            && apt-get install --no-install-recommends -yq \
            wget gnupg ca-certificates gcc g++ make git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code

# Download dependencies and CosmWasm libwasmvm if found.
ADD go.mod go.sum ./
RUN set -eux; \    
    export ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \
    fi; \
    go mod download;

# Copy over code
COPY . /code/
RUN ls -la
RUN cd cmd/Neutaro/ && go build
RUN cp cmd/Neutaro/Neutaro /usr/local/bin

EXPOSE 26656 26657 1317 9090
CMD Neutaro start --rpc.laddr tcp://0.0.0.0:26657
