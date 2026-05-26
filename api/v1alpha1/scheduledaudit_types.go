package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScheduledAuditSpec declares a recurring audit job.
type ScheduledAuditSpec struct {
	// Schedule is a cron expression (5 fields) or a @shortcut (e.g. @daily).
	// +kubebuilder:validation:Required
	Schedule string `json:"schedule"`

	// Image is the container image that runs the audit (e.g. a Kinetic Gain
	// governance CLI like mcp-registry-risk-scanner).
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Command overrides the image entrypoint.
	// +optional
	Command []string `json:"command,omitempty"`

	// Args are passed to the audit container.
	// +optional
	Args []string `json:"args,omitempty"`

	// Suspend pauses the CronJob without deleting it.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// ScheduledAuditStatus reflects the managed CronJob.
type ScheduledAuditStatus struct {
	// +optional
	CronJobName string `json:"cronJobName,omitempty"`
	// +optional
	ActiveJobs int `json:"activeJobs,omitempty"`
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=saudit
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Active",type=integer,JSONPath=`.status.activeJobs`
// +kubebuilder:printcolumn:name="LastRun",type=date,JSONPath=`.status.lastScheduleTime`

// ScheduledAudit is a recurring governance/compliance audit run as a CronJob.
type ScheduledAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledAuditSpec   `json:"spec,omitempty"`
	Status ScheduledAuditStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScheduledAuditList is a list of ScheduledAudit resources.
type ScheduledAuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledAudit `json:"items"`
}
