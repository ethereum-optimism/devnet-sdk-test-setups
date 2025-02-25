DEFAULT_KURTOSIS_PACKAGE := 'github.com/ethpandaops/optimism-package@main'
DEFAULT_KURTOSIS_ENCLAVE := 'devnet'

lint:
    golangci-lint run ./...

lint-fix:
    golangci-lint run ./... --fix

test:
    go test -v ./...

tidy:
    go mod tidy -x

run-kurtosis ARGS ENCLAVE=DEFAULT_KURTOSIS_ENCLAVE PACKAGE=DEFAULT_KURTOSIS_PACKAGE:
    echo "Starting kurtosis enclave {{ENCLAVE}} from package {{PACKAGE}} and args file {{ARGS}}"
    kurtosis run --enclave {{ENCLAVE}} --args-file {{ARGS}} {{PACKAGE}}