FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    build-essential \
    libpam0g-dev \
    curl \
    tar \
    && rm -rf /var/lib/apt/lists/*

# Verify pam_modules.h exists
RUN test -f /usr/include/security/pam_modules.h

RUN curl -fsSL https://go.dev/dl/go1.24.5.linux-amd64.tar.gz -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /src

COPY . .

ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=arm64

RUN go mod tidy

RUN ls /usr/include/security/pam_modules.h
# Build the shared object using musl-gcc to get static libc linking as much as possible
RUN go build -buildmode=c-shared -o pam_jit_pg.so
