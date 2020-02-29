# WASM Demo

This directory contains a simple demo of running the scripting engine
via a browser!

## Compile / Install

Compile the code via `make`, or by running:

```
GOOS=js GOARCH=wasm go build -o lib.wasm
```

Copy the contents of `wasm_exec.js` to the same directory:

```
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

Now you can serve the contents of this directory via any HTTP-server and view the results
