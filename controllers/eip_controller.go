/*
Copyright 2021 Dan Finneran.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/packethost/packngo"
	ketalv1 "github.com/thebsdbox/equinix-ketal/api/v1"
)

// EipReconciler reconciles a Eip object
type EipReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// Our Equinix Metal API client
	emClient *packngo.Client
}

//+kubebuilder:rbac:groups=ketal.equinix.metal,resources=eips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ketal.equinix.metal,resources=eips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ketal.equinix.metal,resources=eips/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Eip object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *EipReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var eip ketalv1.Eip
	if err := r.Get(ctx, req.NamespacedName, &eip); err != nil {
		log.Error(err, "unable to fetch Elastic IP")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If this is Zero the object has been deleted
	if !eip.DeletionTimestamp.IsZero() {
		log.Info("Deleted")
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EipReconciler) SetupWithManager(mgr ctrl.Manager, emClient *packngo.Client) error {
	r.emClient = emClient
	return ctrl.NewControllerManagedBy(mgr).
		For(&ketalv1.Eip{}).
		Complete(r)
}
