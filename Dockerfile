# Development image for Barista
#
# Build:
#   docker build -t barista .
#   docker build --platform linux/arm64 -t barista .
#
# Run:
#   docker run -it barista
#   docker run --rm -it barista

FROM alpine:3.23.0

RUN apk add --no-cache \
    bash \
    curl \
    fd \
    file \
    git \
    npm \
    ripgrep

# Install Pi coding agent
RUN set -eux; \
    npm install -g --ignore-scripts @earendil-works/pi-coding-agent@0.85.1; \
    pi --version; \
    npm cache clean --force

# Install mise-en-place
ENV PATH="/root/.local/bin:/root/.local/share/mise/shims:${PATH}"
RUN set -eux; \
    curl -fsSL https://mise.run | MISE_VERSION=2026.9.6 sh; \
    mise --version

# Install project tooling
COPY mise.toml ./mise.toml
RUN set -eux; \
    mise install; \
    go version; \
    rm mise.toml

# Set up environment
ENV LANG=C.UTF-8 \
    GOCACHE=/root/.cache/go-build \
    GOMODCACHE=/root/go/pkg/mod

CMD ["bash"]
