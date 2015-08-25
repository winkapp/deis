#!/bin/bash
################################################################################
# This is a temporary script for installing Deis into an existing
# Kubernetes cluster.
################################################################################

echo "Creating namespace"
kubectl create -f units/namespaces/deis-namespace.json

echo "Creating a kubectl context named 'deis-dev'"
kubectl config set-context deis-dev --namespace=deis --cluster=vagrant --user=vagrant
kubectl config use-context deis-dev

echo "Loading services"
for s in units/services/*.json; do
  kubectl create -f "$s"
done

echo "Loading RCs"
for s in units/rcs/*-store-*.json; do
  kubectl create -f "$s"
done

for s in units/rcs/*.json; do
  kubectl create -f "$s"
done
