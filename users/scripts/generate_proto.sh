#!/bin/bash

# Generate protobuf code for Go
# This script generates Go code from the protobuf definitions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating protobuf code...${NC}"

# Create output directory
mkdir -p pkg/grpc/users/v1

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
  echo -e "${RED}Error: protoc (Protocol Buffer Compiler) is not installed${NC}"
  echo "Please install it using:"
  echo "  - Ubuntu/Debian: sudo apt-get install protobuf-compiler"
  echo "  - macOS: brew install protobuf"
  echo "  - Or download the binary from: https://github.com/protocolbuffers/protobuf/releases"
  echo "    Choose the release that matches your OS and architecture"
  echo "    Extract and move the protoc binary to /usr/local/bin"
  echo "    Move the include/google folder to /usr/local/include"
  exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Generate Go code from protobuf files
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    proto/users/v1/*.proto

# Move generated files to the correct directory
mkdir -p pkg/grpc/users/v1
mv proto/users/v1/*.pb.go pkg/grpc/users/v1/ 2>/dev/null || true

echo -e "${GREEN}✓ Protobuf code generation completed successfully${NC}"
echo -e "${GREEN}✓ Generated files in pkg/grpc/users/v1/${NC}"
