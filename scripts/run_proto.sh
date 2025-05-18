#!/bin/bash

# In a shell script or Makefile

# Define your module path
GO_MODULE_PATH="xrf197ilz35aq2"
PROTO_DIR="proto"
OUTPUT_DIR="." # Or specify a staging directory if you prefer

# Find all .proto files and execute protoc for each
find "${PROTO_DIR}" -name "*.proto" -print0 | while IFS= read -r -d $'\0' proto_file; do
  echo "Compiling ${proto_file}..."
  protoc \
    --proto_path="${PROTO_DIR}" \
    --go_out="${OUTPUT_DIR}" \
    --go_opt=module="${GO_MODULE_PATH}" \
    --go-grpc_out="${OUTPUT_DIR}" \
    --go-grpc_opt=module="${GO_MODULE_PATH}" \
    "${proto_file}"
done

echo "Protobuf compilation finished."


# --proto_path=proto -> Specifies the directory where protoc should look for the .proto files and any imported .proto files
# --go_out=. -> Specifies the output directory for the generated Go message files (.pb.go)
# --go_opt=xrf197ilz35aq2 -> Tells protoc-gen-go the Go module path. The output dir structure will be created relative
    # to this module path, based on the go_package option in your .proto file
# --go-grpc_out= -> Specifies the output directory for the generated Go gRPC service files (*_grpc.pb.go).
# --go-grpc_opt=module=xrf197ilz35aq2 -> Tells protoc-gen-go-grpc the Go module path for gRPC file generation
