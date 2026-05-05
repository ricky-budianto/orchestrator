# Kubernetes Deployment Guide

This guide provides complete instructions for deploying the CES-iBridge Orchestrator to Kubernetes using k9s.

## 📋 Prerequisites

- Kubernetes cluster (v1.24+)
- kubectl configured with cluster access
- k9s installed (`brew install k9s` on macOS)
- Docker images pushed to registry
- PostgreSQL database (in-cluster or external)
- RabbitMQ (in-cluster or external)

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────┐
│              Kubernetes Cluster                 │
├─────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐          │
│  │  Ingress     │───▶│  Service     │          │
│  │  Controller  │    │  (LoadBalancer)         │
│  └──────────────┘    └──────┬───────┘          │
│                              │                   │
│  ┌──────────────────────────▼────────────────┐ │
│  │      Orchestrator Deployment              │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  │ │
│  │  │  Pod 1  │  │  Pod 2  │  │  Pod 3  │  │ │
│  │  └─────────┘  └─────────┘  └─────────┘  │ │
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌──────────────┐    ┌──────────────┐         │
│  │  PostgreSQL  │    │  RabbitMQ    │         │
│  │  StatefulSet │    │  StatefulSet │         │
│  └──────────────┘    └──────────────┘         │
│                                                 │
│  ┌─────────────────────────────────────────┐  │
│  │         LGTM Stack (Monitoring)         │  │
│  │  Loki | Grafana | Tempo | Prometheus   │  │
│  └─────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

## 📁 Directory Structure

Create the following directory structure:

```
k8s/
├── 01-namespace.yaml
├── 02-configmap.yaml
├── 03-secrets.yaml
├── 04-postgres.yaml
├── 05-rabbitmq.yaml
├── 06-redis.yaml
├── 07-deployment.yaml
├── 08-service.yaml
├── 09-ingress.yaml
├── 10-hpa.yaml
├── monitoring/
│   ├── tempo.yaml
│   ├── loki.yaml
│   ├── prometheus.yaml
│   └── grafana.yaml
└── README.md
```

## 🚀 Quick Start

### 1. Create Kubernetes Manifests

```bash
# Create k8s directory
mkdir -p k8s/monitoring
cd k8s
```

### 2. Apply Manifests in Order

```bash
# Apply all manifests
kubectl apply -f 01-namespace.yaml
kubectl apply -f 02-configmap.yaml
kubectl apply -f 03-secrets.yaml
kubectl apply -f 04-postgres.yaml
kubectl apply -f 05-rabbitmq.yaml
kubectl apply -f 06-redis.yaml
kubectl apply -f 07-deployment.yaml
kubectl apply -f 08-service.yaml
kubectl apply -f 09-ingress.yaml
kubectl apply -f 10-hpa.yaml

# Apply monitoring stack
kubectl apply -f monitoring/
```

### 3. Monitor with k9s

```bash
# Launch k9s
k9s -n ces-orchestrator

# Key commands in k9s:
# :po      - View pods
# :svc     - View services
# :deploy  - View deployments
# :ing     - View ingresses
# :logs    - View pod logs
# :describe - Describe resource
```

## 📝 Kubernetes Manifests

### 01-namespace.yaml

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ces-orchestrator
  labels:
    name: ces-orchestrator
    environment: production
    app: orchestrator
```

### 02-configmap.yaml

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: orchestrator-config
  namespace: ces-orchestrator
data:
  # Application Configuration
  PROJECT_NAME: "ces-orchestrator"
  PROJECT_MODULE: "ces-orchestrator-service"
  APP_PORT: "8081"
  LOG_LEVEL: "INFO"
  DB_PROVIDER: "postgresql"
  LOCALE_LANGUAGE: "en"
  TIMEZONE: "Asia/Jakarta"
  MODULE_VERSION: "1.0.0"
  ENVIRONMENT: "production"

  # Redis Configuration
  REDIS_USE: "true"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  REDIS_DB: "0"
  REDIS_TLS_ENABLED: "false"

  # PostgreSQL Configuration
  POSTGRE_SQL_HOST: "postgres-service"
  POSTGRE_SQL_PORT: "5432"
  POSTGRE_SQL_DB_NAME: "ces_orchestrator"
  POSTGRE_SQL_MAX_OPEN_CONNECTION: "25"
  POSTGRE_SQL_MAX_IDLE_CONNECTION: "10"

  # RabbitMQ Configuration
  RABBITMQURL: "amqp://guest:guest@rabbitmq-service:5672/"
  RABBITMQ_QUEUE: "ces-orchestrator"

  # Telemetry Configuration
  TELEMETRY_ENABLED: "true"
  TEMPO_ENDPOINT: "tempo-service:4318"
  PROMETHEUS_PORT: "8081"

  # Retention Configuration
  RETENTION_LOG: "30"
  CRON_CLEANUP_WORKFLOW: "0 2 * * *"
```

### 03-secrets.yaml

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: orchestrator-secrets
  namespace: ces-orchestrator
type: Opaque
stringData:
  # PostgreSQL Credentials
  POSTGRE_SQL_USER: "postgres"
  POSTGRE_SQL_PASSWORD: "your-secure-password-here"
  POSTGRE_SQL_ORCHESTRATOR_USER: "postgres"
  POSTGRE_SQL_ORCHESTRATOR_PASSWORD: "your-secure-password-here"

  # Redis Password
  REDIS_PASSWORD: ""

  # Crypto RSA Key (base64 encoded private key)
  CRYPTO_RSA: "your-base64-encoded-rsa-private-key"

  # Elasticsearch Credentials (if used)
  ELASTICSEARCH_USER: ""
  ELASTICSEARCH_PASSWORD: ""
```

### 04-postgres.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: ces-orchestrator
spec:
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
  clusterIP: None
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: ces-orchestrator
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14-alpine
        ports:
        - containerPort: 5432
          name: postgres
        env:
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: POSTGRE_SQL_USER
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: POSTGRE_SQL_PASSWORD
        - name: POSTGRES_DB
          valueFrom:
            configMapKeyRef:
              name: orchestrator-config
              key: POSTGRE_SQL_DB_NAME
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
```

### 05-rabbitmq.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: rabbitmq-service
  namespace: ces-orchestrator
spec:
  selector:
    app: rabbitmq
  ports:
    - name: amqp
      port: 5672
      targetPort: 5672
    - name: management
      port: 15672
      targetPort: 15672
  clusterIP: None
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: rabbitmq
  namespace: ces-orchestrator
spec:
  serviceName: rabbitmq-service
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq
  template:
    metadata:
      labels:
        app: rabbitmq
    spec:
      containers:
      - name: rabbitmq
        image: rabbitmq:3-management-alpine
        ports:
        - containerPort: 5672
          name: amqp
        - containerPort: 15672
          name: management
        env:
        - name: RABBITMQ_DEFAULT_USER
          value: "guest"
        - name: RABBITMQ_DEFAULT_PASS
          value: "guest"
        volumeMounts:
        - name: rabbitmq-storage
          mountPath: /var/lib/rabbitmq
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command:
            - rabbitmq-diagnostics
            - ping
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - rabbitmq-diagnostics
            - ping
          initialDelaySeconds: 20
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: rabbitmq-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 5Gi
```

### 06-redis.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: ces-orchestrator
spec:
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
  clusterIP: None
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: ces-orchestrator
spec:
  serviceName: redis-service
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
          name: redis
        command:
        - redis-server
        - --appendonly
        - "yes"
        volumeMounts:
        - name: redis-storage
          mountPath: /data
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          tcpSocket:
            port: 6379
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: redis-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 5Gi
```

### 07-deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orchestrator
  namespace: ces-orchestrator
  labels:
    app: orchestrator
    version: v1
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: orchestrator
  template:
    metadata:
      labels:
        app: orchestrator
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8081"
        prometheus.io/path: "/metrics"
    spec:
      # Init container to wait for dependencies
      initContainers:
      - name: wait-for-postgres
        image: busybox:1.35
        command: ['sh', '-c', 'until nc -z postgres-service 5432; do echo waiting for postgres; sleep 2; done;']
      - name: wait-for-rabbitmq
        image: busybox:1.35
        command: ['sh', '-c', 'until nc -z rabbitmq-service 5672; do echo waiting for rabbitmq; sleep 2; done;']

      containers:
      - name: orchestrator
        image: your-registry/ces-orchestrator:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8081
          name: http
          protocol: TCP
        - containerPort: 8081
          name: metrics
          protocol: TCP

        # Environment variables from ConfigMap
        envFrom:
        - configMapRef:
            name: orchestrator-config

        # Environment variables from Secrets
        env:
        - name: POSTGRE_SQL_USER
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: POSTGRE_SQL_USER
        - name: POSTGRE_SQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: POSTGRE_SQL_PASSWORD
        - name: POSTGRE_SQL_ORCHESTRATOR_USER
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: POSTGRE_SQL_ORCHESTRATOR_USER
        - name: POSTGRE_SQL_ORCHESTRATOR_PASSWORD
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: POSTGRE_SQL_ORCHESTRATOR_PASSWORD
        - name: CRYPTO_RSA
          valueFrom:
            secretKeyRef:
              name: orchestrator-secrets
              key: CRYPTO_RSA

        # Resource limits
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"

        # Liveness probe
        livenessProbe:
          httpGet:
            path: /
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        # Readiness probe
        readinessProbe:
          httpGet:
            path: /
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        # Startup probe (for slow starting apps)
        startupProbe:
          httpGet:
            path: /
            port: 8081
          initialDelaySeconds: 0
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 30

        # Security context
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: false

      # Pod security context
      securityContext:
        fsGroup: 1000

      # Termination grace period
      terminationGracePeriodSeconds: 30

      # Image pull secrets (if using private registry)
      # imagePullSecrets:
      # - name: registry-credentials
```

### 08-service.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: orchestrator-service
  namespace: ces-orchestrator
  labels:
    app: orchestrator
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8081"
    prometheus.io/path: "/metrics"
spec:
  type: LoadBalancer
  selector:
    app: orchestrator
  ports:
  - name: http
    port: 80
    targetPort: 8081
    protocol: TCP
  - name: metrics
    port: 8081
    targetPort: 8081
    protocol: TCP
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800
---
apiVersion: v1
kind: Service
metadata:
  name: orchestrator-headless
  namespace: ces-orchestrator
  labels:
    app: orchestrator
spec:
  clusterIP: None
  selector:
    app: orchestrator
  ports:
  - name: http
    port: 8081
    targetPort: 8081
    protocol: TCP
```

### 09-ingress.yaml

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: orchestrator-ingress
  namespace: ces-orchestrator
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "120"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "120"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - orchestrator.yourdomain.com
    secretName: orchestrator-tls
  rules:
  - host: orchestrator.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: orchestrator-service
            port:
              number: 80
```

### 10-hpa.yaml (Horizontal Pod Autoscaler)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: orchestrator-hpa
  namespace: ces-orchestrator
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: orchestrator
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

## 📊 Monitoring Stack

### monitoring/tempo.yaml

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tempo-config
  namespace: ces-orchestrator
data:
  tempo.yaml: |
    server:
      http_listen_port: 3200

    distributor:
      receivers:
        otlp:
          protocols:
            http:
              endpoint: 0.0.0.0:4318
            grpc:
              endpoint: 0.0.0.0:4317

    storage:
      trace:
        backend: local
        local:
          path: /tmp/tempo/blocks
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tempo
  namespace: ces-orchestrator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tempo
  template:
    metadata:
      labels:
        app: tempo
    spec:
      containers:
      - name: tempo
        image: grafana/tempo:latest
        args:
        - -config.file=/etc/tempo/tempo.yaml
        ports:
        - containerPort: 3200
          name: http
        - containerPort: 4317
          name: otlp-grpc
        - containerPort: 4318
          name: otlp-http
        volumeMounts:
        - name: config
          mountPath: /etc/tempo
        - name: storage
          mountPath: /tmp/tempo
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
      volumes:
      - name: config
        configMap:
          name: tempo-config
      - name: storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: tempo-service
  namespace: ces-orchestrator
spec:
  selector:
    app: tempo
  ports:
  - name: http
    port: 3200
    targetPort: 3200
  - name: otlp-grpc
    port: 4317
    targetPort: 4317
  - name: otlp-http
    port: 4318
    targetPort: 4318
```

## 🔧 Deployment Steps

### Step 1: Prepare Environment

```bash
# Create namespace
kubectl create namespace ces-orchestrator

# Set context to namespace
kubectl config set-context --current --namespace=ces-orchestrator
```

### Step 2: Create Secrets

```bash
# Create PostgreSQL password secret
kubectl create secret generic orchestrator-secrets \
  --from-literal=POSTGRE_SQL_USER=postgres \
  --from-literal=POSTGRE_SQL_PASSWORD='your-secure-password' \
  --from-literal=POSTGRE_SQL_ORCHESTRATOR_USER=postgres \
  --from-literal=POSTGRE_SQL_ORCHESTRATOR_PASSWORD='your-secure-password' \
  --from-literal=CRYPTO_RSA='your-base64-rsa-key' \
  -n ces-orchestrator
```

### Step 3: Apply Infrastructure

```bash
# Apply in order
kubectl apply -f 01-namespace.yaml
kubectl apply -f 02-configmap.yaml
kubectl apply -f 03-secrets.yaml

# Wait for namespace to be ready
kubectl get namespace ces-orchestrator
```

### Step 4: Deploy Dependencies

```bash
# Deploy PostgreSQL
kubectl apply -f 04-postgres.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=300s

# Deploy RabbitMQ
kubectl apply -f 05-rabbitmq.yaml

# Wait for RabbitMQ to be ready
kubectl wait --for=condition=ready pod -l app=rabbitmq --timeout=300s

# Deploy Redis
kubectl apply -f 06-redis.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod -l app=redis --timeout=300s
```

### Step 5: Initialize Database

```bash
# Create database
kubectl exec -it postgres-0 -- psql -U postgres -c "CREATE DATABASE ces_orchestrator;"

# Enable UUID extension
kubectl exec -it postgres-0 -- psql -U postgres -d ces_orchestrator -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
```

### Step 6: Deploy Application

```bash
# Deploy orchestrator
kubectl apply -f 07-deployment.yaml

# Wait for deployment
kubectl rollout status deployment/orchestrator

# Apply service
kubectl apply -f 08-service.yaml

# Apply ingress
kubectl apply -f 09-ingress.yaml

# Apply HPA
kubectl apply -f 10-hpa.yaml
```

### Step 7: Deploy Monitoring (Optional)

```bash
kubectl apply -f monitoring/
```

## 📱 Using k9s

### Basic k9s Commands

```bash
# Launch k9s in ces-orchestrator namespace
k9s -n ces-orchestrator

# Key bindings in k9s:
:po          # View pods
:deploy      # View deployments
:svc         # View services
:ing         # View ingresses
:cm          # View configmaps
:secrets     # View secrets
:pvc         # View persistent volume claims
:events      # View events

# Pod operations (when viewing pods):
l            # View logs
d            # Describe
s            # Shell into pod
<ctrl-k>     # Delete pod
y            # View YAML

# Filter resources:
/search-term # Filter by name
```

### Monitoring Deployment

```bash
# Watch deployment rollout
k9s -n ces-orchestrator
# Press '0' to view pods
# Use arrow keys to select orchestrator pod
# Press 'l' to view logs

# Check resource usage
# Press 'pulse' or use :pulse command
```

### Troubleshooting

```bash
# View pod logs
kubectl logs -f deployment/orchestrator -n ces-orchestrator

# Describe pod for events
kubectl describe pod <pod-name> -n ces-orchestrator

# Execute commands in pod
kubectl exec -it <pod-name> -n ces-orchestrator -- /bin/sh

# Port forward to local
kubectl port-forward service/orchestrator-service 8081:80 -n ces-orchestrator
```

## 🔍 Verification

### Check All Resources

```bash
# Check all resources in namespace
kubectl get all -n ces-orchestrator

# Check specific resources
kubectl get pods -n ces-orchestrator
kubectl get svc -n ces-orchestrator
kubectl get deploy -n ces-orchestrator
kubectl get ing -n ces-orchestrator
kubectl get hpa -n ces-orchestrator
```

### Test Application

```bash
# Get external IP
kubectl get svc orchestrator-service -n ces-orchestrator

# Test endpoint
curl http://<EXTERNAL-IP>/

# Test workflow
curl -X POST http://<EXTERNAL-IP>/api/ibridge/v2/orchestrate/complete-feature-demo \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{"user_id": "test123"}'
```

### Check Metrics

```bash
# Port forward Prometheus metrics
kubectl port-forward service/orchestrator-service 8081:8081 -n ces-orchestrator

# Access metrics
curl http://localhost:8081/metrics
```

## 🔄 Updates and Rollbacks

### Rolling Update

```bash
# Update image
kubectl set image deployment/orchestrator \
  orchestrator=your-registry/ces-orchestrator:v2.0.0 \
  -n ces-orchestrator

# Watch rollout
kubectl rollout status deployment/orchestrator -n ces-orchestrator

# View rollout history
kubectl rollout history deployment/orchestrator -n ces-orchestrator
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/orchestrator -n ces-orchestrator

# Rollback to specific revision
kubectl rollout undo deployment/orchestrator --to-revision=2 -n ces-orchestrator
```

## 🔐 Security Best Practices

1. **Use Secrets for Sensitive Data**
   - Store passwords, API keys in Kubernetes Secrets
   - Never commit secrets to Git
   - Use external secret management (e.g., HashiCorp Vault)

2. **Network Policies**
   - Implement network policies to restrict pod communication
   - Only allow necessary ingress/egress traffic

3. **RBAC**
   - Create service accounts with minimal permissions
   - Use role-based access control

4. **Pod Security**
   - Run as non-root user
   - Use read-only root filesystem where possible
   - Drop all capabilities

5. **Resource Limits**
   - Always set resource requests and limits
   - Prevent resource exhaustion

## 📈 Scaling

### Manual Scaling

```bash
# Scale deployment
kubectl scale deployment orchestrator --replicas=5 -n ces-orchestrator
```

### Auto-scaling

HPA automatically scales based on CPU/memory usage (already configured in 10-hpa.yaml)

### Cluster Auto-scaling

Configure cluster autoscaler to add nodes when needed.

## 🛠️ Maintenance

### Database Backup

```bash
# Backup PostgreSQL
kubectl exec postgres-0 -n ces-orchestrator -- \
  pg_dump -U postgres ces_orchestrator > backup.sql
```

### Log Rotation

Logs are automatically rotated by Kubernetes. Access via:

```bash
kubectl logs -f deployment/orchestrator --tail=100 -n ces-orchestrator
```

## 📞 Support

For issues:
1. Check pod logs: `kubectl logs <pod-name> -n ces-orchestrator`
2. Check events: `kubectl get events -n ces-orchestrator --sort-by='.lastTimestamp'`
3. Describe resources: `kubectl describe <resource> -n ces-orchestrator`
4. Use k9s for interactive debugging

## 🔗 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [k9s Documentation](https://k9scli.io/)
- [Helm Charts](https://helm.sh/)
- [Kustomize](https://kustomize.io/)

---

**Last Updated**: 2025-10-05
**Version**: 1.0.0
