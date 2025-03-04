DEFAULT_KURTOSIS_PACKAGE := 'github.com/ethpandaops/optimism-package@main'
DEFAULT_KURTOSIS_ENCLAVE := 'devnet'

TMPDIR := `mktemp -d`

# This target will compile the solidity contracts required for the tests
[working-directory: 'contracts']
build-contracts:
    forge build

[working-directory: 'contracts']
build-bindings-requirements: build-contracts
    forge inspect MockERC20 abi > {{TMPDIR}}/MockERC20.json
    forge inspect MockERC20 bytecode > {{TMPDIR}}/MockERC20.bytecode

build-bindings: build-bindings-requirements
    mkdir -p bindings/mockERC20
    abigen --abi {{TMPDIR}}/MockERC20.json --bin {{TMPDIR}}/MockERC20.bytecode --pkg mockERC20 --out bindings/mockERC20/bindings.go

build: build-bindings

lint:
    golangci-lint run ./...

lint-fix:
    golangci-lint run ./... --fix

test: build
    go test -v ./...

tidy:
    go mod tidy -x

run-kurtosis ARGS ENCLAVE=DEFAULT_KURTOSIS_ENCLAVE PACKAGE=DEFAULT_KURTOSIS_PACKAGE:
    echo "Starting kurtosis enclave {{ENCLAVE}} from package {{PACKAGE}} and args file {{ARGS}}"
    kurtosis run --enclave {{ENCLAVE}} --args-file {{ARGS}} {{PACKAGE}}