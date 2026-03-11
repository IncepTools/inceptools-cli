#!/bin/bash
set -e

VERSION=$1
ARCH=$2 # amd64 or arm64
BINARY="dist/inceptools-linux-$ARCH"

if [ -z "$VERSION" ] || [ -z "$ARCH" ]; then
    echo "Usage: $0 <version> <arch>"
    exit 1
fi

echo "📦 Building .deb package for inceptools v$VERSION ($ARCH)..."

STAGING="dist/deb-staging-$ARCH"
mkdir -p "$STAGING/usr/local/bin"
mkdir -p "$STAGING/DEBIAN"

cp "$BINARY" "$STAGING/usr/local/bin/inceptools"
chmod +x "$STAGING/usr/local/bin/inceptools"

cat <<EOF > "$STAGING/DEBIAN/control"
Package: inceptools
Version: ${VERSION#v}
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: IncepTools <info@inceptools.com>
Description: A powerful, developer-friendly database migration CLI for Go
 inceptools is a lightweight CLI tool that manages database schema migrations
 with automatic version tracking, rollback support, and multi-database compatibility.
EOF

dpkg-deb --build "$STAGING" "dist/inceptools_${VERSION#v}_$ARCH.deb"

echo "✅ Created: dist/inceptools_${VERSION#v}_$ARCH.deb"
rm -rf "$STAGING"
