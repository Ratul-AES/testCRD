package controller

import (
	"context"

	webappv1 "github.com/Ratul-AES/testCRD/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *CorePodReconciler) reconcileCoreOrgSvc(ctx context.Context, corePodSvc *webappv1.CorePod, l logr.Logger) (webappv1.CorePod, error) {
	l.Info("Enter SVC")

	// corePodSvcDep := &corev1.Service{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      corePodSvc.Name + "-svc",
	// 		Namespace: corePodSvc.Namespace,
	// 	},

	// 	Spec: corev1.ServiceSpec{
	// 		Selector: map[string]string{
	// 			"app": corePodSvc.Name,
	// 		},
	// 		Type: corev1.ServiceTypeLoadBalancer,
	// 		Ports: []corev1.ServicePort{
	// 			{
	// 				Protocol:   corev1.ProtocolTCP,
	// 				Port:       8080,
	// 				TargetPort: intstr.FromInt(8080),
	// 				NodePort:   30000,
	// 			},
	// 		},
	// 	},
	// }

	corePodSvcDep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      corePodSvc.Name + "-svc",
			Namespace: corePodSvc.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": corePodSvc.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					NodePort:   30000,
				},
			},
		},
	}
	l.Info("Creating Deployment SVC...", "DEPL SVC name", corePodSvcDep.Name, "DEPL SVC namespace", corePodSvcDep.Namespace)
	// if err := ctrl.SetControllerReference(orgpod, corg, r.Scheme); err != nil {
	// 	return *corg, err
	// }

	return *corePodSvc, r.Create(ctx, corePodSvc)

}

func (r *CorePodReconciler) reconcileCoreOrg(ctx context.Context, corePod *webappv1.CorePod, l logr.Logger) (webappv1.CorePod, error) {

	labels := map[string]string{
		"app": corePod.Name,
	}
	replicaCount := int32(corePod.Spec.Size)

	corePodDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      corePod.Name + "-org",
			Namespace: corePod.Namespace,
			Labels:    labels,
		},

		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicaCount,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: corePod.Spec.OrgImg,
						Name:  "organization",

						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},

						Env: []corev1.EnvVar{
							{
								Name: "DB_PASS",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: corePod.Name + "-sec",
										},
										Key: "password",
									},
								},
							},
							{
								Name: "DB_USER",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: corePod.Name + "-sec",
										},
										Key: "username",
									},
								},
							},
							{
								Name: "DB_HOST",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: corePod.Name + "-cmap",
										},
										Key: "host",
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
	l.Info("Creating Deployment...", "DEPL name", corePodDep.Name, "DEPL namespace", corePodDep.Namespace)
	// if err := ctrl.SetControllerReference(orgpod, corg, r.Scheme); err != nil {
	// 	return *corg, err
	// }

	return *corePod, r.Create(ctx, corePodDep)

}
