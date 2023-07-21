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
	"encoding/json"
	"os"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coreV1 "k8s.io/api/core/v1"

	argocdv1beta1 "github.com/ArthurVardevanyan/argocd-mca/api/v1beta1"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type Config struct {
	BearerToken     string `json:"bearerToken"`
	TlsClientConfig TlsClientConfig
}

type TlsClientConfig struct {
	Insecure bool `json:"insecure"`
}

//+kubebuilder:rbac:groups=argocd.arthurvardevanyan.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argocd.arthurvardevanyan.com,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=argocd.arthurvardevanyan.com,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(reconcilerContext context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(reconcilerContext)
	log.V(1).Info(req.Name)

	// Common Variables
	var err error
	// var error string

	var cluster argocdv1beta1.Cluster
	if err = r.Get(reconcilerContext, req.NamespacedName, &cluster); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.V(1).Info("Artifact Registry Auth Object Not Found or No Longer Exists!")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "Unable to fetch Artifact Registry Auth Object")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	// GCP project in which to store secrets in Secret Manager.
	projectID := os.Getenv("GCP_PROJECT_ID")

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		println(err.Error())

	}
	defer client.Close()

	// Create the request to create the secret.

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/" + projectID + "/secrets/" + cluster.Namespace + "/versions/latest",
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		println(err.Error())
	}

	var clusterSecretConfig Config
	clusterSecretConfig.TlsClientConfig.Insecure = false
	clusterSecretConfig.BearerToken = string(result.Payload.Data)

	jsonString, err := json.Marshal(clusterSecretConfig)
	if err != nil {
		println(err.Error())
	}

	m := make(map[string]string)

	m["name"] = cluster.Namespace
	m["server"] = cluster.Spec.Server
	m["namespaces"] = cluster.Spec.Namespaces
	m["config"] = string(jsonString)

	secret := &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:        cluster.Namespace,
			Namespace:   cluster.Namespace,
			Annotations: map[string]string{"managed-by": "argocd.argoproj.io"},
			Labels:      map[string]string{"argocd.argoproj.io/secret-type": "cluster"},
		},
		Type:       "Opaque",
		StringData: m,
	}

	err = r.Update(reconcilerContext, secret)
	if err != nil {
		err = r.Create(reconcilerContext, secret)
		if err != nil {
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&argocdv1beta1.Cluster{}).
		Complete(r)
}
