bench:
	cd benchmark && go test -bench=. -benchtime=30s
fmt:
	go fmt ./...
profile:
	cd benchmark && go test -cpuprofile cpu.prof -memprofile mem.prof -bench=. -benchtime=30s && go tool pprof -svg cpu.prof
test:
	# -count=1 forces go to re-run tests instead of caching them
	go test -count=1 ./...
wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm main_wasm.go 
	