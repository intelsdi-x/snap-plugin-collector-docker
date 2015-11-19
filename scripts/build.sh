#!/bin/bash -e

GITVERSION=`git describe --always`
SOURCEDIR=$1
BUILDDIR=$SOURCEDIR/build
PLUGIN=`echo $SOURCEDIR | grep -oh "pulse-.*"`
ROOTFS=$BUILDDIR/rootfs
BUILDCMD='go build -a -ldflags "-w"'

echo
echo "****  Pulse Plugin Build  ****"
echo

# Disable CGO for builds
export CGO_ENABLED=0

# Clean build bin dir
rm -rf $ROOTFS/*

# Make dir
mkdir -p $ROOTFS

# Build plugin
echo "Source Dir = $SOURCEDIR"
echo "Building Pulse Plugin: $PLUGIN"
$BUILDCMD -o $ROOTFS/$PLUGIN

