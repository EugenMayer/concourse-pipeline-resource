#!/bin/sh

set -e

# Vars
CONCOURSE_URL=$1
CONCOURSE_TEAM=$2
CONCOURSE_USER=$3
CONCOURSE_PASS=$4
INSECURE=$5
CONCOURSE_TARGET=$6
FLY_ENDPOINT="/api/v1/cli?arch=amd64&platform=linux"
AUTH_ENDPOINT="/sky/login"

# Get fly binary
TMP_DIR=$(mktemp -d) && trap "rm -rf ${TMP_DIR}" EXIT
export PATH=$PATH:/opt/resource/:/usr/local/bin
FLY_BIN="fly"

if [ ! -x "$(command -v fly)" ]; then
    echo "Fly bin does not exists at $FLY_BIN - fetching new version from ${CONCOURSE_URL}/api/v1/cli?arch=amd64&platform=linux"
    mkdir -p /usr/local/bin
    curl -o "/usr/local/bin/fly" "${CONCOURSE_URL}/api/v1/cli?arch=amd64&platform=linux"
    chmod a+x "/usr/local/bin/fly"
fi

EXTRA_PARAMS=""
if [ "$INSECURE" = "true" ]; then
    echo "allowing insecure"
    EXTRA_PARAMS="${EXTRA_PARAMS} -k"
fi

COOKIE_FILE="${TMP_DIR}/cookie.txt"
echo "getting FORM token"
FORM_TOKEN="$(curl -b ${COOKIE_FILE} -c ${COOKIE_FILE} -s -o /dev/null -L "${CONCOURSE_URL}${AUTH_ENDPOINT}" -D - | \
    grep -i "Location: /sky/issuer/auth" | cut -d ' ' -f 2 | tr -d '\r')"
if [ -z "${FORM_TOKEN}" ];then
  echo "could not retrieve FORM token"
  exit 1
fi

echo "getting OAUTH token"
curl -o /dev/null -s -b ${COOKIE_FILE} -c ${COOKIE_FILE} -L --data-urlencode "login=${CONCOURSE_USER}" \
    --data-urlencode "password=${CONCOURSE_PASS}" "${CONCOURSE_URL}${FORM_TOKEN}"
OAUTH_TOKEN=$(cat ${COOKIE_FILE} | grep 'skymarshal_auth' | grep -o 'Bearer .*$' | tr -d '"')

if [ -z "$OAUTH_TOKEN" ];then
  echo "could not retrieve OAuth token"
  exit 1
fi

FLYCMD="$FLY_BIN -t ${CONCOURSE_TARGET} login -c ${CONCOURSE_URL} -n ${CONCOURSE_TEAM} $EXTRA_PARAMS"
echo "prepared fly login cmd: '$FLYCMD'"
echo "Running fly login with OAuth token"
echo "${OAUTH_TOKEN}" | ${FLYCMD} > /dev/null
echo "Success"
