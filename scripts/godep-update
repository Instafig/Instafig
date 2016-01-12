#!/bin/bash

which godep &>/dev/null || go get github.com/tools/godep

function go-get() {
    local PKG=$1
    if [ ${PKG:0:13} = 'golang.org/x/' ] ; then
        # using github repo for golang.org/x/..., thanks to GFW
        PKG=${PKG:13}
        PKG=${PKG%/*}
        local GIT_REPO="https://github.com/golang/${PKG}.git"
        mkdir -p ../../../golang.org/x/
        echo git clone $GIT_REPO
        git clone $GIT_REPO ../../../golang.org/x/$PKG
    else
        echo go get $PKG
        go get $PKG
    fi
}

[ -n "$TRAVIS_GO_VERSION" ] && { # just use godep restore for TRAVIS
    echo godep restore
    godep restore
    exit 0
}

while :; do
    OUTPUT=$(godep save ./... 2>&1)
    [ "$OUTPUT" = "" ] && break
    OUTPUT=${OUTPUT#*\(}
    OUTPUT=${OUTPUT%\)*}
    go-get $OUTPUT
    [ $? = 0 ] || {
        echo "error on `go get $OUTPUT`"
        exit 1
    }
done

if [ "$1" = update-version ]; then
    godep update
fi
