#!/usr/bin/env bash

set -xeuo pipefail

docker build -t gcr.io/covidtrace/notary .
docker push gcr.io/covidtrace/notary