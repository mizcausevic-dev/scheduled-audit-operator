# Changelog

## v0.1.0 — 2026-05-25

- Initial release: Kubernetes operator that runs recurring governance/compliance audits as CronJobs.
- `ScheduledAudit` CRD (group `ops.kineticgain.com/v1alpha1`): schedule (cron or @shortcut), image, command/args, suspend; status subresource + printer columns (Schedule / Image / Active / LastRun).
- Reconciler creates/updates an owned CronJob (`restartPolicy: OnFailure`, `concurrencyPolicy: Forbid`) and reflects active count + last schedule time; invalid schedules are rejected with `Ready=False`.
- Composes with the Kinetic Gain governance CLIs (e.g. `mcp-registry-risk-scanner`) as the audit image.
- Pure `internal/job` core (schedule validation + desired CronJob) plus fake-client reconciler tests — full `go test ./...` runs without a cluster.
- Helm chart (CRD, least-privilege RBAC, non-root manager), raw `config/` manifests, distroless Dockerfile.
- CI: `go vet` / `go test` / `go build` + `helm lint`. AGPL-3.0-or-later, Dependabot (gomod / actions / docker).
