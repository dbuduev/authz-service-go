build:
	go build main.go

test:
	go test ./...

cover:
	go test -coverprofile=c.out ./...
