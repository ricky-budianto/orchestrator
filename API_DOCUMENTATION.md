# CES-iBridge Orchestrator Service API Documentation

This document provides comprehensive API specifications for all endpoints in the CES-iBridge Orchestrator Service. These endpoints were tested and documented to serve as reference for refactoring while maintaining request/response contracts.

## Base Configuration

- **Base URL**: `http://localhost:3000`
- **Service Path**: `/api/ibridge`
- **Content-Type**: `application/json`
- **Service Info**: Available at root path `/` returning service name and version

## API Versioning

### V1 APIs (`/api/ibridge/v1/`)
Traditional REST APIs requiring service restart for workflow configuration changes.

### V2 APIs (`/api/ibridge/v2/`)
Dynamic routing APIs supporting instant workflow activation without service restart.

---

## 1. Workflow Configuration Management

### 1.1 List Workflow Configurations
- **Endpoint**: `GET /api/ibridge/v1/workflow-configuration`
- **Description**: Retrieve all workflow configurations with pagination support
- **Query Parameters**: 
  - `limit` (optional): Number of records to return
  - `offset` (optional): Number of records to skip
  - `id` (optional): Filter by specific workflow ID

**Request Example:**
```bash
GET /api/ibridge/v1/workflow-configuration
```

**Response Schema:**
```json
{
  "success": true,
  "data": [
    {
      "id": "demo_flow",
      "name": "Demo Flow", 
      "path": "/api/ibridge/v1/demo-flow",
      "method": "POST",
      "descriptions": "Sample demo flow for testing the orchestrator",
      "configuration": "bmFtZTogZGVtb19mbG93...", // Base64 encoded YAML
      "revision_number": 1,
      "created_at": "2025-08-05T23:25:19.794917+07:00",
      "created_by_id": "system",
      "created_by_name": "System Setup",
      "updated_at": "0001-01-01T00:00:00Z",
      "update_by_id": "",
      "update_by_name": "",
      "deleted_at": null
    }
  ],
  "total_data": 3
}
```

### 1.2 Get Single Workflow Configuration
- **Endpoint**: `GET /api/ibridge/v1/workflow-configuration/{id}`
- **Description**: Retrieve a specific workflow configuration

**Request Example:**
```bash
GET /api/ibridge/v1/workflow-configuration/demo_flow
```

**Response Schema:**
```json
{
  "success": true,
  "data": {
    "id": "demo_flow",
    "name": "Demo Flow",
    "path": "/api/ibridge/v1/demo-flow", 
    "method": "POST",
    "descriptions": "Sample demo flow for testing the orchestrator",
    "configuration": "bmFtZTogZGVtb19mbG93...", // Base64 encoded YAML
    "revision_number": 1,
    "created_at": "2025-08-05T23:25:19.794917+07:00",
    "created_by_id": "system",
    "created_by_name": "System Setup", 
    "updated_at": "0001-01-01T00:00:00Z",
    "update_by_id": "",
    "update_by_name": "",
    "deleted_at": null
  }
}
```

### 1.3 Create Workflow Configuration
- **Endpoint**: `POST /api/ibridge/v1/workflow-configuration`
- **Description**: Create a new workflow configuration

**Request Schema:**
```json
{
  "id": "test_api_config",
  "name": "Test API Configuration", 
  "path": "/api/ibridge/v1/test-api-config",
  "method": "POST",
  "descriptions": "Test configuration created via API",
  "configuration": "bmFtZTogdGVzdF9hcGk..." // Base64 encoded YAML workflow definition
}
```

**Response Schema:**
```json
{
  "success": true,
  "data": {
    "id": "test_api_config",
    "name": "Test API Configuration",
    "path": "/api/ibridge/v1/test-api-config",
    "method": "POST", 
    "descriptions": "Test configuration created via API",
    "configuration": "bmFtZTogdGVzdF9hcGk...",
    "revision_number": 1,
    "created_at": "2025-08-15T00:17:18.388189+07:00",
    "created_by_id": "",
    "created_by_name": "",
    "updated_at": "2025-08-15T00:17:18.38819+07:00",
    "update_by_id": "",
    "update_by_name": "",
    "deleted_at": null
  },
  "error_code": ""
}
```

### 1.4 Update Workflow Configuration
- **Endpoint**: `PUT /api/ibridge/v1/workflow-configuration`
- **Description**: Update existing workflow configuration (increments revision_number)

**Request Schema:**
```json
{
  "id": "test_api_config",
  "name": "Test API Configuration Updated",
  "path": "/api/ibridge/v1/test-api-config", 
  "method": "POST",
  "descriptions": "Updated test configuration via API",
  "configuration": "bmFtZTogdGVzdF9hcGk..." // Updated Base64 encoded YAML
}
```

**Response Schema:**
```json
{
  "success": true,
  "data": {
    "id": "test_api_config",
    "name": "Test API Configuration Updated",
    "path": "/api/ibridge/v1/test-api-config",
    "method": "POST",
    "descriptions": "Updated test configuration via API", 
    "configuration": "bmFtZTogdGVzdF9hcGk...",
    "revision_number": 2, // Incremented
    "created_at": "2025-08-15T00:17:18.388189+07:00",
    "created_by_id": "",
    "created_by_name": "",
    "updated_at": "2025-08-15T00:17:38.342479+07:00", // Updated timestamp
    "update_by_id": "",
    "update_by_name": "",
    "deleted_at": null
  },
  "error_code": ""
}
```

### 1.5 Delete Workflow Configuration
- **Endpoint**: `DELETE /api/ibridge/v1/workflow-configuration/{id}`
- **Description**: Soft delete workflow configuration (sets deleted_at timestamp)

**Request Example:**
```bash
DELETE /api/ibridge/v1/workflow-configuration/test_api_config
```

**Response Schema:**
```json
{
  "success": true,
  "data": {
    "id": "test_api_config",
    "name": "",
    "path": "",
    "method": "",
    "descriptions": "",
    "configuration": "",
    "created_at": "0001-01-01T00:00:00Z",
    "created_by_id": "",
    "created_by_name": "",
    "updated_at": "0001-01-01T00:00:00Z",
    "update_by_id": "",
    "update_by_name": "",
    "deleted_at": "2025-08-15T00:18:25.820894+07:00" // Soft delete timestamp
  }
}
```

---

## 2. Workflow Execution APIs

### 2.1 V2 Dynamic Routing Execution
- **Endpoint**: `POST /api/ibridge/v2/orchestrate/{workflow_configuration_id}`
- **Description**: Execute workflow using dynamic routing (instant activation)
- **Supported Methods**: `POST`, `GET`, `PUT`
- **Optional Parameters**: Additional path parameters via `/{workflow_configuration_id}/{param_id}`

**Request Example:**
```bash
POST /api/ibridge/v2/orchestrate/simple_demo_flow
Authorization: Bearer test-token
Content-Type: application/json

{
  "test": "data",
  "user": "test_user"
}
```

**Response Schema:**
```json
{
  "data": {
    "data": {
      "correlation_id": "5f4388b4-5a9c-43e5-b7d3-b23ff5f7e19d",
      "input": {
        "test": "data", 
        "user": "test_user"
      },
      "timestamp": "",
      "workflow_id": "simple_demo_flow"
    },
    "message": "Success",
    "status_code": 200
  },
  "message": "Success",
  "response_code": "S1",
  "success": true
}
```

### 2.2 V1 Static Routing Execution
- **Endpoint**: Dynamic based on workflow configuration path/method
- **Description**: Execute workflows via predefined paths (requires service restart for new workflows)
- **Note**: These endpoints are dynamically registered based on workflow configurations

**Example (if workflow exists):**
```bash
POST /api/ibridge/v1/demo-flow
Authorization: Bearer test-token
Content-Type: application/json

{
  "email": "test@example.com",
  "penghasilan": 5000000,
  "nik": "1234567890123456", 
  "social_media_url": "https://facebook.com/testuser"
}
```

---

## 3. Workflow History Management

### 3.1 List Workflow Histories
- **Endpoint**: `GET /api/ibridge/v1/workflow-history`
- **Description**: Retrieve workflow execution history with audit trail

**Request Example:**
```bash
GET /api/ibridge/v1/workflow-history
```

**Response Schema:**
```json
{
  "success": true,
  "data": [
    {
      "id": "65557b16-7341-48cd-af26-0b0cb249dee3",
      "workflow_configuration_id": "demo_flow",
      "status": "EXECUTED",
      "request": {
        "Authorization": "Bearer test-token",
        "Body": {
          "email": "test@example.com",
          "nik": "1234567890123456",
          "penghasilan": 5000000,
          "social_media_url": "https://facebook.com/testuser"
        },
        "BodyRaw": "ewogICAgImVtYWlsIjog...", // Base64 encoded request body
        "Context": {},
        "CorrelationID": "65557b16-7341-48cd-af26-0b0cb249dee3",
        "Header": {
          "Accept": "*/*",
          "Authorization": "Bearer test-token",
          "Content-Length": "153",
          "Content-Type": "application/json",
          "Host": "localhost:3000",
          "User-Agent": "curl/8.4.0"
        },
        "Headers": { /* Same as Header */ },
        "OrchestrationQueryParams": null,
        "Params": {},
        "Path": "/api/ibridge/v1/demo-flow",
        "QueryParams": {},
        "RequestType": "demo_flow"
      },
      "response": {
        "ResponseData": {
          "data": null,
          "error_code": "1006",
          "message": "Connection Timeout",
          "response_code": "1006",
          "success": false
        },
        "StatusCode": 500
      },
      "additional_info": null,
      "revision_number": 1,
      "created_at": "2025-08-05T23:34:36.149825+07:00",
      "created_by_id": "SYSTEM",
      "created_by_name": "",
      "updated_at": "2025-08-05T23:36:06.152243+07:00",
      "update_by_id": "SYSTEM", 
      "update_by_name": "",
      "deleted_at": null
    }
  ]
}
```

### 3.2 Get Single Workflow History
- **Endpoint**: `GET /api/ibridge/v1/workflow-history/{id}`
- **Description**: Retrieve specific workflow execution history

### 3.3 Create Workflow History (Internal)
- **Endpoint**: `POST /api/ibridge/v1/workflow-history`
- **Description**: Internal endpoint for creating workflow history records

### 3.4 Update Workflow History (Internal) 
- **Endpoint**: `PUT /api/ibridge/v1/workflow-history`
- **Description**: Internal endpoint for updating workflow history records

---

## 4. Workflow State Management

### 4.1 List Workflow States
- **Endpoint**: `GET /api/ibridge/v1/workflow-state`
- **Description**: Retrieve detailed workflow execution states for debugging

**Request Example:**
```bash
GET /api/ibridge/v1/workflow-state
```

**Response Schema:**
```json
{
  "success": true,
  "data": [
    {
      "id": "4c9a76f5-8768-403f-92d1-f71cd356d4e4",
      "workflow_id": "44b038ba-61bf-4fa8-b3bc-3777daaff5fc",
      "workflow_request_type": "demo_flow",
      "state": {
        "demo_check_email": {
          "additional_info": {
            "request_data": "{\"Body\":{\"email\":\"test@example.com\"},...}",
            "response_data": "{\"success\":false,\"data\":null,...}"
          },
          "message": "rabbit mq timeout",
          "request_sent": "2025-08-05T23:39:19.405734+07:00",
          "response_received": "2025-08-05T23:40:19.433665+07:00",
          "success": false
        },
        "demo_check_penghasilan": {
          /* Similar structure for each workflow step */
        },
        "demo_check_nik_socmed": {
          /* Similar structure for each workflow step */
        },
        "demo_get_limit": {
          /* Similar structure for each workflow step */
        }
      },
      "revision_number": 1,
      "created_at": "2025-08-05T23:40:19.433665+07:00",
      "created_by_id": "SYSTEM",
      "created_by_name": "",
      "updated_at": "2025-08-05T23:43:19.52863+07:00",
      "update_by_id": "SYSTEM",
      "update_by_name": "",
      "deleted_at": null
    }
  ]
}
```

### 4.2 Get Single Workflow State
- **Endpoint**: `GET /api/ibridge/v1/workflow-state/{id}`
- **Description**: Retrieve specific workflow state details

### 4.3 Create Workflow State (Internal)
- **Endpoint**: `POST /api/ibridge/v1/workflow-state`
- **Description**: Internal endpoint for creating workflow state records

### 4.4 Update Workflow State (Internal)
- **Endpoint**: `PUT /api/ibridge/v1/workflow-state`
- **Description**: Internal endpoint for updating workflow state records

---

## 5. Workflow Audit Management

### 5.1 List Workflow Audits
- **Endpoint**: `GET /api/ibridge/v1/workflow-audit`
- **Description**: Retrieve workflow audit records

### 5.2 Get Single Workflow Audit
- **Endpoint**: `GET /api/ibridge/v1/workflow-audit/{id}`
- **Description**: Retrieve specific workflow audit record

### 5.3 Create Workflow Audit
- **Endpoint**: `POST /api/ibridge/v1/workflow-audit`
- **Description**: Create new workflow audit record

### 5.4 Update Workflow Audit
- **Endpoint**: `PUT /api/ibridge/v1/workflow-audit`  
- **Description**: Update existing workflow audit record

---

## 6. Logging and Monitoring

### 6.1 Elasticsearch Logs
- **Endpoint**: `GET /api/ibridge/v1/logs`
- **Description**: Retrieve application logs from Elasticsearch
- **Query Parameters**: Various filtering options

**Note**: This endpoint requires proper Elasticsearch configuration and may return errors if not configured.

**Error Response Example:**
```json
{
  "success": false,
  "message": "undefined index", 
  "error_code": "1025"
}
```

---

## 7. RabbitMQ RPC Interface

### 7.1 RPC Message Format
The service consumes RabbitMQ messages for workflow execution through RPC pattern.

**RPC Message Schema:**
```json
{
  "request_id": "12345",
  "correlation_id": "123456", 
  "request_type": "demo_flow",
  "content_type": "application/json",
  "header": {},
  "key": "workflow_key",
  "value": {
    "email": "test@gmail.com",
    "penghasilan": "30000000", 
    "nik": "320124019281871891",
    "social_media_url": "google.com"
  },
  "need_response": true
}
```

**Legacy RPC Message Format:**
```json
{
  "key": "workflow_key",
  "value": {
    "data": "base64_encoded_payload",
    "request_type": "demo_flow"
  }
}
```

### 7.2 RPC Response Format
```json
{
  "success": true,
  "data": {
    /* Workflow execution result */
  },
  "message": "Success",
  "response_code": "S1",
  "error_code": ""
}
```

---

## Common Response Patterns

### Success Response
```json
{
  "success": true,
  "data": { /* Response data */ },
  "message": "Success",
  "response_code": "S1", 
  "error_code": ""
}
```

### Error Response
```json
{
  "success": false,
  "data": null,
  "message": "Error description",
  "response_code": "1XXX",
  "error_code": "1XXX"
}
```

### Common Error Codes
- `1006`: Connection Timeout
- `1025`: Bad Request / Undefined Index
- `1001`: Unknown Error
- `1002`: System Error
- `1003`: Database Error
- `1005`: Third Party System Error

---

## Headers and Authentication

### Standard Headers
- `Content-Type: application/json`
- `Authorization: Bearer {token}` (optional but recommended)
- `X-Request-Id: {correlation_id}` (auto-generated if not provided)

### Activity Tracking
All endpoints include activity tracking middleware that logs request details when access tokens are provided.

---

## Testing Notes

1. **Service Startup**: Requires PostgreSQL, Redis, and RabbitMQ connections
2. **Dynamic Workflows**: V2 endpoints work immediately after workflow configuration creation
3. **Static Workflows**: V1 endpoints require service restart for new configurations  
4. **Worker Integration**: Some workflows depend on external worker services via RabbitMQ
5. **Audit Trail**: All executions are logged to workflow_history and workflow_state tables

This documentation captures the complete API surface as of the current implementation and can be used as a reference during refactoring to ensure request/response contracts remain unchanged.