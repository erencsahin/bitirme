# Build stage
FROM rustlang/rust:nightly-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache musl-dev openssl-dev pkgconfig

# Copy manifests
COPY Cargo.toml ./

# Create dummy main to cache dependencies
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Copy real source code
COPY . .

# Build for release
RUN cargo build --release

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache libgcc

# Copy binary from builder
COPY --from=builder /app/target/release/payment-service .

# Copy .env if exists
COPY .env* ./

EXPOSE 8085

CMD ["./payment-service"]