#!/bin/bash

KMS_KEY="gcp:projects/sample-project-name/locations/global/keyRings/infra-dev-1-kubesec/cryptoKeys/kubesec-key"

file=./base/secret.yaml

output=$(echo $file | sed -e 's/\.yaml/\.enc\.yaml/')
echo $output
kubesec encrypt --key=$KMS_KEY $file > $output
