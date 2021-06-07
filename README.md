# Tutorial for making a kubernetes-operator

## Setting up local dev environment with KIND

After making sure you have the latest kubectl, Docker and Go installed on your machine, and then you can simply install [Kind](https://github.com/kubernetes-sigs/kind) with `GO111MODULE="on" go get sigs.k8s.io/kind@v0.11.0` from your home directory. Make sure you have your kind directory in your PATH in order to run kind commands.

After you succesfully installed kind, fire up the cluster we use for local development:

```
kind create cluster --config kind-config.yaml

Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.21.1) 🖼
 ✓ Preparing nodes 📦 📦 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
 ✓ Joining worker nodes 🚜 
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind
```

## How to create an NGINX app with Ingress and HTTPS?

For this to work, you need a cluster in the clouds and a domain name you own. Once you created the cluster, copy the data necessary for kubectl to connect to it, put it in e.g. `kubeconfig.yaml` and export its location as an env with `export KUBECONFIG=~/path/to/kubeconfig.yaml`.

Following [this](https://www.fosstechnix.com/kubernetes-nginx-ingress-controller-letsencrypt-cert-managertls/) tutorial, the steps of `install.sh` will provide you a nginx web app that uses HTTPS.

First, you need to install nginx-ingress which also creates a LoadBalancer service (`kubectl get services --namespace ingress-nginx`). 

Next, you need to point Nginx Ingress Loadbalancer in domain name provider to access app using domain name.

The third step to install and configure the cert-manager.

Lastly, we need to create an nginx-app deployment and service with ingress.

## Building an operator with Kubebuilder

Following [this](https://www.youtube.com/watch?v=KBTXBUVNF2I) tutorial.

### Download kubebuilder and install locally.

```
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

### Init project

```
go mod init mywebapp
kubebuilder init --domain hellofromtheinternet.hu --repo hellofromtheinternet.hu/mywebapp
kubebuilder create api --group webapp --kind MyWebApp --version v0
```
