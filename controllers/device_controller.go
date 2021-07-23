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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"github.com/packethost/packngo"
	ketalv1 "github.com/thebsdbox/equinix-ketal/api/v1"
)

// DeviceReconciler reconciles a Device object
type DeviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Our Equinix Metal API client
	emClient  *packngo.Client
	projectID string
}

//+kubebuilder:rbac:groups=ketal.equinix.metal,resources=devices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ketal.equinix.metal,resources=devices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ketal.equinix.metal,resources=devices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Device object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DeviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var device ketalv1.Device
	if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
		if errors.IsNotFound(err) {
			// object not found, could have been deleted after
			// reconcile request, hence don't requeue
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Device")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	devices, _, err := r.emClient.Devices.List(r.projectID, nil)
	if err != nil {
		log.Error(err, "")
	}
	err = r.reconcileDevices(log, device, devices)
	if err != nil {
		log.Error(err, "")
	}
	// your logic here
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeviceReconciler) SetupWithManager(mgr ctrl.Manager, emClient *packngo.Client, projectID string) error {
	r.emClient = emClient
	r.projectID = projectID
	return ctrl.NewControllerManagedBy(mgr).
		For(&ketalv1.Device{}).
		Complete(r)
}

// reconcileDevices will compare the devices in the EM API to those in the Kubernetes API,
// it will then create devices that exist in the EM API but haven't been created in the EM
// API.
func (r *DeviceReconciler) reconcileDevices(log logr.Logger, kDevice ketalv1.Device, emDevices []packngo.Device) error {
	found := false
	for x := range emDevices {
		if kDevice.Name == emDevices[x].Hostname {
			found = true
		}
	}
	if !found {
		log.Info("New Device being Created", "Device Name", kDevice.Spec.Hostname)
		newDevice := packngo.DeviceCreateRequest{
			Hostname:     kDevice.Spec.Hostname,
			Facility:     []string{kDevice.Spec.Facility},
			OS:           kDevice.Spec.OS,
			BillingCycle: "hourly",
			ProjectID:    r.projectID,
			Plan:         kDevice.Spec.DeviceType,
		}
		_, _, err := r.emClient.Devices.Create(&newDevice)
		if err != nil {
			return err
		}
	}
	return nil
}
