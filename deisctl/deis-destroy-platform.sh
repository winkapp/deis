#!/bin/bash

echo "Switching to context deis-dev"
kubectl config use-context deis-dev

for s in units/services/*.json units/rcs/*.json units/namespaces/deis-namespace.json; do
  kubectl delete -f "$s"
done
