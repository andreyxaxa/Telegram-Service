BASE_STACK = docker compose -f docker-compose.yml

compose-up: ### Run docker compose
	$(BASE_STACK) up --build -d
.PHONY: compose-up

compose-down: ### Down docker compose
	$(BASE_STACK) down -v
.PHONY: compose-down

proto-v1: ### generate source files from proto
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		docs/proto/v1/*.proto
.PHONY: proto-v1

deps: ### deps tidy + verify
	go mod tidy && go mod verify
.PHONY: deps