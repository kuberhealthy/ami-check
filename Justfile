IMAGE := "kuberhealthy/ami-check"
TAG := "latest"

# Build the AMI check container locally.
build:
	podman build -f Containerfile -t {{IMAGE}}:{{TAG}} .

# Run the unit tests for the AMI check.
test:
	go test ./...

# Build the AMI check binary locally.
binary:
	go build -o bin/ami-check ./cmd/ami-check
