# Export all environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Build the application
all: build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if [ -x "$(GOPATH)/bin/air" ]; then \
	    "$(GOPATH)/bin/air"; \
		@echo "Watching...";\
	else \
	    read -p "air is not installed. Do you want to install it now? (y/n) " choice; \
	    if [ "$$choice" = "y" ]; then \
			go install github.com/cosmtrek/air@latest; \
	        "$(GOPATH)/bin/air"; \
				@echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

# Build Tailwind
css:
	npx tailwindcss -i assets/css/tailwind.css -o assets/css/styles.css --minify

# Live update Tailwind styles
css-watch:
	npx tailwindcss -i assets/css/tailwind.css -o assets/css/styles.css --watch

# Migrations
migrate-up: 
	migrate -path internal/database/migrations -database "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE}?sslmode=disable" -verbose up

migrate-down: 
	migrate -path internal/database/migrations -database "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE}?sslmode=disable" -verbose down


.PHONY: all build run test clean css css-watch migration-up migration-down
