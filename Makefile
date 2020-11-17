bench:
	cd benchmark && go test -bench=. -benchtime=30s
fmt:
	go fmt ./...
profile:
	cd benchmark && go test -cpuprofile cpu.prof -memprofile mem.prof -bench=. -benchtime=30s
test:
	go test ./...
	