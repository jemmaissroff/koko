bench:
	cd benchmark && go test -bench=. -benchtime=30s
fmt:
	go fmt ./...
test:
	go test ./...
wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm main_wasm.go 
	