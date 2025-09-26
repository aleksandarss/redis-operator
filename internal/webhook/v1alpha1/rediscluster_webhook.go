/*
Copyright 2025 aleksandarss.

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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	datav1alpha1 "github.com/aleksandarss/redis-operator.git/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var redisclusterlog = logf.Log.WithName("rediscluster-resource")

// SetupRedisClusterWebhookWithManager registers the webhook for RedisCluster in the manager.
func SetupRedisClusterWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&datav1alpha1.RedisCluster{}).
		WithDefaulter(&RedisClusterCustomDefaulter{}).
		WithValidator(&RedisClusterCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-data-calic-cloud-v1alpha1-rediscluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=data.calic.cloud,resources=redisclusters,verbs=create;update,versions=v1alpha1,name=vrediscluster-v1alpha1.kb.io,admissionReviewVersions=v1

// RedisClusterCustomValidator struct is responsible for validating the RedisCluster resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type RedisClusterCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

type RedisClusterCustomDefaulter struct{}

var _ webhook.CustomValidator = &RedisClusterCustomValidator{}
var _ webhook.CustomDefaulter = &RedisClusterCustomDefaulter{}

func (d *RedisClusterCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	rc, ok := obj.(*datav1alpha1.RedisCluster)
	if !ok {
		return fmt.Errorf("expected a RedisCluster object but got %T", obj)
	}

	// Defaults
	if rc.Spec.Auth.SecretName == "" {
		rc.Spec.Auth.SecretName = rc.Name + "-redis-auth"
	}

	if rc.Spec.Service.Port == nil {
		p := int32(6379)
		rc.Spec.Service.Port = &p
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type RedisCluster.
func (v *RedisClusterCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	rediscluster, ok := obj.(*datav1alpha1.RedisCluster)
	if !ok {
		return nil, fmt.Errorf("expected a RedisCluster object but got %T", obj)
	}

	if rediscluster.Spec.DiscoveryMode == "WriterService" && rediscluster.Spec.Service.Type == "" {
		return nil, fmt.Errorf("service type must be specified when discovery mode is WriterService")
	}

	if rediscluster.Spec.Sentinel.Replicas < 3 {
		return nil, fmt.Errorf("sentinel replicas must be at least 3 for high availability")
	}

	if rediscluster.Spec.Sentinel.Replicas%2 == 0 {
		return nil, fmt.Errorf("sentinel replicas must be an odd number")
	}

	redisclusterlog.Info("Validation for RedisCluster upon creation", "name", rediscluster.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type RedisCluster.
func (v *RedisClusterCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	rediscluster, ok := newObj.(*datav1alpha1.RedisCluster)
	if !ok {
		return nil, fmt.Errorf("expected a RedisCluster object for the newObj but got %T", newObj)
	}
	redisclusterlog.Info("Validation for RedisCluster upon update", "name", rediscluster.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type RedisCluster.
func (v *RedisClusterCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	rediscluster, ok := obj.(*datav1alpha1.RedisCluster)
	if !ok {
		return nil, fmt.Errorf("expected a RedisCluster object but got %T", obj)
	}
	redisclusterlog.Info("Validation for RedisCluster upon deletion", "name", rediscluster.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
