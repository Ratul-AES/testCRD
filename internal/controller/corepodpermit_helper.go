package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	webappv1 "github.com/Ratul-AES/testCRD/api/v1"
	"github.com/go-logr/logr"
)

func (r *CorePodPermissionReconciler) reconcileRole(ctx context.Context, corepodRole *webappv1.CorePodPermission, l logr.Logger) (v1.Role, error) {
	role := &v1.Role{}
	err := r.Get(ctx, types.NamespacedName{Name: corepodRole.Name + "-role", Namespace: corepodRole.Namespace}, role)
	if err == nil {
		l.Info("SVC Found")
		return *role, nil
	}

	if !errors.IsNotFound(err) {
		return *role, err
	}

	l.Info("Role Not found, Creating new role")

	role = &v1.Role{

		ObjectMeta: metav1.ObjectMeta{
			Name:      corepodRole.Name + "-role",
			Namespace: corepodRole.Namespace,
		},

		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
	l.Info("Creating Role...", "Role NodePort name", role.Name, "Role namespace", role.Namespace)
	if err := ctrl.SetControllerReference(corepodRole, role, r.Scheme); err != nil {
		return *role, err
	}

	return *role, r.Create(ctx, role)
}

func (r *CorePodPermissionReconciler) reconcileClusterRole(ctx context.Context, corepodCR *webappv1.CorePodPermission, l logr.Logger) (v1.ClusterRole, error) {
	role := &v1.ClusterRole{}
	err := r.Get(ctx, types.NamespacedName{Name: corepodCR.Name + "-crole", Namespace: corepodCR.Namespace}, role)
	if err == nil {
		l.Info("cRole Found")
		return *role, nil
	}

	if !errors.IsNotFound(err) {
		return *role, err
	}

	l.Info("cRole Not found, Creating new crole")

	role = &v1.ClusterRole{

		ObjectMeta: metav1.ObjectMeta{
			Name: corepodCR.Name + "-crole",
		},

		Rules: []v1.PolicyRule{
			{
				APIGroups: []string{"webapp.dev.cloud"},
				Resources: []string{"corepods"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"webapp.dev.cloud"},
				Resources: []string{"testpvcs"},
				Verbs:     []string{"*"},
			},
		},
	}
	l.Info("Creating Cluster Role...", "CR name", role.Name)

	return *role, r.Create(ctx, role)
}

func (r *CorePodPermissionReconciler) reconcileSA(ctx context.Context, corepod *webappv1.CorePodPermission, l logr.Logger) (corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	err := r.Get(ctx, types.NamespacedName{Name: corepod.Name + "-sa", Namespace: corepod.Namespace}, sa)
	if err == nil {
		l.Info("SA Found")
		return *sa, nil
	}

	if !errors.IsNotFound(err) {
		return *sa, err
	}

	l.Info("SA Not found, Creating new SA")

	sa = &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      corepod.Name + "-sa",
			Namespace: corepod.Namespace,
		},
	}
	l.Info("Creating SA...", "SA name", sa.Name)
	if err := ctrl.SetControllerReference(corepod, sa, r.Scheme); err != nil {
		return *sa, err
	}
	return *sa, r.Create(ctx, sa)
}

func (r *CorePodPermissionReconciler) reconcileRB(ctx context.Context, corepod *webappv1.CorePodPermission, l logr.Logger) (v1.RoleBinding, error) {
	rb := &v1.RoleBinding{}
	err := r.Get(ctx, types.NamespacedName{Name: corepod.Name + "-rb", Namespace: corepod.Namespace}, rb)
	if err == nil {
		l.Info("RoleBinding Found")
		return *rb, nil
	}

	if !errors.IsNotFound(err) {
		return *rb, err
	}

	l.Info("RB Not found, Creating new RoleBinding")
	rb = &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      corepod.Name + "-rb",
			Namespace: corepod.Namespace,
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     corepod.Name + "-role",
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      corepod.Name + "-sa",
				Namespace: corepod.Namespace,
			},
		},
	}
	l.Info("Creating RB...", "RB name", rb.Name)
	if err := ctrl.SetControllerReference(corepod, rb, r.Scheme); err != nil {
		return *rb, err
	}
	return *rb, r.Create(ctx, rb)
}

func (r *CorePodPermissionReconciler) reconcileCRB(ctx context.Context, corepod *webappv1.CorePodPermission, l logr.Logger) (v1.ClusterRoleBinding, error) {
	rb := &v1.ClusterRoleBinding{}
	err := r.Get(ctx, types.NamespacedName{Name: corepod.Name + "-crb", Namespace: corepod.Namespace}, rb)
	if err == nil {
		l.Info("ClusterRoleBinding Found")
		return *rb, nil
	}

	if !errors.IsNotFound(err) {
		return *rb, err
	}

	l.Info("RB Not found, Creating new RoleBinding")
	rb = &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: corepod.Name + "-crb",
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     corepod.Name + "-crole",
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      corepod.Name + "-sa",
				Namespace: corepod.Namespace,
			},
		},
	}
	l.Info("Creating CRB...", "CRB name", rb.Name)

	return *rb, r.Create(ctx, rb)
}
