#!/bin/bash
set -e

APPENV=${APPENV:-bezospherenv}

/opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$APPENV > /$APPENV

source /$APPENV && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$BEZOSPHERE_CERT > /$BEZOSPHERE_CERT && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$BEZOSPHERE_CERT_KEY > /$BEZOSPHERE_CERT_KEY && \
  chmod 600 /$BEZOSPHERE_CERT_KEY && \
  /opt/bin/migrate -url "$BEZOSPHERE_POSTGRES_CONN" -path /migrations up && \
  /bezosphere
