package controller

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	opsv1alpha1 "github.com/mizcausevic-dev/scheduled-audit-operator/api/v1alpha1"
)

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := opsv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

func audit() *opsv1alpha1.ScheduledAudit {
	return &opsv1alpha1.ScheduledAudit{
		ObjectMeta: metav1.ObjectMeta{Name: "nightly", Namespace: "ns1"},
		Spec: opsv1alpha1.ScheduledAuditSpec{
			Schedule: "0 3 * * *",
			Image:    "ghcr.io/mizcausevic-dev/mcp-registry-risk-scanner:latest",
			Args:     []string{"./server.json"},
		},
	}
}

func reconcilerFor(t *testing.T, sa *opsv1alpha1.ScheduledAudit) (*ScheduledAuditReconciler, client.Client) {
	t.Helper()
	s := newScheme(t)
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(sa).WithStatusSubresource(sa).Build()
	return &ScheduledAuditReconciler{Client: cl, Scheme: s}, cl
}

func req() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: "nightly", Namespace: "ns1"}}
}

func TestReconcile_CreatesCronJob(t *testing.T) {
	r, cl := reconcilerFor(t, audit())
	if _, err := r.Reconcile(context.Background(), req()); err != nil {
		t.Fatalf("reconcile error: %v", err)
	}

	var cj batchv1.CronJob
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "nightly-audit", Namespace: "ns1"}, &cj); err != nil {
		t.Fatalf("expected CronJob created: %v", err)
	}
	if cj.Spec.Schedule != "0 3 * * *" {
		t.Fatalf("bad schedule: %s", cj.Spec.Schedule)
	}
	if len(cj.OwnerReferences) != 1 || cj.OwnerReferences[0].Name != "nightly" {
		t.Fatalf("expected owner ref, got %v", cj.OwnerReferences)
	}

	var sa opsv1alpha1.ScheduledAudit
	if err := cl.Get(context.Background(), req().NamespacedName, &sa); err != nil {
		t.Fatal(err)
	}
	if sa.Status.CronJobName != "nightly-audit" {
		t.Fatalf("status.cronJobName not set: %q", sa.Status.CronJobName)
	}
	if c := meta.FindStatusCondition(sa.Status.Conditions, "Ready"); c == nil || c.Status != metav1.ConditionTrue {
		t.Fatalf("expected Ready=True, got %v", sa.Status.Conditions)
	}
}

func TestReconcile_UpdatesExistingCronJob(t *testing.T) {
	r, cl := reconcilerFor(t, audit())
	if _, err := r.Reconcile(context.Background(), req()); err != nil {
		t.Fatal(err)
	}
	// change schedule, reconcile again
	var sa opsv1alpha1.ScheduledAudit
	if err := cl.Get(context.Background(), req().NamespacedName, &sa); err != nil {
		t.Fatal(err)
	}
	sa.Spec.Schedule = "*/15 * * * *"
	if err := cl.Update(context.Background(), &sa); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Reconcile(context.Background(), req()); err != nil {
		t.Fatal(err)
	}
	var cj batchv1.CronJob
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "nightly-audit", Namespace: "ns1"}, &cj); err != nil {
		t.Fatal(err)
	}
	if cj.Spec.Schedule != "*/15 * * * *" {
		t.Fatalf("schedule not updated: %s", cj.Spec.Schedule)
	}
}

func TestReconcile_InvalidSpec(t *testing.T) {
	sa := audit()
	sa.Spec.Schedule = "not-a-cron"
	r, cl := reconcilerFor(t, sa)
	if _, err := r.Reconcile(context.Background(), req()); err != nil {
		t.Fatalf("reconcile should not error on invalid spec: %v", err)
	}
	var cj batchv1.CronJob
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "nightly-audit", Namespace: "ns1"}, &cj); !apierrors.IsNotFound(err) {
		t.Fatalf("expected no CronJob for invalid spec, err=%v", err)
	}
	var got opsv1alpha1.ScheduledAudit
	if err := cl.Get(context.Background(), req().NamespacedName, &got); err != nil {
		t.Fatal(err)
	}
	if c := meta.FindStatusCondition(got.Status.Conditions, "Ready"); c == nil || c.Reason != "InvalidSpec" {
		t.Fatalf("expected Ready=False/InvalidSpec, got %v", got.Status.Conditions)
	}
}

func TestReconcile_NotFound(t *testing.T) {
	s := newScheme(t)
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	r := &ScheduledAuditReconciler{Client: cl, Scheme: s}
	if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "ghost", Namespace: "ns1"}}); err != nil {
		t.Fatalf("missing object should be a no-op, got %v", err)
	}
}
