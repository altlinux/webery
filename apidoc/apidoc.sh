#!/bin/sh -efu

cwd="$(readlink -ev "${0%/*}")"

exec go run apidoc.go \
	-dir "$cwd"/../pkg/ahttp/api \
	-template-dir "$cwd" \
	> "$cwd"/../static/apidoc.html
