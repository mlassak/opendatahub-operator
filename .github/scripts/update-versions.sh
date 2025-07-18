#!/bin/bash

set -euo

NEW_VERSION=$1

CURRENT_VERSION=$(cat Makefile | grep -w "VERSION ?=" | cut -d ' ' -f 3)
CSV_FILE=config/manifests/bases/opendatahub-operator.clusterserviceversion.yaml
sed -i -e "s/^VERSION ?=.*/VERSION ?= $NEW_VERSION/g" Makefile
sed -i -e "s|containerImage.*|containerImage: quay.io/opendatahub/opendatahub-operator:v$NEW_VERSION|g" $CSV_FILE
sed -i -e "s|createdAt.*|createdAt: \"$(date +"%Y-%-m-%dT00:00:00Z")\"|g" $CSV_FILE
sed -i -e "s|name: opendatahub-operator.v.*|name: opendatahub-operator.v$NEW_VERSION|g" $CSV_FILE
sed -i -e "s|version: $CURRENT_VERSION.*|version: $NEW_VERSION|g" $CSV_FILE
