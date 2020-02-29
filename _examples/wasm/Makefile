all: lib.wasm wasm_exec.js

clean:
	rm -f wasm_exec.js lib.wasm

lib.wasm: main.go
	GOOS=js GOARCH=wasm go build -o lib.wasm

wasm_exec.js:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" .
