
test:
	go test ./...

run:
	SESSION_SECRET="1234567890" go run main/api.go

image:
	docker build -t events-api .
