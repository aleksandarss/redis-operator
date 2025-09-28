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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/aleksandarss/redis-operator.git/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:rbac:groups=data.calic.cloud,resources=redisclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.calic.cloud,resources=redisclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=data.calic.cloud,resources=redisclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// RedisClusterReconciler reconciles a RedisCluster object
type RedisClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Helper functions to get names and labels for resources
func (r *RedisClusterReconciler) names(rc *datav1alpha1.RedisCluster) (svcName string) {
	return rc.Name + "-redis"
}

func (r *RedisClusterReconciler) labels(rc *datav1alpha1.RedisCluster) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "redis",
		"app.kubernetes.io/instance": rc.Name,
		"app.kubernetes.io/part-of":  "redis-operator",
	}
}

func (r *RedisClusterReconciler) ensureHeadlessService(ctx context.Context, rc *datav1alpha1.RedisCluster) error {
	svcName := r.names(rc)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: rc.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.ClusterIP = corev1.ClusterIPNone
		svc.Spec.PublishNotReadyAddresses = true

		svc.Labels = r.labels(rc)
		svc.Spec.Selector = map[string]string{
			"app.kubernetes.io/name":     "redis",
			"app.kubernetes.io/instance": rc.Name,
		}

		port := int32(6379)
		if rc.Spec.Service.Port != nil {
			port = *rc.Spec.Service.Port
		}
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "redis",
				Port:       port,
				TargetPort: intstr.FromInt(int(port)),
			},
		}

		return controllerutil.SetControllerReference(rc, svc, r.Scheme)
	})

	return err
}

func (r *RedisClusterReconciler) ensureStatefulSet(ctx context.Context, rc *datav1alpha1.RedisCluster) error {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Name + "-redis",
			Namespace: rc.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, statefulSet, func() error {
		statefulSet.Labels = r.labels(rc)

		statefulSet.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/name":     "redis",
				"app.kubernetes.io/instance": rc.Name,
			},
		}

		statefulSet.Spec.ServiceName = r.names(rc)
		replicas := rc.Spec.Replicas
		statefulSet.Spec.Replicas = &replicas

		statefulSet.Spec.Template.Labels = map[string]string{
			"app.kubernetes.io/name":     "redis",
			"app.kubernetes.io/instance": rc.Name,
		}

		port := int32(6379)
		if rc.Spec.Service.Port != nil {
			port = *rc.Spec.Service.Port
		}

		statefulSet.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "redis",
				Image: rc.Spec.Image,
				Ports: []corev1.ContainerPort{
					{
						Name:          "redis",
						ContainerPort: port,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "redis-data",
						MountPath: "/data",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "REDIS_PASSWORD",
						Value: "secret123",
					},
					{
						Name:  "SERVICE_NAME",
						Value: "rediscluster-sample-redis",
					},
					{
						Name:  "NAMESPACE",
						Value: "redis-operator-demo",
					},
				},
				Command: []string{
					"sh",
					"-lc",
				},
				Args: []string{
					`
					ORD=${HOSTNAME##*-}
					BASE='redis-server --appendonly yes --protected-mode no --requirepass "$REDIS_PASSWORD" --masterauth "$REDIS_PASSWORD" --dir /data'
					if [ "$ORD" = "0" ]; then
					exec sh -lc "$BASE"
					else
					MASTER="${HOSTNAME%-*}-0.${SERVICE_NAME}.${NAMESPACE}.svc.${CLUSTER_DOMAIN:-cluster.local}"
					exec sh -lc "$BASE --replicaof $MASTER 6379"
					fi
					`,
					// "--appendonly yes",
					// "--protected-mode no",
					// "--requirepass $(REDIS_PASSWORD)",
					// "--masterauth $(REDIS_PASSWORD)",
					// "--dir /data",
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"sh",
								"-lc",
								"redis-cli -a \"$REDIS_PASSWORD\" PING | grep -q PONG",
							},
						},
					},
					InitialDelaySeconds: 5,
					PeriodSeconds:       5,
					TimeoutSeconds:      2,
					FailureThreshold:    3,
				},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"sh",
								"-lc",
								"redis-cli -a \"$REDIS_PASSWORD\" PING | grep -q PONG",
							},
						},
					},
					InitialDelaySeconds: 15,
					PeriodSeconds:       10,
					TimeoutSeconds:      2,
					FailureThreshold:    3,
				},
			},
		}

		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "redis-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					StorageClassName: rc.Spec.Storage.StorageClassName,
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(rc.Spec.Storage.Size),
						},
					},
				},
			},
		}

		return controllerutil.SetControllerReference(rc, statefulSet, r.Scheme)
	})

	return err
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *RedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	rc := &datav1alpha1.RedisCluster{}

	if err := r.Get(ctx, req.NamespacedName, rc); err != nil {
		logf.Log.Error(err, "unable to fetch RedisCluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.ensureHeadlessService(ctx, rc); err != nil {
		logf.Log.Error(err, "Failed to ensure headless service")
		return ctrl.Result{}, err
	}

	if err := r.ensureStatefulSet(ctx, rc); err != nil {
		logf.Log.Error(err, "Failed to ensure statefule set")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.RedisCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Named("rediscluster").
		Complete(r)
}
