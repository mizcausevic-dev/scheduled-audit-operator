// Package job holds the pure (cluster-free) logic: validating a ScheduledAudit
// spec and building the desired CronJob.
package job

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	opsv1alpha1 "github.com/mizcausevic-dev/scheduled-audit-operator/api/v1alpha1"
)

var cronShortcuts = map[string]bool{
	"@yearly": true, "@annually": true, "@monthly": true, "@weekly": true,
	"@daily": true, "@midnight": true, "@hourly": true,
}

// ValidateSchedule accepts a 5-field cron expression or a known @shortcut.
func ValidateSchedule(schedule string) error {
	s := strings.TrimSpace(schedule)
	if s == "" {
		return fmt.Errorf("schedule must not be empty")
	}
	if strings.HasPrefix(s, "@") {
		if !cronShortcuts[s] {
			return fmt.Errorf("unknown schedule shortcut %q", s)
		}
		return nil
	}
	if fields := strings.Fields(s); len(fields) != 5 {
		return fmt.Errorf("cron schedule %q must have 5 fields, got %d", s, len(fields))
	}
	return nil
}

// Validate checks a spec independently of the cluster.
func Validate(spec opsv1alpha1.ScheduledAuditSpec) error {
	if err := ValidateSchedule(spec.Schedule); err != nil {
		return err
	}
	if strings.TrimSpace(spec.Image) == "" {
		return fmt.Errorf("image must not be empty")
	}
	return nil
}

// CronJobName returns the name of the managed CronJob.
func CronJobName(cr *opsv1alpha1.ScheduledAudit) string {
	return cr.Name + "-audit"
}

// DesiredCronJob computes the CronJob that should exist for a valid CR.
func DesiredCronJob(cr *opsv1alpha1.ScheduledAudit) (*batchv1.CronJob, error) {
	if err := Validate(cr.Spec); err != nil {
		return nil, err
	}
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "scheduled-audit-operator",
		"ops.kineticgain.com/audit":    cr.Name,
	}
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CronJobName(cr),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          cr.Spec.Schedule,
			Suspend:           &cr.Spec.Suspend,
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: labels},
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Containers: []corev1.Container{{
								Name:    "audit",
								Image:   cr.Spec.Image,
								Command: cr.Spec.Command,
								Args:    cr.Spec.Args,
							}},
						},
					},
				},
			},
		},
	}, nil
}
