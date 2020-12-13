/*


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

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientretry "k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/alexandrevilain/sipp-operator/api/v1alpha1"
	"github.com/alexandrevilain/sipp-operator/internal/resource"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// SippScenarioRunReconciler reconciles a SippScenarioRun object
type SippScenarioRunReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=sipp.alexandrevilain.dev,resources=sippscenarioruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sipp.alexandrevilain.dev,resources=sippscenarioruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sipp.alexandrevilain.dev,resources=sippscenario,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sipp.alexandrevilain.dev,resources=sippscenario/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=jobs/status,verbs=get

func (r *SippScenarioRunReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sippscenariorun", req.NamespacedName)

	log.Info("Retrieved a new reconcile event")

	scenarioRun := &v1alpha1.SippScenarioRun{}
	err := r.Get(ctx, req.NamespacedName, scenarioRun)
	if err != nil {
		log.Error(err, "unable to fetch SippScenarioRun")
		return ctrl.Result{}, err
	}

	// Get the linked scenario
	scenario := &v1alpha1.SippScenario{}
	err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: scenarioRun.Spec.ScenarioRef.Name}, scenario)
	if err != nil {
		log.Error(err, "unable to fetch SippScenario")
		return ctrl.Result{}, err
	}


	resourceBuilder := resource.SippResourceBuilder{
		Instance: scenarioRun,
		Scenario: scenario,
		Scheme:   r.Scheme,
	}

	builders, err := resourceBuilder.ResourceBuilders()
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, builder := range builders {
		resource, err := builder.Build()
		if err != nil {
			return ctrl.Result{}, err
		}

		var operationResult controllerutil.OperationResult
		err = clientretry.RetryOnConflict(clientretry.DefaultRetry, func() error {
			var apiError error
			operationResult, apiError = controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
				return builder.Update(resource)
			})
			return apiError
		})
		if err != nil {
			log.Error(err, "unable to create or update resource")
		}

		log.Info("builder finished", "operationResult", operationResult)
	}

	// Update status
	childJob := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: scenarioRun.ChildResourceName("job")}, childJob)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "unable to get child job")
		return ctrl.Result{}, err
	}

	scenarioRun.Status.Active = childJob.Status.Active
	scenarioRun.Status.Failed = childJob.Status.Failed
	scenarioRun.Status.Succeeded = childJob.Status.Succeeded

	err = r.Status().Update(ctx, scenarioRun)
	if err != nil {
		log.Error(err, "unable to update SippScenarioRun status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SippScenarioRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SippScenarioRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
