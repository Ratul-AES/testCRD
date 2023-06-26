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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/Ratul-AES/testCRD/api/v1"
)

// CorePodPermissionReconciler reconciles a CorePodPermission object
type CorePodPermissionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=corepodpermissions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=corepodpermissions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=corepodpermissions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CorePodPermission object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CorePodPermissionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered CoreEXT Reconcile", "req", req)

	//Get ext
	ext := &webappv1.CorePodPermission{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, ext)

	l.Info("Got EXT", "spec", ext.Spec, "status", ext.Status)

	if ext.Name == "" {
		l.Info("[COREEXT] Rec entered after delete...")
		return ctrl.Result{}, nil
	}

	//Permissions Resources

	sa, err := r.reconcileSA(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Service Account Created", "SA Name", sa.Name, "SA Namespace", sa.Namespace)

	role, err := r.reconcileRole(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Role Created", "Role Name", role.Name, "Role Namespace", role.Namespace)

	rb, err := r.reconcileRB(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Rolebinding Created", "RB Name", rb.Name, "RB Namespace", rb.Namespace)

	// TODO(user): your logic here
	cluster_role, err := r.reconcileClusterRole(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Cluster Role Created", "Cluster Role Name", cluster_role.Name, "Cluster Role Namespace", cluster_role.Namespace)

	crb, err := r.reconcileCRB(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("ClusterR Rolebinding Created", "CRB Name", crb.Name, "CRB Namespace", crb.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CorePodPermissionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.CorePodPermission{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&v1.Role{}).
		Owns(&v1.RoleBinding{}).
		//Owns(&v1.ClusterRole{}).
		//Owns(&v1.ClusterRoleBinding{}).
		Complete(r)
}
