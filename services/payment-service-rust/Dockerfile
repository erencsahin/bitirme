# Build stage
FROM rust:1.75-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache musl-dev openssl-dev

# Copy manifests
COPY Cargo.toml Cargo.lock ./

# Create dummy main to cache dependencies
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Copy actual source
COPY . .

# Build application
RUN cargo build --release

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache libgcc

# Copy binary
COPY --from=builder /app/target/release/payment-service /app/payment-service

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8085

CMD ["/app/payment-service"]