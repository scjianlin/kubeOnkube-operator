package cluster

const (
	tokenFileTemplate = `%s,admin,admin,system:masters
`

	auditWebhookConfig = `
apiVersion: v1
kind: Config
clusters:
  - name: tke
    cluster:
      server: {{.AuditBackendAddress}}/apis/audit.k8s.io/v1/events/sink/{{.ClusterName}}
      insecure-skip-tls-verify: true
current-context: tke
contexts:
  - context:
      cluster: tke
    name: tke
`
)
