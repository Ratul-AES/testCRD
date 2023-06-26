package controller

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	webappv1 "github.com/Ratul-AES/testCRD/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// For DB PVC
func (r *TestPvcReconciler) reconcilePVC(ctx context.Context, testPvc *webappv1.TestPvc, l logr.Logger) (corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: testPvc.Name + "-pvc", Namespace: testPvc.Namespace}, pvc)
	if err == nil {
		l.Info("PVC Found")
		return *pvc, nil
	}

	if !errors.IsNotFound(err) {
		return *pvc, err
	}
	//testPvc.Spec.Size = 1
	l.Info("PVC Not found, Creating new PVC")
	storageClass := "standard"
	storageRequest, _ := resource.ParseQuantity(fmt.Sprintf("%dGi", testPvc.Spec.Size))

	pvc = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testPvc.Namespace,
			Name:      testPvc.Name + "-pvc",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{"ReadWriteMany"},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"storage": storageRequest},
			},
		},
	}
	l.Info("Creating PVC")
	if err := ctrl.SetControllerReference(testPvc, pvc, r.Scheme); err != nil {
		return *pvc, err
	}
	return *pvc, r.Create(ctx, pvc)
}

// FOR DB Secret
const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (r *TestPvcReconciler) reconcileSecret(ctx context.Context, testsecret *webappv1.TestPvc, l logr.Logger) (corev1.Secret, error) {
	sec := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: testsecret.Name + "-sec", Namespace: testsecret.Namespace}, sec)
	if err == nil {
		l.Info("Secret Found")
		return *sec, nil
	}

	if !errors.IsNotFound(err) {
		return *sec, err
	}

	l.Info("Secret Not found, Creating new secret")
	sec = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testsecret.Name + "-sec",
			Namespace: testsecret.Namespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"password": []byte(StringWithCharset(10, charset)),
			"username": []byte("root"),
		},
	}
	l.Info("Creating Secret...", "Secret name", sec.Name, "Secret namespace", sec.Namespace)
	if err := ctrl.SetControllerReference(testsecret, sec, r.Scheme); err != nil {
		return *sec, err
	}

	return *sec, r.Create(ctx, sec)
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// DB CMAP
func (r *TestPvcReconciler) reconcileConfigMap(ctx context.Context, testconfig *webappv1.TestPvc, l logr.Logger) (corev1.ConfigMap, error) {
	cmap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: testconfig.Name + "-cmap", Namespace: testconfig.Namespace}, cmap)
	if err == nil {
		l.Info("SVC Found")
		return *cmap, nil
	}

	if !errors.IsNotFound(err) {
		return *cmap, err
	}

	labels := map[string]string{
		"app": "mysql",
	}
	cmap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testconfig.Name + "-cmap",
			Namespace: testconfig.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			//"host":     testconfig.Name + "-svc",
			"host":     "root",
			"db_name":  "demo_proj",
			"init.sql": "CREATE DATABASE IF NOT EXISTS demo_proj; USE demo_proj;",
		},
	}
	l.Info("Creating OrgNodePort...", "NodePort name", cmap.Name, "NodePort namespace", cmap.Namespace)
	if err := ctrl.SetControllerReference(testconfig, cmap, r.Scheme); err != nil {
		return *cmap, err
	}

	return *cmap, r.Create(ctx, cmap)
}

func (r *TestPvcReconciler) reconcileDBDepl(ctx context.Context, testDB *webappv1.TestPvc, l logr.Logger) (appsv1.Deployment, error) {

	mysql := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: testDB.Name + "-db", Namespace: testDB.Namespace}, mysql)
	if err == nil {
		l.Info("DEPL Found")
		return *mysql, nil
	}

	if !errors.IsNotFound(err) {
		return *mysql, err
	}
	l.Info("DEPL Not found, Creating new DEPL")

	//Deployment definition here

	labels := map[string]string{
		"app": testDB.Name + "-dbsvc",
	}

	//testDB.Spec.DbImg = "mysql:5.6"

	mysql = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testDB.Name + "-db",
			Namespace: testDB.Namespace,
			Labels:    labels,
		},

		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: testDB.Spec.DbImg,
						Name:  "mysql",

						Env: []corev1.EnvVar{
							{
								Name: "MYSQL_ROOT_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: testDB.Name + "-sec",
										},
										Key: "password",
									},
								},
							},
							{
								Name:  "MYSQL_BIND_ADDRESS",
								Value: "0.0.0.0", // Change this to the desired bind address
							},
						},

						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mysql",
						}},
						VolumeMounts: []corev1.VolumeMount{ //add volume mount for /docker-entrypoint
							{
								Name:      "mysql-persistent-storage",
								MountPath: "/var/lib/mysql",
							},
							{
								Name:      "mysql-initdb",
								MountPath: "/docker-entrypoint-initdb.d",
							},
						},
					},
					},

					Volumes: []corev1.Volume{ //refer to above entrypoint and connect configMap defined below
						{
							Name: "mysql-persistent-storage",
							VolumeSource: corev1.VolumeSource{

								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: testDB.Name + "-pvc",
								},
							},
						},
						{
							Name: "mysql-initdb",
							VolumeSource: corev1.VolumeSource{

								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: testDB.Name + "-cmap",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	l.Info("Creating Deployment...", "DEPL name", mysql.Name, "DEPL namespace", mysql.Namespace)
	if err := ctrl.SetControllerReference(testDB, mysql, r.Scheme); err != nil {
		return *mysql, err
	}

	return *mysql, r.Create(ctx, mysql)
}

func (r *TestPvcReconciler) reconcileService(ctx context.Context, testDBService *webappv1.TestPvc, l logr.Logger) (corev1.Service, error) {

	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: testDBService.Name + "-svc", Namespace: testDBService.Namespace}, svc)
	if err == nil {
		l.Info("SVC Found")
		return *svc, nil
	}

	if !errors.IsNotFound(err) {
		return *svc, err
	}

	l.Info("SVC Not found, Creating new SVC")

	svc = &corev1.Service{

		ObjectMeta: metav1.ObjectMeta{
			Name:      testDBService.Name + "-svc",
			Namespace: testDBService.Namespace,
		},

		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": testDBService.Name + "-dbsvc"},

			Ports: []corev1.ServicePort{
				{
					Port: 3306,
					Name: "mysql",
				},
			},
		},
	}
	l.Info("Creating SVC...", "SVC name", svc.Name, "SVC namespace", svc.Namespace)
	if err := ctrl.SetControllerReference(testDBService, svc, r.Scheme); err != nil {
		return *svc, err
	}

	return *svc, r.Create(ctx, svc)
}
