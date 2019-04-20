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

package httpcheck

import (
	"context"

	pingdomv1alpha1 "github.com/fbsb/pingdom-operator/pkg/apis/pingdom/v1alpha1"
	"github.com/fbsb/pingdom-operator/pkg/pingdom/httpcheck"
	"github.com/go-logr/logr"
	"github.com/russellcardullo/go-pingdom/pingdom"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	finalizer = "finalizer.pingdom.fbsb.io"
)

// Add creates a new HttpCheck Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	service, err := httpcheck.ServiceInstance()
	if err != nil {
		return err
	}
	return add(mgr, newReconciler(mgr, service))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, service httpcheck.Service) reconcile.Reconciler {
	return &ReconcileHttpCheck{
		Client:  mgr.GetClient(),
		scheme:  mgr.GetScheme(),
		service: service,
		log:     log.Log.WithName("httpcheck-reconciler"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("httpcheck-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to HttpCheck
	err = c.Watch(&source.Kind{Type: &pingdomv1alpha1.HttpCheck{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileHttpCheck{}

// ReconcileHttpCheck reconciles a HttpCheck object
type ReconcileHttpCheck struct {
	client.Client
	scheme  *runtime.Scheme
	service httpcheck.Service
	log     logr.Logger
}

// Reconcile reads that state of the cluster for a HttpCheck object and makes changes based on the state read
// and what is in the HttpCheck.Spec
// +kubebuilder:rbac:groups=pingdom.fbsb.io,resources=httpchecks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingdom.fbsb.io,resources=httpchecks/status,verbs=get;update;patch
func (r *ReconcileHttpCheck) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log.Info("New reconcile request", "request", request)

	// Fetch the HttpCheck instance
	check := &pingdomv1alpha1.HttpCheck{}
	err := r.Get(context.TODO(), request.NamespacedName, check)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if !check.DeletionTimestamp.IsZero() {
		// The resource is going to be deleted but we need to do some cleanup first

		err := r.deleteHttpCheck(check)
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}

		removeFinalizer(check)
		err = r.Update(context.TODO(), check)
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}

		return reconcile.Result{}, nil
	}

	if !hasFinalizer(check) {
		// The resource is new so we need to make sure we add our finalizer first

		addFinalizer(check)
		err := r.Update(context.TODO(), check)
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}

		// The update will trigger the reconciliation again so we might as well just return here
		return reconcile.Result{}, nil
	}

	err = r.createOrUpdateHttpCheck(check)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileHttpCheck) deleteHttpCheck(check *pingdomv1alpha1.HttpCheck) error {
	if check.Status.PingdomID == 0 {
		return nil
	}

	_, err := r.service.Delete(check.Status.PingdomID)
	if err != nil {
		if _, ok := err.(*pingdom.PingdomError); ok {
			// just return if pingdom id does not exist
			return nil
		}

		return err
	}

	return nil
}

func (r *ReconcileHttpCheck) createOrUpdateHttpCheck(check *pingdomv1alpha1.HttpCheck) error {
	pCheck, err := httpcheck.SimpleHttpCheck(check.Spec.Name, check.Spec.URL)
	if err != nil {
		return r.statusFailure(check, err)
	}

	if check.Status.PingdomID != 0 {
		_, err := r.service.Update(check.Status.PingdomID, pCheck)

		if err == nil {
			return r.statusSuccess(check, check.Status.PingdomID)
		}

		if _, ok := err.(*pingdom.PingdomError); !ok {
			return err
		}
	}

	resp, err := r.service.Create(pCheck)
	if err != nil {
		if pErr, ok := err.(*pingdom.PingdomError); ok {
			return r.statusFailure(check, pErr)
		}
		return err
	}

	return r.statusSuccess(check, resp.ID)
}

func (r *ReconcileHttpCheck) statusFailure(check *pingdomv1alpha1.HttpCheck, err error) error {
	message := err.Error()
	if pErr, ok := err.(*pingdom.PingdomError); ok {
		message = pErr.Message
	}

	check.Status.Error = message
	check.Status.PingdomStatus = pingdomv1alpha1.StatusFail
	return r.Status().Update(context.TODO(), check)
}

func (r *ReconcileHttpCheck) statusSuccess(check *pingdomv1alpha1.HttpCheck, id int) error {
	check.Status.PingdomID = id
	check.Status.Error = ""
	check.Status.PingdomStatus = pingdomv1alpha1.StatusSuccess
	return r.Status().Update(context.TODO(), check)
}

// Helper functions to manage resource finalizers

func hasFinalizer(check *pingdomv1alpha1.HttpCheck) bool {
	for _, fin := range check.Finalizers {
		if fin == finalizer {
			return true
		}
	}

	return false
}

func addFinalizer(check *pingdomv1alpha1.HttpCheck) {
	check.Finalizers = append(check.Finalizers, finalizer)
}

func removeFinalizer(check *pingdomv1alpha1.HttpCheck) {
	var output []string

	for _, f := range check.Finalizers {
		if f != finalizer {
			output = append(output, f)
		}
	}

	check.Finalizers = output
}
