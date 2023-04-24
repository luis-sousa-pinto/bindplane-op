// Copyright observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

const k8sDaemonsetChart = `apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/name: observiq-node-collector
  name: observiq-node-collector
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: observiq-node-collector
  labels:
    app.kubernetes.io/name: observiq-node-collector
rules:
- apiGroups:
  - ""
  resources:
  - events
  - namespaces
  - namespaces/status
  - nodes
  - nodes/spec
  - nodes/stats
  - nodes/proxy
  - pods
  - pods/status
  - replicationcontrollers
  - replicationcontrollers/status
  - resourcequotas
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - daemonsets
  - deployments
  - replicasets
  - statefulsets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - daemonsets
  - deployments
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - get
  - list
  - watch
- apiGroups:
    - autoscaling
  resources:
    - horizontalpodautoscalers
  verbs:
    - get
    - list
    - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: observiq-node-collector
  labels:
    app.kubernetes.io/name: observiq-node-collector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: observiq-node-collector
subjects:
- kind: ServiceAccount
  name: observiq-node-collector
  namespace: default
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: observiq-node-collector
  labels:
    app.kubernetes.io/name: observiq-node-collector
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: observiq-node-collector
  template:
    metadata:
      labels:
        app.kubernetes.io/name: observiq-node-collector
    spec:
      serviceAccount: observiq-node-collector
      containers:
        - name: opentelemetry-collector
          image: observiq/observiq-otel-collector:{{ .version }}
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              memory: 200Mi
              cpu: 100m
            limits:
              memory: 200Mi
          env:
            - name: OPAMP_ENDPOINT
              value: {{ .remoteURL }}
            - name: OPAMP_SECRET_KEY
              value: {{ .secretKey }}
            - name: OPAMP_AGENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPAMP_LABELS
              value: configuration={{ .configuration }},container-platform=kubernetes-daemonset
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - mountPath: /run/log/journal
              name: runlog
              readOnly: true
            - mountPath: /var/log
              name: varlog
              readOnly: true
            - mountPath: /var/lib/docker/containers
              name: dockerlogs
              readOnly: true
            - mountPath: /etc/otel/storage
              name: storage
      volumes:
        - name: runlog
          hostPath:
            path: /run/log/journal
        - name: varlog
          hostPath:
            path: /var/log
        - name: dockerlogs
          hostPath:
            path: /var/lib/docker/containers
        - name: storage
          hostPath:
            path: /var/lib/observiq/otelcol/container
      securityContext:
        runAsUser: 0
`

const k8sDeploymentChart = `apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: observiq-cluster-collector
  name: observiq-cluster-collector
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: observiq-cluster-collector
  labels:
    app.kubernetes.io/name: observiq-cluster-collector
rules:
- apiGroups:
  - ""
  resources:
  - events
  - namespaces
  - namespaces/status
  - nodes
  - nodes/spec
  - nodes/stats
  - nodes/proxy
  - pods
  - pods/status
  - replicationcontrollers
  - replicationcontrollers/status
  - resourcequotas
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - daemonsets
  - deployments
  - replicasets
  - statefulsets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - daemonsets
  - deployments
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - get
  - list
  - watch
- apiGroups:
    - autoscaling
  resources:
    - horizontalpodautoscalers
  verbs:
    - get
    - list
    - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: observiq-cluster-collector
  labels:
    app: observiq-cluster-collector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: observiq-cluster-collector
subjects:
- kind: ServiceAccount
  name: observiq-cluster-collector
  namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: observiq-cluster-collector
  labels:
    app: observiq-cluster-collector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: observiq-cluster-collector
  template:
    metadata:
      labels:
        app: observiq-cluster-collector
    spec:
      serviceAccount: observiq-cluster-collector
      containers:
        - name: opentelemetry-container
          image: observiq/observiq-otel-collector:{{ .version }}
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              memory: 200Mi
              cpu: 100m
            limits:
              memory: 200Mi
          env:
            - name: OPAMP_ENDPOINT
              value: {{ .remoteURL }}
            - name: OPAMP_SECRET_KEY
              value: {{ .secretKey }}
            - name: OPAMP_AGENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPAMP_LABELS
              value: configuration={{ .configuration }},container-platform=kubernetes-deployment
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
`
