/*
Copyright 2023.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/Ratul-AES/testCRD/api/v1"
)

// TestPvcReconciler reconciles a TestPvc object
type TestPvcReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=testpvcs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=testpvcs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=testpvcs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TestPvc object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *TestPvcReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered PVC Reconcile", "req", req)

	//Get TestPVC
	testPvc := &webappv1.TestPvc{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, testPvc)

	l.Info("Got dbpod", "spec", testPvc.Spec, "status", testPvc.Status)
	if testPvc.Name == "" {
		l.Info("[PVC] Rec entered after delete...")
		return ctrl.Result{}, nil
	}

	if testPvc.Name != testPvc.Status.Name {
		testPvc.Status.Name = testPvc.Name
		r.Status().Update(ctx, testPvc)
	}

	if testPvc.Status.Progress != "Ready" { //Initializing statuses
		testPvc.Status.Progress = "Initiating"
		testPvc.Status.Ready = "0/1"
		r.Status().Update(ctx, testPvc)
	}

	////////////////// DB PVC
	pvc, err := r.reconcilePVC(ctx, testPvc, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("PVC Created :)", "PVCName: ", pvc.Name, "PVCNamespace: ", pvc.Namespace, "PVC Size: ", pvc.Size())

	/////////////////// DB Secret
	dbsec, err := r.reconcileSecret(ctx, testPvc, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Secret Created :)", "SecretName", dbsec.Name, "SecretNamespace", dbsec.Namespace)

	////////////////// DB ConfigMap
	configMap, err := r.reconcileConfigMap(ctx, testPvc, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("ConfigMap Created :)", "CMCName", configMap.Name, "CMNamespace", configMap.Namespace)

	/////////////////// DB DEPL
	mysql, err := r.reconcileDBDepl(ctx, testPvc, l)

	if err != nil {
		return ctrl.Result{}, err
	}
	l.Info("dbpod Created :)", "DB NAME", mysql.Name, "DBNamespace", mysql.Namespace) //after this, for some reason next lines are not being reached

	serv, err := r.reconcileService(ctx, testPvc, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("DB Svc Created :)", "SvcName", serv.Name, "SvcNamespace", serv.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TestPvcReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.TestPvc{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
