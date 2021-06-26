/*
Copyright 2021.

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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	addPodNameLabelAnnotation = "lowang.io/add-pod-name-label"
	podNameLabel              = "lowang.io/pod-name"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// your logic here
	var pod corev1.Pod
	if err := r.Client.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "get pod error")
		return ctrl.Result{}, err
	}

	/*
	   Step 1: Add or remove the label.
	*/

	labelShouldBePresent := pod.Annotations[addPodNameLabelAnnotation] == "true"
	labelIsPresent := pod.Labels[podNameLabel] == pod.Name

	if labelShouldBePresent == labelIsPresent {
		// The desired state and actual state of the Pod are the same.
		// No further action is required by the operator at this moment.
		logger.Info("no update required")
		return ctrl.Result{}, nil
	}

	if labelShouldBePresent {
		// If the label should be set but is not, set it.
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels[podNameLabel] = pod.Name
		logger.Info("adding label")
	} else {
		// If the label should not be set but is, remove it.
		delete(pod.Labels, podNameLabel)
		logger.Info("removing label")
	}
	/*
	   Step 2: Update the Pod in the Kubernetes API.
	*/

	if err := r.Update(ctx, &pod); err != nil {
		if apierrors.IsConflict(err) {
			// The Pod has been updated since we read it.
			// Requeue the Pod to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		if apierrors.IsNotFound(err) {
			// The Pod has been deleted since we read it.
			// Requeue the Pod to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "unable to update Pod")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
