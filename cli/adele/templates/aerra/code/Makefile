BINARY_NAME=adeleApp

build:
	@go mod vendor
	@echo "Building..."
	@go build -o tmp/${BINARY_NAME} .
	@echo "Build complete!"

run: build
	@echo "Starting..."
	@./tmp/${BINARY_NAME}

clean:
	@echo "Cleaning..."
	@go clean
	@rm tmp/${BINARY_NAME}
	@echo "Cleaned!"

start: run

stop:
	@echo "Stopping..."
	@-pkill -SIGTERM -f "./tmp/${BINARY_NAME}"
	@PID=$$(pgrep -f "./tmp/${BINARY_NAME}"); \
    while kill -0 $$PID 2>/dev/null; do sleep 0.1; done

restart: stop start

# Migration targets — added by `adele install starter-kit --with-auth`.
migrate\:up:
	@adele migrate up

migrate\:down:
	@adele migrate down

migrate\:down\:all:
	@adele migrate down --all

migrate\:force:
	@adele migrate force
