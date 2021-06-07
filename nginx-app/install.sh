#!/bin/bash

echo 'Install Nginx Ingress Controller'
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.46.0/deploy/static/provider/cloud/deploy.yaml

echo 'Configure cert manager for Nginx Ingress'
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.yaml
kubectl apply -f letsencrypt-issuer.yaml
kubectl apply -f letsencrypt-cert.yaml

echo 'Creating nginx-app deployment and service with ingress'
kubectl apply -f nginx-app.yaml