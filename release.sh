#!/bin/bash

set -e

version=$1

echo "checking ..."
if [ "$1" == "" ]; then
	echo "Version required!"
	exit 1
fi

if [ -d "$version" ]; then
	rm -r $version
fi

mkdir $version

##
cat > ./command/version.go <<EOF
package main

const appVersion = "$version"
EOF

echo "building program ..."
go build -o onepw

echo "coping files ..."
cp onepw $version
cp README.md $version
cp CHANGELOG.md $version

echo "tar ..."
target=$version-`uname -s | tr '[:upper:]' '[:lower:]'`-`uname -m`.tar.gz
tar zcf $target $version
rm -r $version

echo "released: $target"
