#!/bin/bash

set -euxo pipefail

openssl genrsa -out internal/crypto/private.pem 2048
openssl rsa -in internal/crypto/private.pem -out internal/crypto/public.pem -pubout -outform PEM
openssl rand 32 >internal/crypto/hmac_secret
