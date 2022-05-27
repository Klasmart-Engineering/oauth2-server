#!/bin/bash

set -euxo pipefail

openssl genrsa -out internal/keys/private.pem 2048
openssl rsa -in internal/keys/private.pem -out internal/keys/public.pem -pubout -outform PEM
openssl rand 32 >internal/keys/hmac_secret
