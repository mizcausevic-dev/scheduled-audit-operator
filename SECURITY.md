# Security Policy

`scheduled-audit-operator` is a Kubernetes controller. It reconciles
`ScheduledAudit` resources into owned `CronJob` objects in the same cluster. It
makes no outbound network calls and serves only the manager's metrics (`:8080`)
and health (`:8081`) probes.

Operational notes:

- The bundled RBAC is least-privilege: scoped to `scheduledaudits` (+ status)
  and `cronjobs`.
- The container runs as non-root, read-only root filesystem, no privilege
  escalation, all capabilities dropped (see chart `values.yaml`).
- The operator creates CronJobs that run the **image you specify** with the
  privileges of the audit job's own ServiceAccount — review audit images and run
  them with a minimally-scoped ServiceAccount. The operator does not grant the
  audit job any permissions of its own.

## Supported versions

Only the latest tagged release is supported.

## Reporting a vulnerability

Please use GitHub Security Advisories for private disclosure:

- [Open a security advisory](https://github.com/mizcausevic-dev/scheduled-audit-operator/security/advisories/new)

Do not file public issues for security reports.
