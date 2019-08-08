#!/usr/bin/env bash

echo "downloading latest fly"
url=$(curl -s "https://api.github.com/repos/concourse/concourse/releases/latest" \
    | jq -r '.assets[] | select(.name | match("fly-.*linux-amd64[.]tgz$"; "g")) | .browser_download_url') && \
    curl -L "$url" | tar xvz -C bin

