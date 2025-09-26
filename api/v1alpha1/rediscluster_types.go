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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisClusterSpec defines the desired state of RedisCluster
type RedisClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Redis image to use for the cluster nodes
	// +kubebuilder:default="redis:7.0-alpine"
	// +optional
	Image string `json:"image,omitempty"`

	// Number of replicas (not counting primary)
	// +kubebuilder:default=2
	Replicas int32 `json:"replicas,omitempty"`

	// Sentinel configuration
	// +optional
	Sentinel SentinelSpec `json:"sentinel,omitempty"`

	// Auth configuration
	// +optional
	Auth AuthSpec `json:"auth,omitempty"`

	// Storage configuration for redis data
	// +optional
	Storage StorageSpec `json:"storage,omitempty"`

	// Resources configuration (requests and limits) for redis pods
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Service configuration for redis writer service
	// +optional
	Service ServiceSpec `json:"writerService,omitempty"`

	// Discovery mode, default WriterService
	// +kubebuilder:default=WriterService
	// +kubebuilder:validation:Enum=WriterService;SentinelClient
	// +optional
	DiscoveryMode string `json:"discoveryMode,omitempty"`

	// optional redis.conf and sentinel config fragments (k/v or inline string)
	// +optional
	Config map[string]string `json:"config,omitempty"`

	// optional pod template to override default pod spec
	// +optional
	PodTemplate *corev1.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// ServiceSpec defines the configuration for a Kubernetes Service.
type ServiceSpec struct {
	// Type of the service (e.g., ClusterIP, NodePort, LoadBalancer)
	// +kubebuilder:default=ClusterIP
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`

	// Annotations to add to the service
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Additional labels to add to the service
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Port to expose
	// +optional
	Port *int32 `json:"port,omitempty"`
}

// StorageSpec defines storage configuration for Redis data.
type StorageSpec struct {
	// Size of the persistent volume claim (e.g., "1Gi")
	// +kubebuilder:default="1Gi"
	// +optional
	Size string `json:"size,omitempty"`

	// StorageClassName for the PVC
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// SentinelSpec defines the configuration for Redis Sentinel.
type SentinelSpec struct {
	// Number of Sentinel replicas
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas,omitempty"`

	// Image to use for Sentinel
	// +kubebuilder:default="redis:7.0-alpine"
	// +optional
	Image string `json:"image,omitempty"`

	// Port to expose Sentinel on
	// +kubebuilder:default=26379
	// +optional
	Port int32 `json:"port,omitempty"`
}

// AuthSpec defines authentication configuration for Redis.
type AuthSpec struct {
	// Password for Redis authentication
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// RedisClusterStatus defines the observed state of RedisCluster.
type RedisClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the RedisCluster resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// phase
	// +kubebuilder:default=Unknown
	// +kubebuilder:validation:Enum=Provisioning;Ready;Degraded;Failover;InProgress;Unknown
	// +optional
	Phase string `json:"phase,omitempty"`

	// endpoint of the Redis writer service
	// +optional
	Endpoints StatusEndpoints `json:"endpoints,omitempty"`

	// topology of the redis cluster
	// +optional
	Topology Topology `json:"topology,omitempty"`

	// observed version of the RedisCluster resource
	// +optional
	ObservedVersion string `json:"observedVersion,omitempty"`

	// last failover time
	// +optional
	LastFailoverTime *metav1.Time `json:"lastFailoverTime,omitempty"`
}

type StatusEndpoints struct {
	WriterService   string   `json:"writerService"`
	SentinelService []string `json:"sentinelService"`
}

type Topology struct {
	PrimaryPod  string   `json:"primaryPod"`
	ReplicaPods []string `json:"replicaPods"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RedisCluster is the Schema for the redisclusters API
type RedisCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of RedisCluster
	// +required
	Spec RedisClusterSpec `json:"spec"`

	// status defines the observed state of RedisCluster
	// +optional
	Status RedisClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisClusterList contains a list of RedisCluster
type RedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisCluster{}, &RedisClusterList{})
}
