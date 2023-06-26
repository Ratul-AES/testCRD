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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/Ratul-AES/testCRD/api/v1"
)

// CorePodReconciler reconciles a CorePod object
type CorePodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=corepods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=corepods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.dev.cloud,resources=corepods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CorePod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CorePodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered Core Pod Reconcile", "req", req)

	// Create an instance of the TestPvc controller
	testPvcController := &TestPvcReconciler{
		Client: r.Client, // Provide the same client instance
		// Set any other required fields for the TestPvc controller
	}

	// Create a reconcile request for the desired TestPvc resource
	testPvcReq := ctrl.Request{
		NamespacedName: client.ObjectKey{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
	}

	// Invoke the Reconcile method of the TestPvc controller
	_, err := testPvcController.Reconcile(ctx, testPvcReq)
	if err != nil {
		return ctrl.Result{}, err
	}
	l.Info("Successfully deploy DB")

	//Get OrgPod
	core := &webappv1.CorePod{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, core)

	l.Info("Got Core", "spec", core.Spec, "status", core.Status)
	if core.Name == "" {
		l.Info("[COREPOD] Rec entered after delete...")
		return ctrl.Result{}, nil
	}
	dbFinalizer := "webapp.aes.cloud/finalizercore"
	if core.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(core.GetFinalizers(), dbFinalizer) {
			controllerutil.AddFinalizer(core, dbFinalizer)
			if err := r.Update(ctx, core); err != nil {
				return ctrl.Result{}, err
			}
		}
		l.Info("COREPOD Finalizer not called")
	} else {
		// The object is being deleted
		if containsString(core.GetFinalizers(), dbFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, core, l); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(core, dbFinalizer)
			if err := r.Update(ctx, core); err != nil {
				return ctrl.Result{}, err
			}
			l.Info("COREPOD Finalizer used and removed...")
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
	if core.Name != core.Status.Name {
		core.Status.Name = core.Name
		r.Status().Update(ctx, core)
	}

	if core.Status.Progress != "Ready" { //Initializing statuses
		core.Status.Progress = "Initiating"
		core.Status.Ready = "0/2"
		r.Status().Update(ctx, core)
	}

	//Deploy Core
	orgDep, err := r.reconcileCoreOrg(ctx, core, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Deploy created", "deploy name", orgDep.Name, "Deploy Namespcae: ", orgDep.Namespace)

	//Deploy CoreService
	orgDepSvc, err := r.reconcileCoreOrgSvc(ctx, core, l)

	if err != nil {
		l.Info("Got the Error:==> " + err.Error())
		return ctrl.Result{}, err
	}
	l.Info("Deploy SVC created", "deploy SVC name", orgDepSvc.Name, "Deploy SVC Namespcae: ", orgDepSvc.Namespace)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CorePodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.CorePod{}).
		Owns(&webappv1.CorePodPermission{}).
		Owns(&webappv1.TestPvc{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
