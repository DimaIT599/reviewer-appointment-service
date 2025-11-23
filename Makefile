.PHONY: build run test migrate-up migrate-down docker-up docker-down clean

build:
	go build -o bin/reviewer-appointment-service ./cmd/reviewer-appointment-service

run:
	go run ./cmd/reviewer-appointment-service

test:
	go test -v ./...

test-coverage:
	go test -v -cover ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

clean:
	rm -rf bin/


