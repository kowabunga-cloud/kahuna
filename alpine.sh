#!/bin/bash

sed -i "s%pkgver=\(.*\)-%pkgver=${VERSION}%g" APKBUILD
abuild -r
