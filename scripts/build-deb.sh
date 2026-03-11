#!/usr/bin/env bash
set -euo pipefail

VERSION=$1
ARCH=$2

if [ -z "${VERSION:-}" ] || [ -z "${ARCH:-}" ]; then
  echo "Usage: $0 <version> <arch>"
  exit 1
fi

VERSION_STR=${VERSION#v}
BINARY="dist/inceptools-linux-${ARCH}"

if [ ! -f "$BINARY" ]; then
  echo "❌ Binary not found: $BINARY"
  exit 1
fi

echo "📦 Building .deb package for inceptools v${VERSION_STR} (${ARCH})..."

STAGING="dist/deb-staging-${ARCH}"

# Clean staging
rm -rf "$STAGING"

# Create structure
mkdir -p "$STAGING/usr/local/bin"
mkdir -p "$STAGING/DEBIAN"

# Copy binary
cp "$BINARY" "$STAGING/usr/local/bin/inceptools"
chmod 755 "$STAGING/usr/local/bin/inceptools"

# Control file
cat <<EOF > "$STAGING/DEBIAN/control"
Package: inceptools
Version: ${VERSION_STR}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: IncepTools <info@inceptools.com>
Description: A developer-friendly database migration CLI for Go
 IncepTools is a lightweight CLI for managing database schema migrations.
 It supports version tracking, rollback, and multi-database compatibility.
EOF

# Optional post-install message
cat <<EOF > "$STAGING/DEBIAN/postinst"
#!/bin/bash
echo "IncepTools installed successfully!"
echo "Run: inceptools --help"
EOF

chmod 755 "$STAGING/DEBIAN/postinst"

# Build package
OUTPUT="dist/inceptools_${VERSION_STR}_${ARCH}.deb"

dpkg-deb --build "$STAGING" "$OUTPUT"

echo "✅ Created: $OUTPUT"

# Cleanup
rm -rf "$STAGING"
