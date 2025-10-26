#!/bin/bash

# Generate code for each API
for file in ./api/http/*.yaml; do
    f=$(basename $file .yaml)
    echo "Generating code for $f..."
    
    # Create directory if it doesn't exist
    mkdir -p ./internal/app/$f/port/genhttp

    # Generate types
    oapi-codegen -generate types \
        -o ./internal/app/$f/port/genhttp/openapi_types.gen.go \
        -package genhttp \
        api/http/$f.yaml

    # Generate server (standard net/http server)
    oapi-codegen -generate std-http-server \
        -o ./internal/app/$f/port/genhttp/openapi_server.gen.go \
        -package genhttp \
        api/http/$f.yaml

done