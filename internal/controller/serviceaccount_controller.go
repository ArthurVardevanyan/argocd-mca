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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	authenticationV1 "k8s.io/api/authentication/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	argocdv1beta1 "github.com/ArthurVardevanyan/argocd-mca/api/v1beta1"
)

// ServiceAccountReconciler reconciles a ServiceAccount object
type ServiceAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func kubernetesAuthToken(expirationSeconds int) *authenticationV1.TokenRequest {
	ExpirationSeconds := int64(expirationSeconds)

	tokenRequest := &authenticationV1.TokenRequest{
		Spec: authenticationV1.TokenRequestSpec{
			Audiences:         []string{"openshift"},
			ExpirationSeconds: &ExpirationSeconds,
		},
	}

	return tokenRequest
}

//+kubebuilder:rbac:groups=argocd.arthurvardevanyan.com,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argocd.arthurvardevanyan.com,resources=serviceaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=argocd.arthurvardevanyan.com,resources=serviceaccounts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *ServiceAccountReconciler) Reconcile(reconcilerContext context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(reconcilerContext)
	log.V(1).Info(req.Name)

	// Common Variables
	var err error
	var error string

	var namespace = "argocd-mca-sa"

	var serviceAccount argocdv1beta1.ServiceAccount
	if err = r.Get(reconcilerContext, req.NamespacedName, &serviceAccount); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.V(1).Info("Artifact Registry Auth Object Not Found or No Longer Exists!")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "Unable to fetch Artifact Registry Auth Object")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	//Reset Error
	// serviceAccount.Status.Error = ""

	tenantServiceAccount := &coreV1.ServiceAccount{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      serviceAccount.Namespace,
			Namespace: namespace,
		},
	}
	err = r.Update(reconcilerContext, tenantServiceAccount)
	if err != nil {
		err = r.Create(reconcilerContext, tenantServiceAccount)
		if err != nil {
			error = "Unable to Create Service Account"
			log.Error(err, error)
		}
	}

	// Generate k8s Auth Token
	const expirationSeconds = 3600
	k8sAuthToken := kubernetesAuthToken(expirationSeconds)
	var retrievedServiceAccount coreV1.ServiceAccount

	err = r.Get(reconcilerContext, client.ObjectKey{Name: serviceAccount.Namespace, Namespace: namespace}, &retrievedServiceAccount)
	if err != nil {
		error = "Service Account Not Found"
		log.Error(err, error)
	}

	err = r.SubResource("token").Create(reconcilerContext, &retrievedServiceAccount, k8sAuthToken)
	if err != nil {
		error = "Unable to Create Kubernetes Token"
		log.Error(err, error)
	}

	log.Info(k8sAuthToken.Status.Token)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&argocdv1beta1.ServiceAccount{}).
		Complete(r)
}
