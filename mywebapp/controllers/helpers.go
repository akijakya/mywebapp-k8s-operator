package controllers

import (
	cmacme "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	webappv0 "hellofromtheinternet.hu/mywebapp/api/v0"
)

const (
	certName = "letsencrypt-prod"
)

func (r *MyWebappReconciler) desiredIssuer(webapp webappv0.MyWebapp) (certmanager.ClusterIssuer, error) {
	ingressClass := "nginx"
	issuer := certmanager.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{APIVersion: certmanager.SchemeGroupVersion.String(), Kind: "ClusterIssuer"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: webapp.Namespace,
		},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{
				ACME: &cmacme.ACMEIssuer{
					Server: "https://acme-v02.api.letsencrypt.org/directory",
					Email:  webapp.Spec.Email,
					PrivateKey: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{
							Name: certName,
						},
					},
					Solvers: []cmacme.ACMEChallengeSolver{
						{
							HTTP01: &cmacme.ACMEChallengeSolverHTTP01{
								Ingress: &cmacme.ACMEChallengeSolverHTTP01Ingress{
									Class: &ingressClass,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &issuer, r.Scheme); err != nil {
		return issuer, err
	}

	return issuer, nil
}

func (r *MyWebappReconciler) desiredCertificate(webapp webappv0.MyWebapp) (certmanager.Certificate, error) {
	cert := certmanager.Certificate{
		TypeMeta: metav1.TypeMeta{APIVersion: certmanager.SchemeGroupVersion.String(), Kind: "Certificate"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Spec.Host,
			Namespace: webapp.Namespace,
		},
		Spec: certmanager.CertificateSpec{
			SecretName: webapp.Spec.Host + "-tls",
			IssuerRef: cmmeta.ObjectReference{
				Name: certName,
				Kind: "ClusterIssuer",
			},
			CommonName: webapp.Spec.Host,
			DNSNames:   []string{webapp.Spec.Host},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &cert, r.Scheme); err != nil {
		return cert, err
	}

	return cert, nil
}

func (r *MyWebappReconciler) desiredDeployment(webapp webappv0.MyWebapp) (appsv1.Deployment, error) {
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
			Labels:    map[string]string{"webapp": webapp.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: webapp.Spec.Replicas, // won't be nil because defaulting
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"webapp": webapp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"webapp": webapp.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mywebapp",
							Image: webapp.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 80, Name: "http", Protocol: "TCP"},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &depl, r.Scheme); err != nil {
		return depl, err
	}

	return depl, nil
}

func (r *MyWebappReconciler) desiredService(webapp webappv0.MyWebapp) (corev1.Service, error) {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"webapp": webapp.Name},
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: "TCP", TargetPort: intstr.FromString("http")},
			},
			// Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	// always set the controller reference so that we know which object owns this.
	if err := ctrl.SetControllerReference(&webapp, &svc, r.Scheme); err != nil {
		return svc, err
	}

	return svc, nil
}

func (r *MyWebappReconciler) desiredIngress(webapp webappv0.MyWebapp) (networkv1.Ingress, error) {
	pathType := networkv1.PathTypePrefix
	ingress := networkv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: networkv1.SchemeGroupVersion.String(), Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name + "-ingress",
			Namespace: webapp.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":    "nginx",
				"cert-manager.io/cluster-issuer": "letsencrypt-prod",
			},
		},
		Spec: networkv1.IngressSpec{
			TLS: []networkv1.IngressTLS{
				{Hosts: []string{webapp.Spec.Host}, SecretName: "letsencrypt-prod"},
			},
			Rules: []networkv1.IngressRule{
				{
					Host: webapp.Spec.Host,
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									PathType: &pathType,
									Path:     "/",
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: webapp.Name,
											Port: networkv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// always set the controller reference so that we know which object owns this.
	if err := ctrl.SetControllerReference(&webapp, &ingress, r.Scheme); err != nil {
		return ingress, err
	}

	return ingress, nil
}
