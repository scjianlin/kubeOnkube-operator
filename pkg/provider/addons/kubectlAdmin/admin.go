package kubectlAdmin

const (
	kubectlAdmin = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
  generation: 1
  name: kubectl-admin
  namespace: kubesphere-controls-system
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      kubesphere.io/username: admin
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        kubesphere.io/username: admin
    spec:
      containers:
      - image: kubesphere/kubectl:v1.0.0
        imagePullPolicy: IfNotPresent
        name: kubectl
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/localtime
          name: host-time
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: kubesphere-cluster-admin
      serviceAccountName: kubesphere-cluster-admin
      terminationGracePeriodSeconds: 30
      volumes:
      - hostPath:
          path: /etc/localtime
          type: ""
        name: host-time
`
)
