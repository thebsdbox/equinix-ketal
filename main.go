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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/packethost/packngo"
	ketalv1 "github.com/thebsdbox/equinix-ketal/api/v1"
	"github.com/thebsdbox/equinix-ketal/controllers"
	"github.com/thebsdbox/equinix-ketal/pkg/metal"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(ketalv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var providerConfig string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&providerConfig, "provider-config", "/cloud-sa.json", "The path to a config file usually located in a secret")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ketal.equinix.metal",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Get Equinix Metal details
	metalConfig, err := getMetalConfig(providerConfig)
	if err != nil {
		setupLog.Error(err, "unable to parse Equinix Metal API configuration")
		os.Exit(1)
	}

	// set up our client
	client := packngo.NewClientWithAuth("", metalConfig.AuthToken, nil)
	client.UserAgent = fmt.Sprintf("cloud-provider-equinix-metal/%s %s", "hackathon", client.UserAgent)

	// TODO: We perform a test here to determine if we can actually query the API
	// this could be a goroutine and error later..
	_, _, err = client.Projects.List(nil)
	if err != nil {
		setupLog.Error(err, "unable to connect to Equinix Metal API")
		os.Exit(1)
	}

	// Start the controllers
	if err = (&controllers.EipReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, client); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Eip")
		os.Exit(1)
	}
	if err = (&controllers.DeviceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, client, metalConfig.ProjectID); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Device")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	// START THE RECONCILER
	go metal.Reconcile(logger, client, mgr.GetClient(), metalConfig.ProjectID)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

const (
	apiKeyName    = "METAL_API_KEY"
	projectIDName = "METAL_PROJECT_ID"
)

// Config configuration for a provider, includes authentication token, project ID ID, and optional override URL to talk to a different Equinix Metal API endpoint
type Config struct {
	AuthToken string  `json:"apiKey"`
	ProjectID string  `json:"projectId"`
	BaseURL   *string `json:"base-url,omitempty"`
}

func getMetalConfig(providerConfig string) (Config, error) {
	// get our token and project
	var config, rawConfig Config
	if providerConfig != "" {
		configBytes, err := ioutil.ReadFile(providerConfig)
		if err != nil {
			return config, fmt.Errorf("failed to get read configuration file at path %s: %v", providerConfig, err)
		}
		err = json.Unmarshal(configBytes, &rawConfig)
		if err != nil {
			return config, fmt.Errorf("failed to process json of configuration file at path %s: %v", providerConfig, err)
		}
	}

	// read env vars; if not set, use rawConfig
	apiToken := os.Getenv(apiKeyName)
	if apiToken == "" {
		apiToken = rawConfig.AuthToken
	}
	config.AuthToken = apiToken

	projectID := os.Getenv(projectIDName)
	if projectID == "" {
		projectID = rawConfig.ProjectID
	}
	config.ProjectID = projectID
	return config, nil
}
