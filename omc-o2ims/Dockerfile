# Build stage
FROM golang:1.23.6 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

# Copy source code
COPY . .

#run test before build
RUN make test

# Build
RUN make build 


# Step 2: Create a smaller distroless image
FROM gcr.io/distroless/base
# Copy the Go binary from the builder image


WORKDIR /app
COPY --from=builder /app/build/omc-o2ims /app/server

# Bind-mount directory with configs including kubeconfig
VOLUME /app/config

# Copy the kubeconfig file from the configs
COPY config/config.yaml /app/config/config.yaml
COPY config/crd-list/ /app/config/crd-list/

#FIXME during the final build this will not be avaialble
# COPY config/kubeconfig /app/config/kubeconfig


EXPOSE 8080

ENTRYPOINT ["/app/server"]