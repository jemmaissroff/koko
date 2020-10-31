bench:
	cd benchmark && go test -bench=. -benchtime=10s
fmt:
	go fmt ./...
test:
	go test ./...
	