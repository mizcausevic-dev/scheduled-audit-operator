# scheduled-audit-operator

A Kubernetes operator that runs **recurring governance/compliance audits as CronJobs**, declaratively. You write a `ScheduledAudit` — a schedule and an audit image; the operator reconciles an owned `CronJob` and reflects its run status back on the resource.

The scheduling control-plane of the [Kinetic Gain](https://suite.kineticgain.com) cloud-native lane: it turns the portfolio's governance CLIs into in-cluster scheduled checks. Point it at [`mcp-registry-risk-scanner`](https://github.com/mizcausevic-dev/mcp-registry-risk-scanner), a disclosure validator, or any audit image, and the audit runs on a cron you can `kubectl get`.

## Why

Governance only holds if it runs continuously, not once at review time. Hand-maintaining CronJobs per audit is toil — wrong RBAC, drift, no status surface. This operator makes the recurring audit a first-class resource: declare the schedule + image once, and the controller keeps the CronJob in sync, owns it (so deleting the audit cleans up the job), and surfaces `Active` count and last run time. Invalid schedules are rejected with a `Ready=False` condition instead of a silently-broken cron.

## Custom resource

```yaml
apiVersion: ops.kineticgain.com/v1alpha1
kind: ScheduledAudit
metadata:
  name: nightly-mcp-scan
spec:
  schedule: "0 3 * * *"          # cron or @daily/@hourly/...
  image: ghcr.io/mizcausevic-dev/mcp-registry-risk-scanner:latest
  args: ["/manifests/server.json", "--gate", "high"]
```

```
$ kubectl get saudit
NAME               SCHEDULE     IMAGE                                  ACTIVE   LASTRUN
nightly-mcp-scan   0 3 * * *    .../mcp-registry-risk-scanner:latest   0        5m
```

The CronJob runs with `restartPolicy: OnFailure` and `concurrencyPolicy: Forbid`, owned by the ScheduledAudit (garbage-collected with it). `suspend: true` pauses it without deletion.

## Install

```bash
helm install audits charts/scheduled-audit-operator
kubectl apply -f config/samples/sample.yaml
kubectl get saudit
```

The chart installs the CRD, least-privilege RBAC (scoped to `cronjobs` + the CRD), a non-root ServiceAccount, and the manager Deployment. Build the manager image from the distroless `Dockerfile`.

## Architecture

- `api/v1alpha1` — the `ScheduledAudit` CRD types.
- `internal/job` — **pure, cluster-free** validation + desired-CronJob construction (the unit-tested core).
- `internal/controller` — the reconciler: create-or-update the owned CronJob + reflect its status.
- `cmd` — the manager entrypoint.

`go test ./...` runs the full gate with no cluster: the builder is unit-tested and the reconciler is exercised end-to-end with controller-runtime's fake client (create, update, invalid-spec, not-found).

## License

AGPL-3.0-or-later — see [LICENSE](LICENSE).
