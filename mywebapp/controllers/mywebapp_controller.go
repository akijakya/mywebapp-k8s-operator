/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv0 "hellofromtheinternet.hu/mywebapp/api/v0"
)

// MyWebappReconciler reconciles a MyWebapp object
type MyWebappReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.hellofromtheinternet.hu,resources=mywebapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.hellofromtheinternet.hu,resources=mywebapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.hellofromtheinternet.hu,resources=mywebapps/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;get;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=list;watch;get;patch

func (r *MyWebappReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("mywebapp", req.NamespacedName)
	log.Info("reconciling mywebapp")

	var webapp webappv0.MyWebapp
	if err := r.Get(ctx, req.NamespacedName, &webapp); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	issuer, err := r.desiredIssuer(webapp)
	if err != nil {
		return ctrl.Result{}, err
	}

	certificate, err := r.desiredCertificate(webapp)
	if err != nil {
		return ctrl.Result{}, err
	}

	deployment, err := r.desiredDeployment(webapp)
	if err != nil {
		return ctrl.Result{}, err
	}

	service, err := r.desiredService(webapp)
	if err != nil {
		return ctrl.Result{}, err
	}

	ingress, err := r.desiredIngress(webapp)
	if err != nil {
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("mywebapp-controller")}

	err = r.Patch(ctx, &issuer, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &certificate, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &service, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &ingress, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciled mywebapp")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyWebappReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv0.MyWebapp{}).
		Owns(&certmanager.ClusterIssuer{}).
		Owns(&certmanager.Certificate{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Owns(&networkv1.Ingress{}).
		Complete(r)
}
