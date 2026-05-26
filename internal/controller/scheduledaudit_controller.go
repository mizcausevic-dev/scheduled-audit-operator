package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1alpha1 "github.com/mizcausevic-dev/scheduled-audit-operator/api/v1alpha1"
	"github.com/mizcausevic-dev/scheduled-audit-operator/internal/job"
)

// ScheduledAuditReconciler reconciles a ScheduledAudit into an owned CronJob.
type ScheduledAuditReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ops.kineticgain.com,resources=scheduledaudits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.kineticgain.com,resources=scheduledaudits/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile validates a ScheduledAudit and syncs its CronJob.
func (r *ScheduledAuditReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var sa opsv1alpha1.ScheduledAudit
	if err := r.Get(ctx, req.NamespacedName, &sa); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := job.Validate(sa.Spec); err != nil {
		log.Info("invalid ScheduledAudit", "error", err.Error())
		meta.SetStatusCondition(&sa.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidSpec",
			Message: err.Error(),
		})
		return ctrl.Result{}, r.Status().Update(ctx, &sa)
	}

	cj := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: job.CronJobName(&sa), Namespace: sa.Namespace}}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, cj, func() error {
		desired, derr := job.DesiredCronJob(&sa)
		if derr != nil {
			return derr
		}
		cj.Labels = desired.Labels
		cj.Spec = desired.Spec
		return ctrl.SetControllerReference(&sa, cj, r.Scheme)
	})
	if err != nil {
		if apierrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("synced audit CronJob", "cronjob", cj.Name, "op", op)
	}

	sa.Status.CronJobName = cj.Name
	sa.Status.ActiveJobs = len(cj.Status.Active)
	sa.Status.LastScheduleTime = cj.Status.LastScheduleTime
	meta.SetStatusCondition(&sa.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "CronJobSynced",
		Message: "audit CronJob " + cj.Name + " is in sync",
	})
	return ctrl.Result{}, r.Status().Update(ctx, &sa)
}

// SetupWithManager wires the reconciler into the manager.
func (r *ScheduledAuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1alpha1.ScheduledAudit{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
