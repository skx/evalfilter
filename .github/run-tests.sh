#!/bin/bash


# Install tools to test our code-quality.
go get -u golang.org/x/lint/golint
go get -u honnef.co/go/tools/cmd/staticcheck

# Run the static-check tool - we ignore errors in goserver/static.go
t=$(mktemp)
staticcheck -checks all ./... | grep -v "no Go files in" > $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck'"
    cat $t
    rm $t
    exit 1
fi
rm $t

# At this point failures cause aborts
set -e

# Run the linter
echo "Launching linter .."
golint -set_exit_status ./...
echo "Completed linter .."

# Run golang tests
go test ./...

# If that worked build our examples, to ensure they work
# and that we've not broken compatibility
for i in _examples/*; do
    pushd $i
    echo "Building example in $(pwd)"
    go build .
    popd
done
