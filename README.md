<a name="readme-top"></a>

<!-- PROJECT LOGO
<br />
<div align="center">

  <h3 align="center">CES-iBridge Orchestrartor</h3>

  <p align="center">
    An awesome README template to jumpstart your projects!
    <br />
    <a href="https://github.com/othneildrew/Best-README-Template"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/othneildrew/Best-README-Template">View Demo</a>
    ·
    <a href="https://github.com/othneildrew/Best-README-Template/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    ·
    <a href="https://github.com/othneildrew/Best-README-Template/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div> -->

## CES-iBridge Orchestrartor

<!-- TABLE OF CONTENTS -->

  <h3>Table of Contents</h3>
  <ol>
    <li>
      <a href="#about-the-project">About The Service</a>
    </li>
    <li><a href="#feature">Features</a></li>
    <li><a href="#dynamic-routing">Dynamic Routing</a>
      <ul>
        <li><a href="#instant-activation-of-new-workflows">Instant Activation of New Workflows</a></li>
        <li><a href="#updating-workflow-configurations">Updating Workflow Configurations</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a>
      <ul>
        <li><a href="#workflow-configuration">Workflow Configuration</a></li>
        <li><a href="#rest-api">REST API</a></li>
        <li><a href="#rabbitmq-rpc">RabbitMQ RPC</a></li>
      </ul>
    </li>
    <li><a href="#monitoring-and-resources">Monitoring and Resources</a>
      <ul>
        <li><a href="#table-definition">Table Definition</a>
          <ul>
            <li><a href="#workflow-configurations">Workflow Configurations</a></li>
            <li><a href="#workflow-histories">Workflow Histories</a></li>
            <li><a href="#workflow-states">Workflow States</a></li>
          </ul>
        </li>
        <li><a href="#apis-definition">APIs Definition</a></li>
      </ul>
    </li>
    <li><a href="#deployment-and-configuration">Deployment and Configuration</a></li>
    <li><a href="#milestones">Milestones</a></li>
  </ol>

## About The Project

<div style="text-align: justify">
The CES Orchestrator Service is a microservice that acts as an orchestrator for workflows defined in YAML files. It currently supports listening to REST and RabbitMQ RPC, and can trigger workers and conditions based on the defined workflow.
</p>

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Features

- Define workflow configuration using YAML file
- Listen to REST and RabbitMQ RPC
- Define workers and conditions in the workflow
- Send data to workers and define next target
- Send messages to RabbitMQ RPC and wait for response
- Define conditions and next targets based on the results
- Send data back to client at the end of the workflow
- Instant dynamic routing with ```/v2/orchestrate/:workflow_configuration_id```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Dynamic Routing

The Dynamic Routing feature enables instant handling of workflow configurations, ensuring seamless and uninterrupted service activation. Here's how it works:

### Instant Activation of New Workflows:
When a new workflow configuration is created through the API POST ```{{url}}/api/ibridge/v1/workflow_configuration``` and a 200 OK response is received, the new workflow becomes active immediately. It can be triggered instantly using the API POST ```{{url}}/api/ibridge/orchestrate/v2/:workflow_configuration_id```. For example, if a new workflow configuration with the ID ```demo_flow_v2``` is created and a 200 OK response is received, you can trigger this orchestration through POST ```{{url}}/api/ibridge/v2/orchestrate/demo_flow_v2``` immediately. **This process does not require the service to be relaunched or restarted.**

Notes:

The instant usage feature is currently available only for the POST method.
Using specific path definitions in the workflow configuration, such as POST ```{{url}}/api/ibridge/v1/demo_flow/sub_event```, still requires the service to be relaunched or restarted.

> **DO NOT USE** `/v2` for workflow configuration path because it was reserved only for dynamic routing process

### Updating Workflow Configurations:
Updating an existing workflow configuration is just as seamless. By using the API PUT ```{{url}}/api/ibridge/v1/workflow_configuration``` and receiving a 200 OK response, the existing workflow definition is updated instantly. Triggering this workflow endpoint will immediately process the orchestration with the new configuration, without needing to relaunch or restart the service. ***Updating path and method will not impacted as the same.***


<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Usage

To use the CES Orchestrator Service, you'll need to:

1. Define your workflow in a YAML file, specifying the listeners, workers, conditions, and end of workflow as needed.
2. Start the CES Orchestrator Service, passing in the path to the YAML file as a configuration parameter.
3. Send a REST or RabbitMQ RPC request to the CES Orchestrator Service, triggering the defined workflow.

Check this out :
* [Postman Collection](https://github.com/soluixdeveloper/ces-orchestratorService/tree/v4/sample)
* [Workflow Sample](https://github.com/soluixdeveloper/ces-orchestratorService/tree/v4/sample)
   
### Workflow Configuration

The workflow configuration is defined in a YAML file, with the following structure:

> **DO NOT USE** `/v2` for workflow configuration path because it was reserved only for dynamic routing process

```yaml
# Demo Flow
#
# 

name: demo_flow
requestType: demo_flow
path: /api/ibridge/v1/demo-flow
method: POST
startEvent:
  targetRef: demo_check_email

serviceTask:
  demo_check_email:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef: 
      authorization: ${{startEvent.Authorization}}
      body: 
        email: ${{startEvent.Body.email}}
    targetRef: 
      - condition_demo_check_email
    worker : middleware
  
  condition_demo_check_email:
    type: conditional
    condition: 
      variableCheck:
        - ${{demo_check_email.data.data.disposable_email}}
      value:
        - case:
          - true
          target: ibridge-reply
          status_code: 400
          save_global:
            status_code: 400
            message: "Invalid request"
            error_code: "5001"
    defaultTasK: demo_check_penghasilan

  demo_check_penghasilan:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef: 
      authorization: ${{startEvent.Authorization}}
      body: 
        penghasilan: ${{startEvent.Body.penghasilan}}
    targetRef: 
      - condition_demo_check_penghasilan
    worker : middleware
  
  condition_demo_check_penghasilan:
    type: conditional
    condition: 
      variableCheck:
        - ${{demo_check_penghasilan.data.data.reject}}
      value:
        - case:
          - true
          target: ibridge-reply
          status_code: 400
          save_global:
            status_code: 400
            message: "Invalid request"
            error_code: "5001"
    defaultTasK: demo_check_nik_socmed

  demo_check_nik_socmed:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef: 
      authorization: ${{startEvent.Authorization}}
      body: 
        nik: ${{startEvent.Body.nik}}
        social_media_url: ${{startEvent.Body.social_media_url}}
    targetRef: 
      - demo_get_limit
    worker : middleware

  demo_get_limit:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef: 
      authorization: ${{startEvent.Authorization}}
      body: 
        email_score: ${{demo_check_email.data.data.score}}
        penghasilan_score: ${{demo_check_penghasilan.data.data.score}}
        nik_socmed_score: ${{demo_check_nik_socmed.data.data.score}}
    targetRef: 
      - ibridge-reply
    worker : middleware
  
  ibridge-reply:
    endParsing: Separated
    sourceRef: 
      step_check_email: ${{demo_check_email.data}}
      step_check_penghasilan: ${{demo_check_penghasilan.data}}
      step_check_nik_socmed: ${{demo_check_nik_socmed.data}}
      step_get_limit: ${{demo_get_limit.data}}
    type: endEvent
    taskRef: ibridge-reply
```

<div style="text-align: justify">
This is an example of a workflow configuration for CES Orchestrator Service, which is a microservice orchestrator that allows users to define workflows using a YAML file. The example workflow is named `demo_flow` and it defines a set of tasks that will be executed when triggered by a REST API call.

The configuration file starts by defining the basic information about the REST API endpoint that will trigger this workflow. In this case, the endpoint is named `/api/ibridge/v1/demo-flow` and it accepts POST requests.

The workflow starts with a `startEvent` that defines the first task to be executed, which is `demo_check_email`. This task is of type `worker` and is defined with a `sourceRef` that specifies the input data for the task. The `targetRef` specifies the next task or tasks that will be executed after this task.

The next task is `condition_demo_check_email`, which is a conditional task. The `condition` specifies a check to be performed on the output data of the previous task, `demo_check_email`. If the check evaluates to true, the workflow will end with an HTTP 400 error code and a message indicating that the request was invalid. If the check evaluates to false, the workflow will continue to the next task, `demo_check_penghasilan`.

The `demo_get_limit` task takes the output data of the previous tasks and combines it into a single object that will be used as the final output of the workflow. This output object is passed to the `ibridge-reply` task, which marks the end of the workflow and sends the output object back to the client that triggered the workflow.

`ibridge-reply` is an end event in the workflow, which indicates the end of the workflow execution. It is triggered after all the preceding tasks and conditions are completed successfully.

In this workflow, `ibridge-reply` is configured with `endParsing: Separated`, which means that the data sent back in response to the original API request will be separated into different fields, each corresponding to a step in the workflow. These fields are:

- `step_check_email`: contains the data returned by the `demo_check_email` worker task
- `step_check_penghasilan`: contains the data returned by the `demo_check_penghasilan` worker task
- `step_check_nik_socmed`: contains the data returned by the `demo_check_nik_socmed` worker task
- `step_get_limit`: contains the data returned by the `demo_get_limit` worker task

This allows the client that initiated the workflow to receive a detailed response, indicating the outcome of each step in the workflow. The data contained in these fields can be used by the client to make further decisions or to display information to the user.

Overall, `ibridge-reply` is an important part of the workflow, as it indicates the successful completion of the workflow and provides a way for the client to receive and process the output data.

Overall, this example workflow demonstrates how CES Orchestrator Service can be used to define and execute complex workflows that involve multiple tasks and conditional logic.
</p>

<p align="right">(<a href="#readme-top">back to top</a>)</p>


### REST API

To trigger a workflow via REST, send a POST request to the `/api/ibridge/v1/demo-flow` endpoint with the workflow data in the request body. For example:

```bash
curl --location 'https://ces-new-api.soluix.ai/api/ibridge/v1/demo-flow' \
--data-raw '{
    "email": "test@gmail.com",
    "penghasilan": "30000000",
    "nik": "320124019281871891",
    "social_media_url": "google.com"
}'
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### RabbitMQ RPC

To trigger a workflow via RabbitMQ RPC, send a message to the specified queue with the workflow data in the message body. The response will be sent back to the specified reply queue. For example:

```go
type messageQueue struct {
 RequestID     string      `json:"request_id,omitempty"`
 CorrelationID string      `json:"correlation_id,omitempty"`
 RequestType   string      `json:"request_type,omitempty"`
 ContentType   string      `json:"content_type,omitempty"`
 Header        interface{} `json:"header,omitempty"`
 Key           string      `json:"key,omitempty"`
 Value         interface{} `json:"value,omitempty"`
 NeedResponse  bool        `json:"need_response,omitempty"`
}

bodyMessage := map[string]string{
    "email": "test@gmail.com",
    "penghasilan": "30000000",
    "nik": "320124019281871891",
    "social_media_url": "google.com"
}
message := messageQueue{RequestID: "12345", CorrelationID: "123456", RequestType: "demo_flow", Value: bodyMessage}
mqMessage, _ := json.Marshal(message)
q, err := ch.QueueDeclare(
  "ces_orchestrator",       // name
  false,                    // durable
  false,                    // delete when unused
  false,                    // exclusive
  false,                    // noWait
  nil,                      // arguments
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err = ch.PublishWithContext(ctx,
  "",           // exchange
  "my_replyto", // routing key
  false,        // mandatory
  false,        // immediate
  amqp.Publishing{
    ContentType:   "text/plain",
    CorrelationId: corrId,
    ReplyTo:       q.Name,
    Body:          mqMessage,
})
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Monitoring and Resources
Service resources and tables have been reworked, and new tables have been added to facilitate easier monitoring and debugging. Existing table data must be migrated to the new table format to ensure continued functionality.

### Table Definition
### Workflow Configurations
The `workflow_configurations` table contains all workflow definitions along with their orchestration process configurations. This new table replaces the old `workflow_configuration` table and includes additional columns for enhanced functionality. 

Below is the table structure : 

| id      | name      | path  | method | descriptions | configuration | revision_number | audit.. |
| ------- | --------- | ----- | ------ | ------------ | ------------- | --------------- | ------- |
| flow_1  | Flow 1    | /api/ibridge/v1/flow1 | GET | Flow 1 Desc | ICAjCiAgIwx | 0 | ... |
| flow_2  | Flow 2    | /api/ibridge/v1/flow2 | PUT | Flow 2 Desc | ICAjCiAgIwx | 0 | ... |
| ....    | ....      | ....                  | ... | ....        | ....        | . | ... |

Notes: 
In the new model, the path and method for each workflow **must be explicitly defined** in their respective columns. These definitions will **override** any previously set values within the workflow configuration itself.

### Workflow Histories

The `workflow_configuration_histories` table stores historical data of executed orchestration processes. This table can be used to track and debug processes by retrieving data using a unique ID as a reference, and it also allows you to specify which workflow configuration ID you want to query.

Below is the table structure : 

| id         | workflow_configuration_id | status | request | response | additional_info | audit.. |
| ---------- | --------- | -------- | ----------------- | --------------------- | ---- | --- |
| WH_uuid_1  | flow_1    | EXECUTED | {"Body":{ ...}}   | {"ResponseData":{..}} | .... | ... |
| WH_uuid_2  | flow_1    | ACTIVE   | {"Params":{ ...}} | {"ResponseData":{..}} | .... | ... |
| ....       | ....      | ....     | ...               | ....                  | .... | ... |

### Workflow States

The `workflow_states` table contains all worker state data for each orchestration process that refers to `workflow_configuration_histories` in the column `workflow_id`. This table provides detailed information about each orchestration process, including which workers were executed and where any failure points occurred. Obtain comprehensive details about each step of the orchestration process, including the status and results of individual workers.

Below is the table structure : 

| id         | workflow_id  | workflow_request_type  | state                           | audit.. |
| ---------- | ------------ | ---------------------- | ------------------------------- | ------- |
| WS_uuid_1  | WH_uuid_1    | flow_1                 | {"worker1:{.....},"worker2":..} | ....    | 
| WS_uuid_2  | WH_uuid_2    | flow_1                 | {"worker1:{.....},"worker2":..} | ....    | 
| ....       | ....         | ....                   | ...                             | ....    | 

### APIs Definition
Each resource has its own API functionality, utilizing CRUD-based requests. However, someresources are immutable, meaning they cannot be updated or deleted once created. 

* Immutable Resources: Certain resources, such as historical data and workflow states, may beimmutable. This means once created, they cannot be updated or deleted.

* CRUD Operations: Standard CRUD (Create, Read, Update, Delete) operations are used whereapplicable, ensuring a consistent and intuitive API design.

Below are the APIs for each resource:

| Resources              | Action  | Path                                          | Method |
| ---------------------- | ------- | --------------------------------------------- | ----   |
| Workflow Configuration | CREATE  | /api/ibridge/v1/workflow_configuration        | POST   | 
|                        | READ    | /api/ibridge/v1/workflow_configuration/:id    | GET    | 
|                        | READ    | /api/ibridge/v1/workflow_configuration?query  | GET    | 
|                        | UPDATE  | /api/ibridge/v1/workflow_configuration        | PUT    | 
|                        | DELETE  | /api/ibridge/v1/workflow_configuration        | DEL    | 
| Workflow Histories     | READ    | /api/ibridge/v1/workflow_history/:id          | GET    | 
|                        | READ    | /api/ibridge/v1/workflow_history?query        | GET    | 
| Workflow States        | READ    | /api/ibridge/v1/workflow_state/:id            | GET    | 
|                        | READ    | /api/ibridge/v1/workflow_state?query          | GET    |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Observability and Tracing

CES Orchestrator Service includes comprehensive observability through OpenTelemetry integration with the LGTM stack (Loki, Grafana, Tempo, Prometheus).

### Features

- **Distributed Tracing**: End-to-end visibility of request flows from HTTP ingress through workflow execution to database operations
- **Database Tracing**: Automatic instrumentation of all GORM database operations with OpenTelemetry
- **Metrics**: Prometheus metrics for workflows, HTTP requests, and performance
- **Log Aggregation**: Structured logs sent to Elasticsearch or Loki

### Database Tracing

All database operations are automatically traced with the following attributes:

| Attribute | Description | Example |
| --------- | ----------- | ------- |
| `db.system` | Database system | `postgresql` |
| `db.name` | Database name | `ces_orchestrator` |
| `db.operation` | SQL operation type | `SELECT`, `INSERT`, `UPDATE`, `DELETE` |
| `db.table` | Table name | `workflow_histories` |
| `db.statement` | Actual SQL query | `SELECT * FROM workflow_histories WHERE id = ?` |
| `db.rows_affected` | Rows modified (write ops) | `1` |

### Span Hierarchy Example

```
HTTP POST /api/ibridge/v1/workflow-path (17ms)
├── workflow: execute workflow_name (15ms)
│   ├── db.query: query workflow_configurations (2ms)
│   ├── db.query: create workflow_histories (1ms)
│   ├── worker: call_middleware_service (8ms)
│   ├── db.query: create workflow_states (1ms)
│   └── db.query: update workflow_histories (1ms)
```

### Configuration

Enable telemetry in your environment configuration:

```env
# Enable OpenTelemetry
TELEMETRY_ENABLED=true

# Tempo OTLP endpoint (no protocol prefix)
TEMPO_ENDPOINT=localhost:4318

# Prometheus metrics port
PROMETHEUS_PORT=8081
```

### Starting the LGTM Stack

```bash
# Start observability stack
docker-compose -f docker-compose.monitoring.yml up -d

# Or use the convenience script
./start-lgtm-stack.sh
```

Services will be available at:
- **Grafana**: http://localhost:3001 (admin/admin)
- **Tempo**: http://localhost:3200
- **Prometheus**: http://localhost:9090
- **Loki**: http://localhost:3100

### Querying Traces in Grafana

Access Grafana Explore and use these TraceQL queries:

```
# All database operations
{span.db.system="postgresql"}

# Specific operations
{span.db.operation="INSERT"}
{span.db.table="workflow_histories"}

# Combined filters
{span.db.system="postgresql" && span.db.operation="SELECT" && span.db.table="workflow_configurations"}
```

For detailed E2E testing instructions, see `.spec-workflow/specs/database-tracing/E2E_TESTING_GUIDE.md`

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Deployment and Configuration

CES Orchestrator Service is containerized, you can deploy it on kubernetes. Below is the deployment example:

```yaml
--- # Common Environment Variables (config)
apiVersion: v1
kind: ConfigMap
metadata:
  name: ces-config
  namespace: ces
data:
  APP_NAME: CES
  APP_PORT: "3000"
  LOG_LEVEL: "0"
  REDIS_USE: "true"
  PROJECT_NAME: "ces"
--- # Common Environment Variables (secret)
apiVersion: v1
kind: Secret
metadata:
  name: ces-secret
  namespace: ces
data:
  POSTGRE_SQL_USER: "LQ=="
  POSTGRE_SQL_PASSWORD: "LQ=="
--- #Deployment ces-orchestrator
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: ces
  name: ces-orchestrator
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: ces-orchestrator
  replicas: 3
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ces-orchestrator
    spec:
      nodeSelector:
        eks.amazonaws.com/nodegroup: ng-sbx
      containers:
        - image: 821916800614.dkr.ecr.ap-southeast-1.amazonaws.com/ibridge-services:3.8.2
          imagePullPolicy: Always
          name: ces-orchestrator
          ports:
            - containerPort: 3005
          resources:
            limits:
              cpu: "200m"
              memory: "250Mi"
          env:
            - name: "ORCHESTRATORTOPIC"
              value: "orchestrator_rpc"
            - name: "CONSUMERTIMEOUT"
              value: "30"
          envFrom:
            - secretRef:
                name: ces-secret
            - configMapRef:
                name: ces-config
--- #Service ces-orchestrator-svc
apiVersion: v1
kind: Service
metadata:
  namespace: ces
  name: ces-orchestrator-svc
spec:
  ports:
    - port: 443
      targetPort: 3000
      protocol: TCP
  type: NodePort
  selector:
    app.kubernetes.io/name: ces-orchestrator
--- #ingress public-api
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: ces
  name: public-api
  annotations:
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: instance
    alb.ingress.kubernetes.io/target-node-labels: eks.amazonaws.com/nodegroup=ng-sbx
spec:
  ingressClassName: alb
  rules:
    - host: ces-new-api.soluix.ai
      http:
        paths:
          - path: /api/ibridge
            pathType: Prefix
            backend:
              service:
                name: ces-orchestrator-svc
                port:
                  number: 443
```

Below is environment variables that is used by CES Orchestrator:

| Variable             | Descriptions            | Default Value | Required |
| -------------        | ----------------------- | ------------- | -------- |
| POSTGRE_SQL_USER     | Postgresql DB Username  |               | true     |
| POSTGRE_SQL_PASSWORD | Postgresql DB Password  |               | true     |
| ...                  |                         |               |          |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Milestones

- [x] Listen REST API
- [x] RabbitMQ Consumer
- [x] RabbitMQ Producer RPC
- [ ] RabbitMQ Producer Event
- [x] Define Workflow in YAML
- [x] Dynamic routing
- [x] Monitoring resources
- [ ] Logging
- [ ] Workflow configuration audit trails
- [ ] Create workflow configuration with file
- [x] Readme file
- [ ] List of ibridge Helpers
- [ ] List of ibridge Operator
- [ ] ...
  
<p align="right">(<a href="#readme-top">back to top</a>)</p>