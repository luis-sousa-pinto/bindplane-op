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
      initContainers:
        - name: setup-volumes
          image: ghcr.io/observiq/observiq-otel-collector:{{ .version }}
          securityContext:
            # Required for changing permissions from
            # root to otel user in emptyDir volume.
            runAsUser: 0
          command:
            - "chown"
            - "otel:"
            - "/etc/otel/config"
          volumeMounts:
            - mountPath: /etc/otel/config
              name: config
        - name: copy-configs
          image: ghcr.io/observiq/observiq-otel-collector:{{ .version }}
          command:
            - 'sh'
            - '-c'
            - 'cp config.yaml config/ && cp logging.yaml config/ && chown -R otel:otel config/'
          volumeMounts:
            - mountPath: /etc/otel/config
              name: config
      containers:
        - name: opentelemetry-collector
          image: ghcr.io/observiq/observiq-otel-collector:{{ .version }}
          imagePullPolicy: IfNotPresent
          securityContext:
            readOnlyRootFilesystem: true
            # Required for reading container logs hostPath.
            runAsUser: 0
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
            # The collector process updates config.yaml
            # and manager.yaml when receiving changes
            # from the OpAMP server.
            - name: CONFIG_YAML_PATH
              value: /etc/otel/config/config.yaml
            - name: MANAGER_YAML_PATH
              value: /etc/otel/config/manager.yaml
            - name: LOGGING_YAML_PATH
              value: /etc/otel/config/logging.yaml
          volumeMounts:
            - mountPath: /etc/otel/config
              name: config
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
        - name: config
          emptyDir: {}
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
`

const k8sDeploymentChart = `apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/name: observiq-cluster-collector
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
    app.kubernetes.io/name: observiq-cluster-collector
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
    app.kubernetes.io/name: observiq-cluster-collector
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: observiq-cluster-collector
  template:
    metadata:
      labels:
        app.kubernetes.io/name: observiq-cluster-collector
    spec:
      serviceAccount: observiq-cluster-collector
      initContainers:
        - name: setup-volumes
          image: ghcr.io/observiq/observiq-otel-collector:{{ .version }}
          securityContext:
            # Required for changing permissions from
            # root to otel user in emptyDir volume.
            runAsUser: 0
          command:
            - "chown"
            - "otel:"
            - "/etc/otel/config"
          volumeMounts:
            - mountPath: /etc/otel/config
              name: config
        - name: copy-configs
          image: ghcr.io/observiq/observiq-otel-collector:{{ .version }}
          command:
            - 'sh'
            - '-c'
            - 'cp config.yaml config/ && cp logging.yaml config/ && chown -R otel:otel config/'
          volumeMounts:
            - mountPath: /etc/otel/config
              name: config
      containers:
        - name: opentelemetry-container
          image: ghcr.io/observiq/observiq-otel-collector:{{ .version }}
          imagePullPolicy: IfNotPresent
          securityContext:
            readOnlyRootFilesystem: true
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
            # The collector process updates config.yaml
            # and manager.yaml when receiving changes
            # from the OpAMP server.
            - name: CONFIG_YAML_PATH
              value: /etc/otel/config/config.yaml
            - name: MANAGER_YAML_PATH
              value: /etc/otel/config/manager.yaml
            - name: LOGGING_YAML_PATH
              value: /etc/otel/config/logging.yaml
          volumeMounts:
          - mountPath: /etc/otel/config
            name: config
      volumes:
        - name: config
          emptyDir: {}
`
