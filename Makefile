build: 
	go build main.go

test: build
	go test ./...

cover:
	go test -coverprofile=c.out ./...
