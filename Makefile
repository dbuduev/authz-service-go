build: 
	go build main.go

dynamodb:
	./scripts/run_dynamodb

create_table: dynamodb
	./scripts/create_table

test: build create_table
	go test ./...

cover:
	go test -coverprofile=c.out ./...
