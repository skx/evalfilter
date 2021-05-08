#!/bin/bash

# Remove our WASM test.
if [ -d "wasm/" ]; then
    rm -rf wasm/
fi

# Install tools to test our code-quality.
go get -u golang.org/x/lint/golint
go get -u honnef.co/go/tools/cmd/staticcheck

# Run the static-check tool
#
# Ignore ST1003 - "should not use underscores in Go names;
t=$(mktemp)
staticcheck -checks all ./...  | grep -v ST1003 > $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck'"
    cat $t
    rm $t
    exit 1
fi
rm $t


# Run the linter-tool
echo "Launching linter .."
golint ./... | grep -v underscores > $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck'"
    cat $t
    rm $t
    exit 1
fi
echo "Completed linter .."


# At this point failures cause aborts
set -e

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

# Build the helper
start=$(pwd)
cd cmd/evalfilter && go build .
cd ${start}

# Now make sure there are no errors in our examples
for src in _examples/scripts/*.script; do

    # Is there a JSON file too?
    name=$(basename ${src} .script)

    if [ -e "_examples/scripts/${name}.json" ]; then
        echo "Running ${src} with JSON input ${name}.json"
        ./cmd/evalfilter/evalfilter run -json "_examples/scripts/${name}.json" ${src}
    else
        echo "Running ${src}"
        ./cmd/evalfilter/evalfilter run ${src}
    fi

done

# Finally run our benchmarks for completeness
go test -test.bench=evalfilter -benchtime=1s -run=^t
