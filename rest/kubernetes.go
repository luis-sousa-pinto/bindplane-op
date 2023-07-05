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
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  name: bindplane-agent
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  name: bindplane-agent
  namespace: bindplane-agent
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: bindplane-agent
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
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
  name: bindplane-agent
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: bindplane-agent
subjects:
- kind: ServiceAccount
  name: bindplane-agent
  namespace: bindplane-agent
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  name: bindplane-node-agent
  namespace: bindplane-agent
spec:
  ports:
  - appProtocol: grpc
    name: otlp-grpc
    port: 4317
    protocol: TCP
    targetPort: 4317
  - appProtocol: http
    name: otlp-http
    port: 4318
    protocol: TCP
    targetPort: 4318
  selector:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  name: bindplane-node-agent-headless
  namespace: bindplane-agent
spec:
  clusterIP: None
  ports:
  - appProtocol: grpc
    name: otlp-grpc
    port: 4317
    protocol: TCP
    targetPort: 4317
  - appProtocol: http
    name: otlp-http
    port: 4318
    protocol: TCP
    targetPort: 4318
  selector:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: bindplane-node-agent
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: node
  namespace: bindplane-agent
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: bindplane-agent
      app.kubernetes.io/component: node
  template:
    metadata:
      labels:
        app.kubernetes.io/name: bindplane-agent
        app.kubernetes.io/component: node
    spec:
      serviceAccount: bindplane-agent
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
          ports:
            - containerPort: 4317
              name: otlpgrpc
            - containerPort: 4318
              name: otlphttp
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
                  fieldPath: spec.nodeName
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
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: cluster
  name: bindplane-agent
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: cluster
  name: bindplane-agent
  namespace: bindplane-agent
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: bindplane-agent
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: cluster
  namespace: bindplane-agent
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
  name: bindplane-agent
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: cluster
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: bindplane-agent
subjects:
- kind: ServiceAccount
  name: bindplane-agent
  namespace: bindplane-agent
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bindplane-cluster-agent
  labels:
    app.kubernetes.io/name: bindplane-agent
    app.kubernetes.io/component: cluster
  namespace: bindplane-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: bindplane-agent
      app.kubernetes.io/component: cluster
  template:
    metadata:
      labels:
        app.kubernetes.io/name: bindplane-agent
        app.kubernetes.io/component: cluster
    spec:
      serviceAccount: bindplane-agent
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
          - mountPath: /etc/otel/storage
            name: storage
      volumes:
        - name: config
          emptyDir: {}
        - name: storage
          emptyDir: {}
`
