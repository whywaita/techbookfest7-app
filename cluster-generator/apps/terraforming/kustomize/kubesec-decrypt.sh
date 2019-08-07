#!/bin/bash

file=./base/secret.enc.yaml

output=$(echo $file | sed -e 's/\.enc//')
echo $output
kubesec decrypt $file > $output
