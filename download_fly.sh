#!/usr/bin/env bash

echo "downloading latest fly"
url=$(curl -s "https://api.github.com/repos/concourse/concourse/releases/latest?access_token=$GITHUB_TOKEN" \
    | jq -r '.assets[] | select(.name | test("fly_linux_amd64$")) | .browser_download_url') &&\
    curl -L "$url" -o bin/fly

