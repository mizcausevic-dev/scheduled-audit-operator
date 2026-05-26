// Package v1alpha1 contains the ScheduledAudit API — declarative recurring
// governance/compliance audits run as Kubernetes CronJobs.
// +kubebuilder:object:generate=true
// +groupName=ops.kineticgain.com
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupVersion is the group/version for this API.
var GroupVersion = schema.GroupVersion{Group: "ops.kineticgain.com", Version: "v1alpha1"}

// SchemeBuilder registers the API types with a runtime scheme.
var SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

// AddToScheme adds the API types to a scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func init() {
	SchemeBuilder.Register(&ScheduledAudit{}, &ScheduledAuditList{})
}
