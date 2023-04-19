# Build the application from source
FROM golang:1.19-alpine AS build-stage

# docker build -f Dockerfile.build . --target dependencies # -t papi:deps .
WORKDIR /bridge

# Run Go file as a script
ENTRYPOINT ["/usr/local/go/bin/go", "run", "main.go"]
