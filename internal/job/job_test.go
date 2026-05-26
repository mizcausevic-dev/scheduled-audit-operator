package job

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	opsv1alpha1 "github.com/mizcausevic-dev/scheduled-audit-operator/api/v1alpha1"
)

func cr() *opsv1alpha1.ScheduledAudit {
	return &opsv1alpha1.ScheduledAudit{
		ObjectMeta: metav1.ObjectMeta{Name: "nightly", Namespace: "ns1"},
		Spec: opsv1alpha1.ScheduledAuditSpec{
			Schedule: "0 3 * * *",
			Image:    "ghcr.io/mizcausevic-dev/mcp-registry-risk-scanner:latest",
			Args:     []string{"./server.json", "--gate", "high"},
		},
	}
}

func TestValidateSchedule(t *testing.T) {
	ok := []string{"0 3 * * *", "*/5 * * * *", "@daily", "@hourly"}
	for _, s := range ok {
		if err := ValidateSchedule(s); err != nil {
			t.Fatalf("expected %q valid: %v", s, err)
		}
	}
	bad := []string{"", "0 3 * *", "@bogus", "0 3 * * * *"}
	for _, s := range bad {
		if err := ValidateSchedule(s); err == nil {
			t.Fatalf("expected %q invalid", s)
		}
	}
}

func TestValidate(t *testing.T) {
	if err := Validate(cr().Spec); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
	s := cr().Spec
	s.Image = ""
	if err := Validate(s); err == nil {
		t.Fatal("expected error for empty image")
	}
}

func TestDesiredCronJob(t *testing.T) {
	cj, err := DesiredCronJob(cr())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cj.Name != "nightly-audit" || cj.Namespace != "ns1" {
		t.Fatalf("bad name/ns: %s/%s", cj.Name, cj.Namespace)
	}
	if cj.Spec.Schedule != "0 3 * * *" {
		t.Fatalf("bad schedule: %s", cj.Spec.Schedule)
	}
	if cj.Spec.ConcurrencyPolicy != batchv1.ForbidConcurrent {
		t.Fatalf("expected ForbidConcurrent")
	}
	c := cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
	if c.Image != cr().Spec.Image {
		t.Fatalf("bad image: %s", c.Image)
	}
	if len(c.Args) != 3 {
		t.Fatalf("args not propagated: %v", c.Args)
	}
	if cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy != corev1.RestartPolicyOnFailure {
		t.Fatalf("expected OnFailure restart policy")
	}
	if cj.Labels["ops.kineticgain.com/audit"] != "nightly" {
		t.Fatalf("missing audit label: %v", cj.Labels)
	}
}

func TestDesiredCronJob_Invalid(t *testing.T) {
	bad := cr()
	bad.Spec.Schedule = "nope"
	if _, err := DesiredCronJob(bad); err == nil {
		t.Fatal("expected error for invalid schedule")
	}
}

func TestSuspendPropagates(t *testing.T) {
	c := cr()
	c.Spec.Suspend = true
	cj, _ := DesiredCronJob(c)
	if cj.Spec.Suspend == nil || !*cj.Spec.Suspend {
		t.Fatal("expected suspend true to propagate")
	}
}
