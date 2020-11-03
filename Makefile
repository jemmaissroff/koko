bench:
	cd benchmark && go test -bench=. -benchtime=30s
fmt:
	go fmt ./...
test:
	go test ./...
	