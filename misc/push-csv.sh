#!/bin/bash

set -e

# QUAY_REPO ?= quay.io/siji
# OPERATOR_NAME ?= nginx-operator
# VERSION ?= 0.0.1

DELETE_EXIST=${DELETE_EXIST:-no}
QUAY_ORG=${QUAY_REPO##*/}
BUNDLE_DIR=deploy/olm-catalog/${OPERATOR_NAME}

[[ "X$QUAY_USERNAME" == "X" ]] && read -rp "Enter username quay.io: " QUAY_USERNAME
[[ "X$QUAY_PASSWORD" == "X" ]] && read -rsp "Enter password quay.io: " QUAY_PASSWORD && echo

# Fetch authentication token used to push to Quay.io
AUTH_TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{
    "user": {
        "username": "'"${QUAY_USERNAME}"'",
        "password": "'"${QUAY_PASSWORD}"'"
    }
}' | awk -F'"' '{print $4}')

function cleanup() {
    rm -f bundle.tar.gz
}
trap cleanup EXIT

tar czf bundle.tar.gz "${BUNDLE_DIR}"

if [[ "${OSTYPE}" == "darwin"* ]]; then
  BLOB=$(base64 -b0 < bundle.tar.gz)
else
  BLOB=$(base64 -w0 < bundle.tar.gz)
fi

# Push application to repository
function push_csv() {
  echo "Push package ${OPERATOR_NAME} into namespace ${QUAY_ORG}"
  curl -H "Content-Type: application/json" \
      -H "Authorization: ${AUTH_TOKEN}" \
      -XPOST https://quay.io/cnr/api/v1/packages/"${QUAY_ORG}"/"${OPERATOR_NAME}" -d '
  {
      "blob": "'"${BLOB}"'",
      "release": "'"${VERSION}"'",
      "media_type": "helm"
  }'
}

# Delete application release in repository
function delete_csv() {
  echo "Delete release ${VERSION} of package ${OPERATOR_NAME} from namespace ${QUAY_ORG}"
  curl -H "Content-Type: application/json" \
      -H "Authorization: ${AUTH_TOKEN}" \
      -XDELETE https://quay.io/cnr/api/v1/packages/"${QUAY_ORG}"/"${OPERATOR_NAME}"/"${VERSION}"/helm
}

#-------------------------------------- Main --------------------------------------#
[[ "$DELETE_EXIST" == "no" ]] || delete_csv
push_csv
