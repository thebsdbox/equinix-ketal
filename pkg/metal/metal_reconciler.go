package metal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/packethost/packngo"
	ketalv1 "github.com/thebsdbox/equinix-ketal/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// eipCacher is a cache for all eips the reconciller knows about
var eipCacher map[string]string

// deviceCacher is a cache for all devices the reconciler knows about
var deviceCacher map[string]string

// This package contains the reconcilliation loop that will synchronise Equinix Metal objects
// and Kubernetes objects. It will run as a blocking function, so should be started as a go
// routing from elsewhere.

// Reconcile will handle the "polling" of equinix Metal objects
func Reconcile(log logr.Logger, emClient *packngo.Client, kClient client.Client, projectID string) error {

	// Initialise the caches
	eipCacher = make(map[string]string)
	deviceCacher = make(map[string]string)

	// Begin the loop
	for {

		// Get the Kubernetes API objects
		var deviceObjects ketalv1.DeviceList
		err := kClient.List(context.TODO(), &deviceObjects, &client.ListOptions{})
		if err != nil {
			log.Error(err, "")
		}

		devices, _, err := emClient.Devices.List(projectID, nil)
		if err != nil {
			log.Error(err, "")
		}

		err = reconcileDevices(log, kClient, deviceObjects, devices)
		if err != nil {
			log.Error(err, "")
		}
		for x := range devices {

			newDevice := ketalv1.Device{
				ObjectMeta: v1.ObjectMeta{
					Name:      devices[x].Hostname,
					Namespace: "default",
				},
				Spec: ketalv1.DeviceSpec{
					Hostname:   devices[x].Hostname,
					DeviceType: devices[x].Plan.Name,
					UUID:       devices[x].ID,
					Address:    devices[x].GetNetworkInfo().PublicIPv4,
					Facility:   fmt.Sprintf("%s(%s)", devices[x].Facility.Name, devices[x].Facility.Code),
					Metro:      devices[x].Facility.Metro.Name,
					OS:         devices[x].OS.Distro,
				},
			}
			err := kClient.Create(context.TODO(), &newDevice, &client.CreateOptions{})
			if err != nil {
				if !errors.IsAlreadyExists(err) {
					log.Error(err, "")
				}
			}
		}

		eips, _, err := emClient.ProjectIPs.List(projectID, nil)
		if err != nil {
			log.Error(err, "")
		}
		for x := range eips {
			// IPv4 support today.. seriously .. does anyone use IPv6?
			if eips[x].AddressFamily == 4 {
				newEIP := ketalv1.Eip{
					ObjectMeta: v1.ObjectMeta{
						Name:      eips[x].Address,
						Namespace: "default",
					},
					Spec: ketalv1.EipSpec{
						Address: eips[x].Address,
						UUID:    eips[x].ID,
						Public:  eips[x].Public,
					},
				}
				err := kClient.Create(context.TODO(), &newEIP, &client.CreateOptions{})
				if err != nil {
					if !errors.IsAlreadyExists(err) {
						log.Error(err, "")
					}
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
	// If we get here something has gone wrong...
	// should this end gracefully?
	// is the answer to the ultimate question 42?
}

// reconcileDevices will compare the devices in the EM API to those in the Kubernetes API,
// it will then remove devices from the Kubernetes API that are no longer present in the
// EM API
func reconcileDevices(log logr.Logger, kClient client.Client, kDevices ketalv1.DeviceList, emDevices []packngo.Device) error {
	for x := range kDevices.Items {
		found := false
		for y := range emDevices {
			if kDevices.Items[x].Name == emDevices[y].Hostname {
				found = true
			}
		}
		if !found {
			log.Info("Removing Device from Kubernetes API", "device", kDevices.Items[x].Name)
			err := kClient.Delete(context.TODO(), kDevices.Items[x].DeepCopy(), &client.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
