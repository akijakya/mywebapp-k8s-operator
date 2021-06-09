# mywebapp-k8s-operator

Practice kubernetes operator that creates an nxinx app with HTTPS

## Prerequisites

1) Docker, kubectl installed
2) Access to a kubernetes cluster via kubectl

## Installation

Mywebapp-k8s-operator needs the ingress-nginx and cert-manager to be installed on your cluster, for which you can simply run:

```
make install-pre
```

After that, the installation of the `mywebapp` custom resource can be done with the following command:

```
make deploy
```

To start your webapp, create a manifest similar to this (using your domain and email address):

### mywebapp.yaml
```
apiVersion: webapp.hellofromtheinternet.hu/v0
kind: MyWebapp
metadata:
  name: mywebapp-sample
spec:
  replicas: 4
  host: yourdomain.com
  email: email@yourdomain.com
  image: nginx:1.20.1
```

And create it in your cluster with 

```
kubectl create -f mywebapp.yaml
```
