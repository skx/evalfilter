#!/bin/bash

# Remove our WASM test.
if [ -d "wasm/"]; then
    rm -rf wasm/
fi

# Install tools to test our code-quality.
go get -u golang.org/x/lint/golint
go get -u honnef.co/go/tools/cmd/staticcheck

# Run the static-check tool
t=$(mktemp)
staticcheck -checks all ./... > $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck'"
    cat $t
    rm $t
    exit 1
fi
rm $t



# At this point failures cause aborts
set -e

# Run the linter-tool
echo "Launching linter .."
golint -set_exit_status ./...
echo "Completed linter .."

# Run the vet-tool
echo "Running go vet .."
go vet ./...
echo "Completed go vet .."

# Run our golang tests
go test ./...

# If that worked build our examples, to ensure they work
# and that we've not broken compatibility
for i in _examples/embedded/*; do
    if [ -d $i ]; then
        pushd $i
        echo "Building example in $(pwd)"
        go build .
        popd
    else
        echo "Skipping non-directory $i"
    fi
done

# Finally run our benchmarks for completeness
go test -test.bench=evalfilter -benchtime=1s -run=^t
