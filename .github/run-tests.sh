#!/bin/bash


# Install tools to test our code-quality.
go get -u golang.org/x/lint/golint
go get -u honnef.co/go/tools/cmd/staticcheck

# Run the static-check tool - we ignore errors in goserver/static.go
t=$(mktemp)
staticcheck -checks all ./... | grep -v "no Go files in" | grep -v CamelCase> $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck'"
    cat $t
    rm $t
    exit 1
fi
rm $t


# Run the linter
echo "Launching linter .."
t=$(mktemp)
golint -set_exit_status ./...| grep -v "CamelCase" > $t
if [ -s $t ]; then
    echo "Found errors via 'staticcheck'"
    cat $t
    rm $t
    exit 1
fi
rm $t
echo "Completed linter .."

# At this point failures cause aborts
set -e

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
