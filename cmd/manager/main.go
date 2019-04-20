/*
Copyright 2019 Fabian Sabau <fabian.sabau@gmail.com>.

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
	"errors"
	"flag"
	"os"

	"github.com/fbsb/pingdom-operator/pkg/apis"
	"github.com/fbsb/pingdom-operator/pkg/controller"
	"github.com/fbsb/pingdom-operator/pkg/pingdom/httpcheck"
	"github.com/fbsb/pingdom-operator/pkg/webhook"
	"github.com/russellcardullo/go-pingdom/pingdom"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	metricsAddr     string
	pingdomUsername string
	pingdomPassword string
	pingdomApiKey   string
)

func main() {
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&pingdomUsername, "pingdom-username", "", "The pingdom username.")
	flag.StringVar(&pingdomPassword, "pingdom-password", "", "The pingdom password.")
	flag.StringVar(&pingdomApiKey, "pingdom-api-key", "", "The pingdom API key.")

	flag.Parse()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")

	pingdomConfig, err := getPingdomConfig()
	if err != nil {
		log.Error(err, "could not get pingdom config")
		os.Exit(1)
	}

	pingdomClient, err := pingdom.NewClientWithConfig(pingdomConfig)
	if err != nil {
		log.Error(err, "could not create pingdom client")
		os.Exit(1)
	}

	err = httpcheck.InitService(pingdomClient)
	if err != nil {
		log.Error(err, "could not initialize httpcheck service")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	log.Info("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	log.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	// Setup all Controllers
	log.Info("Setting up controller")
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	log.Info("setting up webhooks")
	if err := webhook.AddToManager(mgr); err != nil {
		log.Error(err, "unable to register webhooks to the manager")
		os.Exit(1)
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}

func getPingdomConfig() (conf pingdom.ClientConfig, err error) {
	username := flagOrEnv(pingdomUsername, "PINGDOM_USERNAME")
	if username == "" {
		err = errors.New("could not find pingdom username")
		return
	}

	password := flagOrEnv(pingdomPassword, "PINGDOM_PASSWORD")
	if password == "" {
		err = errors.New("could not find pingdom password")
		return
	}

	apiKey := flagOrEnv(pingdomApiKey, "PINGDOM_API_KEY")
	if apiKey == "" {
		err = errors.New("could not find pingdom api key")
		return
	}

	conf.User = username
	conf.Password = password
	conf.APIKey = apiKey

	return
}

func flagOrEnv(v string, env string) string {
	if len(v) > 0 {
		return v
	}

	e, ok := os.LookupEnv(env)
	if ok && len(e) > 0 {
		return e
	}

	return ""
}
