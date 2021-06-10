# Tutorial for making a kubernetes operator

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

## How to create an NGINX app with Ingress and HTTPS (manually)?

For this to work, you need a cluster in the clouds and a domain name you own. Once you created the cluster, copy the data necessary for kubectl to connect to it, put it in e.g. `kubeconfig.yaml` and export its location as an env with `export KUBECONFIG=~/path/to/kubeconfig.yaml`.

Following [this](https://www.fosstechnix.com/kubernetes-nginx-ingress-controller-letsencrypt-cert-managertls/) tutorial, the steps to create a nginx web app that uses HTTPS:

1) Install nginx-ingress (`kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.46.0/deploy/static/provider/cloud/deploy.yaml`) which also creates a LoadBalancer service (`kubectl get services --namespace ingress-nginx`). 

2) Point the address of the loadbalancer to your domain in the settings of your provider.

3) Install (`kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.yaml`) and configure the cert-manager, by creating these manifests and then apply them to your cluster:

### letsencrypt-issuer.yaml
```
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
  namespace: default
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: your@email.com
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod
    # Enable the HTTP-01 challenge provider
    solvers:
    - http01:
        ingress:
          class: nginx
```

### letsencrypt-cert.yaml
```
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: hellofromtheinternet.hu
  namespace: default
spec:
  secretName: hellofromtheinternet.hu-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  commonName: hellofromtheinternet.hu
  dnsNames:
  - hellofromtheinternet.hu
```

```
kubectl apply -f letsencrypt-issuer.yaml
kubectl apply -f letsencrypt-cert.yaml
```

4) Create an nginx-app deployment and service with ingressm and apply them as well:

### nginx-app.yaml
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-app
  namespace: default
  labels:
    app: nginx-app
spec:
  replicas: 4
  selector:
    matchLabels:
      app: nginx-app
  template:
    metadata:
      labels:
        app: nginx-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.20.1
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-app
  namespace: default
spec:
  selector:
    app: nginx-app
  ports:
  - name: http
    targetPort: 80
    port: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - hellofromtheinternet.hu
    secretName: hellofromtheinternet.hu-tls
  rules:
  - host: hellofromtheinternet.hu
    http:
      paths:
      - pathType: Prefix
        path: /
        backend:
          service:
             name: nginx-app
             port:
               number: 80
```

```
kubectl apply -f nginx-app.yaml
```

## Building an operator with Kubebuilder

Following [this](https://www.youtube.com/watch?v=KBTXBUVNF2I) tutorial.
Kubebuilder docs: https://book.kubebuilder.io/quick-start.html#installation

### Download kubebuilder and install locally.

```
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

### Init project

```
go mod init mywebapp
kubebuilder init --domain example.com --repo example.com/mywebapp
kubebuilder create api --group webapp --kind MyWebapp --version v0
```

Make changes in `/mywebapp/api/v0/mywebapp_types.go`, providing the options needs to be set for the operator.

`make manifests` will create the manifest to install this new CRD to your cluster, which you will find here: `mywebapp/config/crd/bases/webapp.hellofromtheinternet.hu_mywebapps.yaml`

Install it with `kubectl create -f config/crd/bases`

Now you can modify the sample yaml to create one resource: `config/samples/webapp_v0_mywebapp.yaml`

You can apply that as well with `kubectl create -f config/samples/webapp_v0_mywebapp.yaml`. If you made a mistake there, like a wrong type of value as an option, the validation will work.

After creating it successfully you can get it right away:

```
kubectl get mywebapps

NAME              AGE
mywebapp-sample   10s
```

To have more fields than NAME and AGE, specify more columns in `api/v0/mywebapp_types.go` with adding lines like this one:

```
// +kubebuilder:printcolumn:JSONPath=".spec.host",name="URL",type="string"
```

Now its time to write the reconsiliation of the manifests we want to apply: `controllers/mywebapp_controller.go` and `controllers/helpers.go`.

To check that it would work properly: `make run`

To create an image, push that to a repository (if you use dockerhub, login first with `docker login`) and deploy:

```
export IMG=akijakya/mywebapp-k8s-operator:v0
make docker-build docker-push deploy
```
