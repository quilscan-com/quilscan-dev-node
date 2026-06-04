# syntax=docker.io/docker/dockerfile:1.7-labs
FROM --platform=${TARGETPLATFORM} ubuntu:24.04 AS base

ARG GO_VERSION=1.23.5

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="${PATH}:/root/.cargo/bin/"

RUN apt-get update && apt-get install -y \
    autoconf \
    automake \
    build-essential \
    ca-certificates \
    clang \
    cmake \
    curl \
    g++ \
    gcc \
    git \
    git-lfs \
    libflint-dev \
    libgmp-dev \
    libmpfr-dev \
    libmpfr6 \
    libssl-dev \
    libtool \
    m4 \
    make \
    pkg-config \
    protobuf-compiler \
    wget \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

FROM base AS gmp-builder

ARG GMP_VERSION=6.3.0
RUN set -eux; \
    archive="gmp-${GMP_VERSION}.tar.xz"; \
    for url in \
      "https://ftp.gnu.org/gnu/gmp/${archive}" \
      "https://ftpmirror.gnu.org/gmp/${archive}" \
      "https://gmplib.org/download/gmp/${archive}"; do \
      if curl -fL --retry 3 --connect-timeout 20 --max-time 120 "$url" -o "$archive"; then \
        break; \
      fi; \
      rm -f "$archive"; \
    done; \
    test -f "$archive" && \
    tar -xf gmp-${GMP_VERSION}.tar.xz && \
    cd gmp-${GMP_VERSION} && \
    ./configure \
      --prefix=/usr/local \
      --enable-static \
      --disable-shared \
      --with-pic \
      CFLAGS="-O3" && \
    make -j"$(nproc)" && \
    make install

FROM gmp-builder AS flint-builder

RUN git clone --branch flint-3.0 --depth 1 https://github.com/flintlib/flint.git && \
    cd flint && \
    ./bootstrap.sh && \
    ./configure \
      --prefix=/usr/local \
      --with-gmp=/usr/local \
      --with-mpfr=/usr/local \
      --enable-static \
      --disable-shared \
      CFLAGS="-O3 -fPIC" && \
    make -j"$(nproc)" && \
    make install

FROM base AS emp-builder

WORKDIR /opt/ceremonyclient
COPY upstream/install-emp.sh .
COPY upstream/emp-tool emp-tool
COPY upstream/emp-ot emp-ot

RUN bash install-emp.sh
RUN cd emp-tool && \
    sed -i 's/add_library(${NAME} SHARED ${sources})/add_library(${NAME} STATIC ${sources})/g' CMakeLists.txt && \
    mkdir build && \
    cd build && \
    cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local && \
    cd .. && \
    make -j"$(nproc)" && \
    make install
RUN cd emp-ot && \
    mkdir build && \
    cd build && \
    cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local && \
    cd .. && \
    make -j"$(nproc)" && \
    make install

FROM base AS go-base

ARG GO_VERSION
RUN ARCH="$(dpkg --print-architecture)" && \
    case "$ARCH" in \
      amd64) GOARCH=amd64 ;; \
      arm64) GOARCH=arm64 ;; \
      *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    wget "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GOARCH}.tar.gz" && \
    rm "go${GO_VERSION}.linux-${GOARCH}.tar.gz"

FROM base AS rust-base

COPY upstream/docker/rustup-init.sh /opt/rustup-init.sh
RUN /opt/rustup-init.sh -y --profile minimal
RUN cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go --tag v0.4.0+v0.28.3

FROM base AS linux-warm-builder

COPY --from=flint-builder /usr/local /usr/local
COPY --from=emp-builder /usr/local /usr/local
COPY --from=go-base /usr/local/go /usr/local/go
COPY --from=rust-base /root/.cargo /root/.cargo
COPY --from=rust-base /root/.rustup /root/.rustup
COPY build-official-node.sh /usr/local/bin/quilscan-build-node-linux-official-local

ENV PATH="${PATH}:/usr/local/go/bin:/root/.cargo/bin/"
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOEXPERIMENT=arenas
ENV QUILIBRIUM_SIGNATURE_CHECK=true

RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

COPY upstream /opt/monorepo
WORKDIR /opt/monorepo

RUN set -eux; \
    chmod +x /usr/local/bin/quilscan-build-node-linux-official-local; \
    git lfs version; \
    git lfs install --local --force; \
    git rev-parse HEAD; \
    quilscan-build-node-linux-official-local .; \
    test -f node/build/amd64_linux/node

ENV QUILSCAN_WARM_BUILDER_SOURCE_DIR=/opt/monorepo
CMD ["/bin/bash"]
