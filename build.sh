#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-dev}"
BINARY="horizon"
LDFLAGS="-s -w -X main.Version=${VERSION}"
GCFLAGS="all=-trimpath=${GOPATH:-$HOME/go}"
BUILD_DIR="./build"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    output="${BINARY}"
    archive="${BUILD_DIR}/${GOOS}-${GOARCH}-${BINARY}"

    if [ "${GOOS}" = "windows" ]; then
        output="${BINARY}.exe"
    fi

    echo "Building ${GOOS}/${GOARCH}..."
    CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
        -ldflags "${LDFLAGS}" \
        -gcflags "${GCFLAGS}" \
        -trimpath \
        -o "${BUILD_DIR}/${output}" \
        ./cmd/horizon/

    if [ "${GOOS}" = "windows" ]; then
        (cd "${BUILD_DIR}" && zip -q "${GOOS}-${GOARCH}-${BINARY}.zip" "${output}" && rm "${output}")
    else
        tar -czf "${archive}.tar.gz" -C "${BUILD_DIR}" "${output}"
        rm "${BUILD_DIR}/${output}"
    fi
done

echo ""
echo "Artifacts:"
ls -lh "${BUILD_DIR}"/*
