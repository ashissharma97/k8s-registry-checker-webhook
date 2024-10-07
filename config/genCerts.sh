#!/bin/bash

openssl genrsa -out ca.key 2048

openssl req -new -x509 -days 365 -key ca.key \
  -subj "/C=AU/CN=registry-checker-webhook" \
  -out ca.crt

openssl req -newkey rsa:2048 -nodes -keyout server.key \
  -subj "/C=AU/CN=registry-checker-webhook" \
  -out server.csr

openssl x509 -req \
  -extfile <(printf "subjectAltName=DNS:registry-checker-webhook.default.svc") \
  -days 365 \
  -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt

echo
echo ">> Generating kube secrets..."
kubectl create secret tls registry-checker-webhook-tls \
  --cert=server.crt \
  --key=server.key \
  --dry-run=client -o yaml \
  >../manifests/tls-secret.yaml

CRT=$(cat ../tls/ca.crt | base64)

sed -i "" "/^\([[:space:]]*caBundle: \).*/s//\1$CRT/" ../manifest/webhook.yaml

rm ca.crt ca.key ca.srl server.crt server.csr server.key
