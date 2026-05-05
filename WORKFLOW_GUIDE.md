# Workflow Configuration Guide

Complete reference for creating and managing workflows in CES-iBridge Orchestrator Service.

## Table of Contents

1. [Workflow Anatomy](#workflow-anatomy)
2. [Variable Substitution](#variable-substitution)
3. [Task Types](#task-types)
4. [Conditional Logic](#conditional-logic)
5. [Global Variables](#global-variables)
6. [Response Formatting](#response-formatting)
7. [Advanced Patterns](#advanced-patterns)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

---

## Workflow Anatomy

### Basic Structure

```yaml
name: workflow_id                    # Unique identifier (must match DB id)
requestType: request_type_name       # For RabbitMQ routing
path: /api/ibridge/v1/endpoint       # REST endpoint (DB overrides this)
method: POST                         # HTTP method (DB overrides this)
startEvent:
  targetRef: first_task_name         # Entry point task

serviceTask:
  task_name_1:
    type: worker|conditional|endEvent
    # ... task configuration

  task_name_2:
    # ... more tasks

  ibridge-reply:
    type: endEvent
    # ... end event configuration
```

### Header Fields

| Field | Description | Required | Notes |
|-------|-------------|----------|-------|
| `name` | Workflow identifier | ✅ Yes | Must match database `id` field |
| `requestType` | RabbitMQ routing key | ✅ Yes | Used for RabbitMQ message routing |
| `path` | REST API endpoint | ⚠️ Optional | Database value takes precedence |
| `method` | HTTP method | ⚠️ Optional | Database value takes precedence |
| `startEvent` | Entry point | ✅ Yes | Must reference a valid task |
| `serviceTask` | Task definitions | ✅ Yes | Contains all workflow tasks |

### Important Notes

- **Never use `/v2` in path**: This is reserved for dynamic routing
- **Database overrides YAML**: `path` and `method` in database take precedence
- **Unique constraint**: Optional field for rate limiting/deduplication
- **Task names**: Must be unique within the workflow

---

## Variable Substitution

### Variable Syntax

Variables use the format: `${{scope.path.to.value}}`

### Available Scopes

#### 1. `startEvent` - Initial Request Data

Access incoming request data:

```yaml
${{startEvent.Body.field_name}}           # Request body field
${{startEvent.Header.Authorization}}      # HTTP header
${{startEvent.Header.X-User-Id}}         # Custom header
${{startEvent.Params.id}}                # URL parameters
${{startEvent.QueryParams.filter}}       # Query string parameters
${{startEvent.Authorization}}            # Authorization header (shortcut)
```

**Example:**
```yaml
sourceRef:
  authorization: ${{startEvent.Authorization}}
  body:
    user_email: ${{startEvent.Body.email}}
    user_id: ${{startEvent.Params.userId}}
```

#### 2. `worker_name` - Previous Task Results

Access data from any previous worker task:

```yaml
${{worker_name.data}}                    # Full worker response
${{worker_name.data.field}}              # Specific field
${{worker_name.data.nested.field}}       # Nested field
${{worker_name.data.array[0]}}           # Array element (if supported)
```

**Example:**
```yaml
# After running 'check_email' worker
sourceRef:
  body:
    email_score: ${{check_email.data.data.score}}
    is_disposable: ${{check_email.data.data.disposable_email}}
    verification_id: ${{check_email.data.id}}
```

#### 3. `ibridge` - Special System Variables

Built-in system utilities:

```yaml
${{ibridge.timestamp}}                   # Current timestamp
${{ibridge.global.field_name}}           # Global variable value
${{ibridge.httprequest}}                 # Special HTTP request format
```

**Example:**
```yaml
sourceRef:
  body:
    created_at: ${{ibridge.timestamp}}
    error_message: ${{ibridge.global.message}}
    error_code: ${{ibridge.global.error_code}}
```

#### 4. Global Variables

Access saved global state:

```yaml
${{ibridge.global.status_code}}
${{ibridge.global.error_message}}
${{ibridge.global.custom_field}}
```

### sourceRefParsing

Controls how worker request is formatted:

```yaml
# Option 1: HTTP Request Format (most common)
sourceRefParsing: ${{ibridge.httprequest}}
sourceRef:
  authorization: ${{startEvent.Authorization}}
  body:
    field1: value1
    field2: ${{previous_task.data.field}}

# Option 2: Raw Format (direct pass-through)
sourceRefParsing: raw
sourceRef:
  any_structure: here
  is_passed: directly
```

**HTTP Request Format Structure:**
```json
{
  "authorization": "Bearer token...",
  "body": {
    "field1": "value1",
    "field2": "value2"
  },
  "header": {
    "Custom-Header": "value"
  }
}
```

---

## Task Types

### 1. Worker Tasks

Execute external service calls or internal processing.

**Structure:**
```yaml
task_name:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      field1: ${{startEvent.Body.field1}}
  targetRef:
    - next_task_name
  worker: middleware
  save_global:                          # Optional
    custom_field: ${{task_name.data.value}}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✅ Yes | Must be `worker` |
| `sourceRefParsing` | string | ⚠️ Recommended | Format type: `${{ibridge.httprequest}}` or `raw` |
| `sourceRef` | object | ✅ Yes | Data to send to worker |
| `targetRef` | array | ✅ Yes | Next task(s) to execute |
| `worker` | string | ✅ Yes | Worker identifier (e.g., `middleware`) |
| `save_global` | object | ❌ No | Save values to global state |

**Worker Types:**

- **`middleware`**: RabbitMQ RPC call to middleware service
- Custom worker names map to specific RabbitMQ queues

**Example - API Call Worker:**
```yaml
validate_email:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      email: ${{startEvent.Body.email}}
      check_disposable: true
      check_domain: true
  targetRef:
    - check_email_result
  worker: middleware
  save_global:
    email_trust_score: ${{validate_email.data.data.trust_score}}
```

### 2. Conditional Tasks

Decision points based on previous results.

**Structure:**
```yaml
task_name:
  type: conditional
  condition:
    variableCheck:
      - ${{worker.data.field1}}
      - ${{worker.data.field2}}
    value:
      - case:
          - value1
          - value2
        target: target_task_if_match
        status_code: 400              # Optional
        save_global:                  # Optional
          error_code: "ERR_001"
      - case:
          - value3
        target: another_task
  defaultTasK: default_task_name      # Note: Capital 'K' is required
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✅ Yes | Must be `conditional` |
| `variableCheck` | array | ✅ Yes | Variables to evaluate |
| `value` | array | ✅ Yes | Case conditions |
| `defaultTasK` | string | ✅ Yes | Fallback task (note capital K) |

**Condition Matching:**

- **Multiple variables**: ALL must match their respective case values
- **Case array**: Values are matched by index with variableCheck array
- **First match wins**: Evaluates cases in order, executes first match

**Example - Simple Boolean Check:**
```yaml
check_email_validity:
  type: conditional
  condition:
    variableCheck:
      - ${{validate_email.data.data.is_valid}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          message: "Invalid email address"
          error_code: "EMAIL_001"
  defaultTasK: proceed_to_next_step
```

**Example - Multiple Conditions:**
```yaml
check_risk_and_score:
  type: conditional
  condition:
    variableCheck:
      - ${{risk_assessment.data.data.risk_level}}
      - ${{credit_score.data.data.score}}
    value:
      - case:                           # HIGH risk AND score < 500
          - "HIGH"
          - 400
        target: reject_application
        save_global:
          rejection_reason: "High risk and low score"
      - case:                           # MEDIUM risk (any score)
          - "MEDIUM"
        target: manual_review
  defaultTasK: approve_application
```

**Example - Multi-Case Switch:**
```yaml
route_by_status:
  type: conditional
  condition:
    variableCheck:
      - ${{process_result.data.data.status}}
    value:
      - case:
          - "APPROVED"
        target: send_approval_notification
      - case:
          - "REJECTED"
        target: send_rejection_notification
      - case:
          - "PENDING"
        target: send_pending_notification
  defaultTasK: handle_unknown_status
```

### 3. End Event Tasks

Terminal task that formats the final response.

**Structure:**
```yaml
ibridge-reply:
  type: endEvent
  taskRef: ibridge-reply              # Must match task name
  endParsing: Separated|Combined|raw
  sourceRef:
    field1: ${{worker1.data}}
    field2: ${{worker2.data.field}}
    error_code: ${{ibridge.global.error_code}}
  status_code: 200                    # Optional, defaults from save_global
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✅ Yes | Must be `endEvent` |
| `taskRef` | string | ✅ Yes | Must match task name (usually `ibridge-reply`) |
| `endParsing` | string | ⚠️ Recommended | Response format type |
| `sourceRef` | object | ✅ Yes | Response data structure |
| `status_code` | integer | ❌ No | HTTP status code override |

**endParsing Options:**

#### `Separated` (Recommended for Complex Workflows)

Returns structured response with all fields preserved:

```yaml
endParsing: Separated
sourceRef:
  user_data: ${{create_user.data}}
  verification: ${{verify_email.data}}
  metadata:
    workflow_id: ${{ibridge.global.workflow_id}}
```

**Output:**
```json
{
  "user_data": { "id": 123, "name": "John" },
  "verification": { "verified": true },
  "metadata": { "workflow_id": "abc-123" }
}
```

#### `Combined` (Legacy/Simple Responses)

Merges all data into single object:

```yaml
endParsing: Combined
sourceRef:
  id: ${{create_user.data.id}}
  name: ${{create_user.data.name}}
  verified: ${{verify_email.data.verified}}
```

#### `raw` (Direct Pass-Through)

Returns exact structure without processing:

```yaml
endParsing: raw
sourceRef:
  ${{final_worker.data}}
```

---

## Conditional Logic

### Comparison Types

The conditional system supports exact value matching:

```yaml
variableCheck:
  - ${{worker.data.field}}
value:
  - case:
      - "expected_value"      # String match
      - true                  # Boolean match
      - 123                   # Number match
      - null                  # Null check
```

### Complex Conditional Patterns

#### Pattern 1: Early Exit on Error

```yaml
validate_input:
  type: worker
  # ... worker config ...
  targetRef:
    - check_validation_result

check_validation_result:
  type: conditional
  condition:
    variableCheck:
      - ${{validate_input.data.data.valid}}
    value:
      - case:
          - false
        target: ibridge-reply         # Exit workflow
        status_code: 400
        save_global:
          status_code: 400
          message: "Validation failed"
          error_code: "VAL_001"
          errors: ${{validate_input.data.data.errors}}
  defaultTasK: continue_processing    # Happy path
```

#### Pattern 2: Risk-Based Routing

```yaml
assess_risk:
  type: conditional
  condition:
    variableCheck:
      - ${{risk_check.data.data.risk_level}}
    value:
      - case:
          - "HIGH"
        target: ibridge-reply
        status_code: 403
        save_global:
          status_code: 403
          message: "High risk detected"
      - case:
          - "MEDIUM"
        target: require_manual_review
        save_global:
          review_required: true
      - case:
          - "LOW"
        target: auto_approve
  defaultTasK: handle_unknown_risk
```

#### Pattern 3: Multi-Field Validation

```yaml
check_eligibility:
  type: conditional
  condition:
    variableCheck:
      - ${{age_check.data.data.eligible}}
      - ${{income_check.data.data.sufficient}}
      - ${{credit_check.data.data.approved}}
    value:
      - case:                         # All three must be true
          - true
          - true
          - true
        target: approve_application
      - case:                         # Age not eligible
          - false
        target: reject_age
        save_global:
          rejection_reason: "Age requirement not met"
  defaultTasK: reject_other_reasons   # Any other combination
```

#### Pattern 4: Disposable Email Detection

```yaml
check_email_type:
  type: conditional
  condition:
    variableCheck:
      - ${{email_validation.data.data.disposable_email}}
      - ${{email_validation.data.data.valid_domain}}
    value:
      - case:
          - true              # disposable = true
          - false             # valid_domain = false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          message: "Invalid email: disposable or invalid domain"
  defaultTasK: accept_email
```

---

## Global Variables

### Saving Global Variables

Use `save_global` in workers or conditionals to store values accessible throughout the workflow:

```yaml
# In a worker task
check_user:
  type: worker
  # ... config ...
  save_global:
    user_id: ${{check_user.data.data.id}}
    user_tier: ${{check_user.data.data.tier}}
    timestamp: ${{ibridge.timestamp}}

# In a conditional task
validate_result:
  type: conditional
  condition:
    # ... condition config ...
    value:
      - case:
          - false
        target: ibridge-reply
        save_global:
          status_code: 400
          message: "Validation failed"
          error_code: "ERR_001"
          error_details: ${{validate.data.errors}}
  defaultTasK: next_step
```

### Accessing Global Variables

Reference saved globals using `ibridge.global`:

```yaml
# In subsequent worker
send_notification:
  type: worker
  sourceRef:
    body:
      user_id: ${{ibridge.global.user_id}}
      tier: ${{ibridge.global.user_tier}}
      error_occurred: ${{ibridge.global.error_code}}

# In end event
ibridge-reply:
  type: endEvent
  sourceRef:
    status: ${{ibridge.global.status_code}}
    message: ${{ibridge.global.message}}
    error_code: ${{ibridge.global.error_code}}
```

### Status Code Handling

The `status_code` global variable has special behavior:

```yaml
# Set in conditional
some_check:
  type: conditional
  condition:
    # ... config ...
    value:
      - case:
          - error_condition
        status_code: 422          # Sets HTTP status
        save_global:
          status_code: 422        # Also saves to global
          message: "Error occurred"

# End event uses global status_code if not explicitly set
ibridge-reply:
  type: endEvent
  # status_code: 200            # Optional override
  sourceRef:
    # ... response data ...
  # If no status_code specified, uses ${{ibridge.global.status_code}}
```

### Global Variable Scope

- **Workflow-scoped**: Available only within current workflow execution
- **Read-write**: Can be updated multiple times
- **Persistent**: Maintained throughout entire workflow execution
- **Not persisted**: Lost after workflow completes

---

## Response Formatting

### End Event Configuration

```yaml
ibridge-reply:
  type: endEvent
  taskRef: ibridge-reply
  endParsing: Separated|Combined|raw
  sourceRef:
    # Your response structure
  status_code: 200                    # Optional
```

### Response Examples

#### Success Response (Separated)

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    status: "success"
    user_profile:
      id: ${{create_user.data.data.id}}
      email: ${{create_user.data.data.email}}
      verified: true
    verification_details:
      email_check: ${{check_email.data}}
      phone_check: ${{check_phone.data}}
    timestamp: ${{ibridge.timestamp}}
```

**HTTP Response:**
```json
{
  "status": "success",
  "user_profile": {
    "id": 12345,
    "email": "user@example.com",
    "verified": true
  },
  "verification_details": {
    "email_check": { "valid": true, "score": 95 },
    "phone_check": { "verified": true }
  },
  "timestamp": "2025-10-05T10:30:00Z"
}
```
**Status Code:** 200 (default or from save_global)

#### Error Response (Separated)

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    status: "error"
    error_code: ${{ibridge.global.error_code}}
    message: ${{ibridge.global.message}}
    details: ${{ibridge.global.error_details}}
    timestamp: ${{ibridge.timestamp}}
  # status_code taken from save_global (e.g., 400, 422, 403)
```

**HTTP Response:**
```json
{
  "status": "error",
  "error_code": "VAL_001",
  "message": "Validation failed",
  "details": {
    "field": "email",
    "reason": "Invalid format"
  },
  "timestamp": "2025-10-05T10:30:00Z"
}
```
**Status Code:** 400 (from save_global)

#### Partial Success (Separated)

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    status: "partial_success"
    user_created: ${{create_user.data.success}}
    notification_sent: ${{send_notification.data.success}}
    user_data: ${{create_user.data.data}}
    errors:
      notification_error: ${{send_notification.data.error}}
```

---

## Advanced Patterns

### Pattern 1: Multi-Step Validation Pipeline

```yaml
name: validation_pipeline
startEvent:
  targetRef: validate_format

serviceTask:
  validate_format:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body: ${{startEvent.Body}}
    targetRef:
      - check_format_result
    worker: middleware

  check_format_result:
    type: conditional
    condition:
      variableCheck:
        - ${{validate_format.data.data.valid}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 400
          save_global:
            status_code: 400
            error_code: "FORMAT_ERROR"
            message: ${{validate_format.data.data.error}}
    defaultTasK: validate_business_rules

  validate_business_rules:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        data: ${{validate_format.data.data.normalized}}
        rules: ["rule1", "rule2"]
    targetRef:
      - check_business_rules
    worker: middleware

  check_business_rules:
    type: conditional
    condition:
      variableCheck:
        - ${{validate_business_rules.data.data.passed}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 422
          save_global:
            status_code: 422
            error_code: "BUSINESS_RULE_VIOLATION"
            message: ${{validate_business_rules.data.data.violation}}
    defaultTasK: process_data

  process_data:
    type: worker
    # ... processing logic ...
    targetRef:
      - ibridge-reply
    worker: middleware

  ibridge-reply:
    endParsing: Separated
    sourceRef:
      success: true
      validation:
        format_check: ${{validate_format.data}}
        business_rules: ${{validate_business_rules.data}}
      result: ${{process_data.data}}
      errors:
        code: ${{ibridge.global.error_code}}
        message: ${{ibridge.global.message}}
    type: endEvent
    taskRef: ibridge-reply
```

### Pattern 2: Parallel Data Enrichment

```yaml
name: data_enrichment
startEvent:
  targetRef: fetch_user_basic

serviceTask:
  fetch_user_basic:
    type: worker
    # ... fetch user ...
    targetRef:
      - enrich_with_preferences
    worker: middleware
    save_global:
      user_id: ${{fetch_user_basic.data.data.id}}

  enrich_with_preferences:
    type: worker
    sourceRef:
      body:
        user_id: ${{ibridge.global.user_id}}
    targetRef:
      - enrich_with_activity
    worker: middleware

  enrich_with_activity:
    type: worker
    sourceRef:
      body:
        user_id: ${{ibridge.global.user_id}}
    targetRef:
      - enrich_with_analytics
    worker: middleware

  enrich_with_analytics:
    type: worker
    sourceRef:
      body:
        user_id: ${{ibridge.global.user_id}}
    targetRef:
      - merge_all_data
    worker: middleware

  merge_all_data:
    type: worker
    sourceRef:
      body:
        basic: ${{fetch_user_basic.data}}
        preferences: ${{enrich_with_preferences.data}}
        activity: ${{enrich_with_activity.data}}
        analytics: ${{enrich_with_analytics.data}}
    targetRef:
      - ibridge-reply
    worker: middleware

  ibridge-reply:
    endParsing: Separated
    sourceRef:
      user_profile: ${{merge_all_data.data}}
    type: endEvent
    taskRef: ibridge-reply
```

### Pattern 3: Retry with Fallback

```yaml
name: service_with_fallback
startEvent:
  targetRef: call_primary_service

serviceTask:
  call_primary_service:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      body: ${{startEvent.Body}}
    targetRef:
      - check_primary_result
    worker: middleware

  check_primary_result:
    type: conditional
    condition:
      variableCheck:
        - ${{call_primary_service.data.success}}
      value:
        - case:
            - false
          target: call_fallback_service
          save_global:
            primary_failed: true
            primary_error: ${{call_primary_service.data.error}}
    defaultTasK: ibridge-reply

  call_fallback_service:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      body: ${{startEvent.Body}}
      metadata:
        reason: "primary_service_failed"
        primary_error: ${{ibridge.global.primary_error}}
    targetRef:
      - check_fallback_result
    worker: middleware

  check_fallback_result:
    type: conditional
    condition:
      variableCheck:
        - ${{call_fallback_service.data.success}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 503
          save_global:
            status_code: 503
            error_code: "SERVICE_UNAVAILABLE"
            message: "Both primary and fallback services failed"
    defaultTasK: ibridge-reply

  ibridge-reply:
    endParsing: Separated
    sourceRef:
      result: ${{call_primary_service.data}}
      fallback_used: ${{ibridge.global.primary_failed}}
      fallback_result: ${{call_fallback_service.data}}
      error:
        code: ${{ibridge.global.error_code}}
        message: ${{ibridge.global.message}}
    type: endEvent
    taskRef: ibridge-reply
```

---

## Best Practices

### 1. Naming Conventions

```yaml
# Use descriptive, hierarchical names
validate_email_format          # Good
check_user_eligibility         # Good
assess_risk_profile           # Good

email_check                    # Less clear
validate                       # Too generic
x1                            # Bad
```

### 2. Error Handling

**Always handle error cases explicitly:**

```yaml
some_worker:
  type: worker
  # ... config ...
  targetRef:
    - check_worker_result       # Always add conditional check

check_worker_result:
  type: conditional
  condition:
    variableCheck:
      - ${{some_worker.data.success}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 500
        save_global:
          status_code: 500
          error_code: "WORKER_FAILED"
          message: ${{some_worker.data.error}}
  defaultTasK: continue_workflow
```

### 3. Use Global Variables for Cross-Task Data

```yaml
# Set once, use everywhere
identify_user:
  type: worker
  # ... config ...
  save_global:
    user_id: ${{identify_user.data.data.id}}
    user_tier: ${{identify_user.data.data.tier}}
    user_country: ${{identify_user.data.data.country}}

# Reference in multiple subsequent tasks
task_2:
  sourceRef:
    body:
      user_id: ${{ibridge.global.user_id}}

task_3:
  sourceRef:
    body:
      user_id: ${{ibridge.global.user_id}}
      tier: ${{ibridge.global.user_tier}}
```

### 4. Structured Error Responses

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    # Success fields (may be null on error)
    result: ${{process_data.data}}

    # Always include error structure
    error:
      occurred: ${{ibridge.global.status_code}}  # truthy if error
      code: ${{ibridge.global.error_code}}
      message: ${{ibridge.global.message}}
      details: ${{ibridge.global.error_details}}

    # Metadata
    timestamp: ${{ibridge.timestamp}}
    workflow_id: ${{ibridge.global.workflow_id}}
```

### 5. Conditional Defaults

**Always provide `defaultTasK`:**

```yaml
some_conditional:
  type: conditional
  condition:
    # ... conditions ...
  defaultTasK: handle_default_case  # Required, not optional
```

### 6. Variable Existence Checks

When referencing optional fields, use conditionals to verify existence:

```yaml
check_optional_field:
  type: conditional
  condition:
    variableCheck:
      - ${{previous_task.data.optional_field}}
    value:
      - case:
          - null
        target: handle_missing_field
  defaultTasK: process_with_field
```

### 7. Keep Tasks Focused

Each task should have a single responsibility:

```yaml
# Good: Separate concerns
validate_email:
  type: worker
  # ... validates email only ...

validate_phone:
  type: worker
  # ... validates phone only ...

# Avoid: Kitchen sink tasks
validate_everything:
  type: worker
  # ... validates email, phone, address, documents ...
```

---

## Troubleshooting

### Common Issues

#### 1. Variable Not Resolving

**Problem:** `${{worker.data.field}}` returns null or empty

**Solutions:**
- Check worker response structure in `workflow_states` table
- Verify task name matches exactly (case-sensitive)
- Ensure worker completed before conditional references it
- Use `workflow_histories` to inspect actual worker responses

**Debug Example:**
```sql
-- Check what data worker actually returned
SELECT state FROM workflow_states
WHERE workflow_id = 'your_workflow_history_id';
```

#### 2. Conditional Not Matching

**Problem:** Conditional always takes defaultTasK path

**Solutions:**
- Verify exact value match (type-sensitive: `"true"` ≠ `true`)
- Check variable path is correct
- Ensure case array length matches variableCheck array length
- Log variable values using save_global for debugging

**Debug Pattern:**
```yaml
debug_conditional:
  type: conditional
  condition:
    variableCheck:
      - ${{worker.data.field}}
    value:
      - case:
          - expected_value
        target: matched
        save_global:
          debug_value: ${{worker.data.field}}  # Log actual value
  defaultTasK: not_matched
```

#### 3. Workflow Not Loading

**Problem:** Workflow not available via V1 or V2 API

**Solutions:**
- Verify `workflow_configurations` table has entry
- Check `path` and `method` in database (these override YAML)
- For V1: Restart service to load new routes
- For V2: Use workflow `id` field, not path
- Verify YAML is valid base64 in `configuration` column

**Check Database:**
```sql
SELECT id, name, path, method, status
FROM workflow_configurations
WHERE id = 'your_workflow_id';
```

#### 4. Worker Timeout

**Problem:** Workflow hangs or times out

**Solutions:**
- Check RabbitMQ worker is running and consuming
- Verify `RABBIT_MQ_TIMEOUT` environment variable
- Check worker queue name matches `worker` field
- Monitor RabbitMQ for unconsumed messages

#### 5. Status Code Not Applied

**Problem:** Wrong HTTP status code returned

**Solutions:**
- Check `save_global.status_code` is set in conditional
- Verify `status_code` field in conditional case block
- End event `status_code` can override global
- Default is 200 if not set anywhere

**Priority Order:**
1. End event `status_code` field (highest)
2. `${{ibridge.global.status_code}}` from save_global
3. Default 200 (lowest)

#### 6. endParsing Issues

**Problem:** Response format not as expected

**Solutions:**
- Use `Separated` for most cases (clearest structure)
- Use `Combined` only for simple, flat responses
- Avoid `raw` unless you need exact pass-through
- Check actual response in `workflow_histories.response`

---

## Quick Reference

### Task Type Cheat Sheet

```yaml
# Worker
task_name:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef: { ... }
  targetRef: [next_task]
  worker: middleware
  save_global: { field: value }  # optional

# Conditional
task_name:
  type: conditional
  condition:
    variableCheck: [var1, var2]
    value:
      - case: [val1, val2]
        target: task_if_match
        status_code: 400  # optional
        save_global: {}   # optional
  defaultTasK: default_task

# End Event
ibridge-reply:
  type: endEvent
  taskRef: ibridge-reply
  endParsing: Separated
  sourceRef: { ... }
  status_code: 200  # optional
```

### Variable Reference Cheat Sheet

```yaml
# Request Data
${{startEvent.Body.field}}
${{startEvent.Header.HeaderName}}
${{startEvent.Params.param}}
${{startEvent.QueryParams.query}}
${{startEvent.Authorization}}

# Worker Results
${{worker_name.data.field}}
${{worker_name.data.nested.field}}

# Global Variables
${{ibridge.global.field_name}}
${{ibridge.global.status_code}}
${{ibridge.timestamp}}
${{ibridge.httprequest}}  # sourceRefParsing only
```

### Common Patterns

```yaml
# Early exit on error
check_error:
  type: conditional
  condition:
    variableCheck: [${{task.data.success}}]
    value:
      - case: [false]
        target: ibridge-reply
        status_code: 400
        save_global:
          status_code: 400
          error_code: "ERR_CODE"
  defaultTasK: continue

# Save for later use
some_task:
  save_global:
    key: ${{some_task.data.value}}

# Multi-case switch
route_task:
  type: conditional
  condition:
    variableCheck: [${{task.data.type}}]
    value:
      - case: ["TYPE_A"]
        target: handle_type_a
      - case: ["TYPE_B"]
        target: handle_type_b
  defaultTasK: handle_default
```

---

## Workflow Code Smells & Refactoring

### Anti-Patterns to Avoid

#### 🚫 Anti-Pattern 1: God Workflow

**Smell:** Single workflow does everything (50+ tasks, multiple unrelated domains).

```yaml
# BAD: One workflow for user registration, verification, payment, and analytics
name: do_everything_workflow
serviceTask:
  validate_user:
    # ...
  verify_email:
    # ...
  verify_phone:
    # ...
  process_payment:
    # ...
  calculate_analytics:
    # ...
  send_notifications:
    # ...
  # ... 45 more tasks
```

**Problems:**
- Hard to debug (which of 50 tasks failed?)
- Impossible to reuse
- Long execution time
- High failure rate
- Difficult to maintain

**Refactoring:**

✅ **Split into focused workflows:**

```yaml
# GOOD: Separate workflows by domain
# Workflow 1: user_registration
# Workflow 2: user_verification
# Workflow 3: payment_processing
# Workflow 4: analytics_calculation
```

**How to Split:**
1. Identify distinct business domains
2. Create separate workflow for each domain
3. Use RabbitMQ or REST API to chain workflows
4. Each workflow should have single responsibility

---

#### 🚫 Anti-Pattern 2: Spaghetti Conditionals

**Smell:** Complex nested conditionals with unclear flow.

```yaml
# BAD: Impossible to follow logic
check_a:
  type: conditional
  condition:
    variableCheck: [${{a}}]
    value:
      - case: [true]
        target: check_b
  defaultTasK: check_c

check_b:
  type: conditional
  condition:
    variableCheck: [${{b}}]
    value:
      - case: [true]
        target: check_d
      - case: [false]
        target: check_e
  defaultTasK: check_f

check_c:
  type: conditional
  condition:
    variableCheck: [${{c}}]
    value:
      - case: [true]
        target: check_g
  defaultTasK: check_b  # Circular reference potential!
```

**Problems:**
- Cannot understand flow without drawing diagram
- High cyclomatic complexity
- Circular reference risks
- Maintenance nightmare

**Refactoring:**

✅ **Flatten with early exits:**

```yaml
# GOOD: Linear flow with early exits
validate_all_rules:
  type: worker
  sourceRef:
    body:
      checks: ["a", "b", "c", "d"]
  targetRef:
    - check_validation_result
  worker: middleware

check_validation_result:
  type: conditional
  condition:
    variableCheck:
      - ${{validate_all_rules.data.data.all_passed}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          failed_checks: ${{validate_all_rules.data.data.failures}}
  defaultTasK: process_approved_data
```

**Refactoring Strategy:**
1. Combine related checks into single worker
2. Worker returns consolidated result
3. Single conditional for early exit
4. Linear flow for happy path

---

#### 🚫 Anti-Pattern 3: Variable Hell

**Smell:** Uncontrolled variable proliferation.

```yaml
# BAD: Variables everywhere, no organization
worker_1:
  save_global:
    a: ${{worker_1.data.x}}
    b: ${{worker_1.data.y}}
    c: ${{worker_1.data.z}}

worker_2:
  save_global:
    d: ${{worker_2.data.m}}
    e: ${{worker_2.data.n}}
    a2: ${{worker_2.data.p}}  # Name collision risk!

worker_3:
  save_global:
    f: ${{worker_3.data.q}}
    temp1: ${{worker_3.data.r}}
    temp2: ${{worker_3.data.s}}

# ... now you have 50+ global variables
```

**Problems:**
- Name collisions (a, a2, a_final?)
- Unclear which variables are still relevant
- Memory waste
- Hard to understand dependencies

**Refactoring:**

✅ **Namespace and minimize:**

```yaml
# GOOD: Organized, minimal globals
identify_user:
  save_global:
    user_id: ${{identify_user.data.data.id}}
    user_tier: ${{identify_user.data.data.tier}}

# Access previous worker data directly when possible
process_user:
  sourceRef:
    body:
      user_id: ${{ibridge.global.user_id}}
      user_name: ${{identify_user.data.data.name}}  # Direct access
      user_email: ${{identify_user.data.data.email}}  # No need to save
```

**Guidelines:**
- Only save to global if used in 3+ places
- Use namespacing: `user_id`, `order_id`, `payment_id`
- Access worker data directly when possible
- Clear globals that are no longer needed
- Document what each global variable represents

---

#### 🚫 Anti-Pattern 4: Silent Failures

**Smell:** No error handling, failures go unnoticed.

```yaml
# BAD: No error checks anywhere
step_1:
  type: worker
  targetRef:
    - step_2  # What if step_1 fails?
  worker: middleware

step_2:
  type: worker
  targetRef:
    - step_3  # What if step_2 fails?
  worker: middleware

step_3:
  type: worker
  targetRef:
    - ibridge-reply  # What if step_3 fails?
  worker: middleware
```

**Problems:**
- Silent failures
- Invalid data propagates
- No visibility into what went wrong
- Users get generic 200 OK with error data

**Refactoring:**

✅ **Add explicit error handling:**

```yaml
# GOOD: Error check after each critical step
step_1:
  type: worker
  targetRef:
    - check_step_1_result
  worker: middleware

check_step_1_result:
  type: conditional
  condition:
    variableCheck:
      - ${{step_1.data.success}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 500
        save_global:
          status_code: 500
          error_code: "STEP_1_FAILED"
          message: ${{step_1.data.error}}
          failed_at: "step_1"
  defaultTasK: step_2

step_2:
  type: worker
  targetRef:
    - check_step_2_result
  worker: middleware

# ... continue pattern
```

---

#### 🚫 Anti-Pattern 5: Magic Variables

**Smell:** Unexplained variable references.

```yaml
# BAD: Where did these come from?
final_step:
  sourceRef:
    body:
      x: ${{ibridge.global.xyz}}  # What is xyz?
      data: ${{some_task.data.data.thing}}  # Which task?
      value: ${{ibridge.global.temp_val}}  # Temporary from where?
```

**Problems:**
- Cannot understand data flow
- Breaks when referenced task removed
- Hard to maintain

**Refactoring:**

✅ **Document and use clear names:**

```yaml
# GOOD: Clear provenance
# Step 1: Extract user ID from authentication
authenticate_user:
  # ...
  save_global:
    authenticated_user_id: ${{authenticate_user.data.data.user_id}}

# Step 2: Fetch user profile using authenticated ID
fetch_user_profile:
  sourceRef:
    body:
      user_id: ${{ibridge.global.authenticated_user_id}}
  # ...
  save_global:
    user_email: ${{fetch_user_profile.data.data.email}}
    user_tier: ${{fetch_user_profile.data.data.subscription_tier}}

# Step 3: Process order for authenticated user
process_order:
  sourceRef:
    body:
      user_id: ${{ibridge.global.authenticated_user_id}}
      user_email: ${{ibridge.global.user_email}}
      subscription_tier: ${{ibridge.global.user_tier}}
```

**Guidelines:**
- Use descriptive variable names
- Comment where globals are set
- Comment why they're needed
- Show data lineage in comments

---

#### 🚫 Anti-Pattern 6: Copy-Paste Duplication

**Smell:** Same logic repeated multiple times.

```yaml
# BAD: Same validation copied 5 times
validate_email_1:
  type: conditional
  condition:
    variableCheck:
      - ${{email_1.data.data.valid}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          message: "Email 1 invalid"
  defaultTasK: validate_email_2

validate_email_2:
  type: conditional
  condition:
    variableCheck:
      - ${{email_2.data.data.valid}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          message: "Email 2 invalid"
  defaultTasK: validate_email_3

# ... 3 more copies
```

**Refactoring:**

✅ **Consolidate in worker:**

```yaml
# GOOD: Single worker validates all
validate_all_emails:
  type: worker
  sourceRef:
    body:
      emails:
        - ${{startEvent.Body.primary_email}}
        - ${{startEvent.Body.secondary_email}}
        - ${{startEvent.Body.billing_email}}
  targetRef:
    - check_all_emails_valid
  worker: middleware

check_all_emails_valid:
  type: conditional
  condition:
    variableCheck:
      - ${{validate_all_emails.data.data.all_valid}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          message: "Email validation failed"
          invalid_emails: ${{validate_all_emails.data.data.invalid}}
  defaultTasK: continue_workflow
```

**Or extract to separate workflow:**

```yaml
# GOOD: Reusable workflow
# In email_validation workflow:
name: email_validation
# ... validation logic

# In main workflow:
call_email_validation:
  type: worker
  sourceRef:
    body:
      emails: [...]
  worker: email_validation_service
```

---

#### 🚫 Anti-Pattern 7: Monolithic Workers

**Smell:** Worker does too many unrelated things.

```yaml
# BAD: Worker does everything
do_everything:
  type: worker
  sourceRef:
    body:
      # Validate
      validate_email: true
      validate_phone: true
      # Process
      create_user: true
      send_notifications: true
      # Analytics
      track_event: true
      update_dashboard: true
  worker: god_worker
```

**Problems:**
- Long execution time
- Hard to debug
- Cannot reuse parts
- All-or-nothing failure

**Refactoring:**

✅ **Split into focused workers:**

```yaml
# GOOD: Single-purpose workers
validate_user_data:
  type: worker
  sourceRef:
    body:
      email: ${{startEvent.Body.email}}
      phone: ${{startEvent.Body.phone}}
  targetRef:
    - check_validation
  worker: validation_service

create_user_record:
  type: worker
  sourceRef:
    body:
      email: ${{startEvent.Body.email}}
      phone: ${{startEvent.Body.phone}}
  targetRef:
    - send_notifications
  worker: user_service

send_notifications:
  type: worker
  sourceRef:
    body:
      user_id: ${{create_user_record.data.data.id}}
  targetRef:
    - track_analytics
  worker: notification_service

track_analytics:
  type: worker
  sourceRef:
    body:
      event: "user_created"
      user_id: ${{create_user_record.data.data.id}}
  targetRef:
    - ibridge-reply
  worker: analytics_service
```

---

### Performance Code Smells

#### 🐌 Smell 1: Sequential When Parallel Possible

```yaml
# BAD: Sequential when independent
fetch_user_profile:
  targetRef:
    - fetch_user_preferences

fetch_user_preferences:
  targetRef:
    - fetch_user_activity

fetch_user_activity:
  targetRef:
    - fetch_user_analytics

fetch_user_analytics:
  targetRef:
    - combine_all

# Total time: 4 workers × avg_time
```

**Current Limitation:** Orchestrator doesn't support parallel execution within workflow.

**Workaround Options:**

1. **Create parallel worker:**
```yaml
# GOOD: Worker handles parallelization
fetch_all_user_data:
  type: worker
  sourceRef:
    body:
      user_id: ${{ibridge.global.user_id}}
      fetch_parallel:
        - profile
        - preferences
        - activity
        - analytics
  worker: parallel_fetch_service
```

2. **Use async workers (if available):**
```yaml
# Trigger async jobs, poll for completion
trigger_async_fetches:
  type: worker
  sourceRef:
    body:
      jobs: [profile, preferences, activity, analytics]
  targetRef:
    - poll_job_completion
  worker: async_orchestrator
```

---

#### 🐌 Smell 2: N+1 Pattern

```yaml
# BAD: Loop pattern (workflow runs multiple times)
# Workflow called once per item
get_user:
  targetRef:
    - get_user_orders

get_user_orders:
  # Returns 100 order IDs
  targetRef:
    - get_order_details

get_order_details:
  # Only gets 1 order!
  # Workflow needs to run 100 times!
```

**Refactoring:**

✅ **Batch operations:**

```yaml
# GOOD: Batch fetch
get_user:
  targetRef:
    - get_user_orders

get_user_orders:
  targetRef:
    - get_all_order_details

get_all_order_details:
  type: worker
  sourceRef:
    body:
      order_ids: ${{get_user_orders.data.data.order_ids}}  # All IDs
  worker: batch_order_service  # Returns all orders in one call
```

---

#### 🐌 Smell 3: Unnecessary Data Transfer

```yaml
# BAD: Passing huge datasets between workers
fetch_all_transactions:
  # Returns 10MB of data
  targetRef:
    - process_transactions

process_transactions:
  sourceRef:
    body:
      all_transactions: ${{fetch_all_transactions.data}}  # 10MB passed
  targetRef:
    - analyze_transactions

analyze_transactions:
  sourceRef:
    body:
      all_transactions: ${{fetch_all_transactions.data}}  # 10MB passed again
```

**Refactoring:**

✅ **Pass references, not data:**

```yaml
# GOOD: Pass transaction ID or reference
fetch_all_transactions:
  # Saves to temp storage, returns reference
  save_global:
    transaction_batch_id: ${{fetch_all_transactions.data.data.batch_id}}
  targetRef:
    - process_transactions

process_transactions:
  sourceRef:
    body:
      batch_id: ${{ibridge.global.transaction_batch_id}}  # Just ID
  targetRef:
    - analyze_transactions

analyze_transactions:
  sourceRef:
    body:
      batch_id: ${{ibridge.global.transaction_batch_id}}  # Just ID
```

---

### Security Code Smells

#### 🔓 Smell 1: Sensitive Data in Logs

```yaml
# BAD: Passwords, tokens logged
authenticate_user:
  sourceRef:
    body:
      username: ${{startEvent.Body.username}}
      password: ${{startEvent.Body.password}}  # LOGGED IN WORKFLOW STATE!
      api_key: ${{startEvent.Body.api_key}}    # LOGGED IN WORKFLOW STATE!
```

**Problems:**
- Stored in `workflow_states` table
- Visible in traces
- Compliance violations

**Refactoring:**

✅ **Never pass sensitive data through workflow:**

```yaml
# GOOD: Pass only references
authenticate_user:
  sourceRef:
    body:
      username: ${{startEvent.Body.username}}
      # Password sent in Authorization header, not body
      # Worker extracts from header, doesn't log
  worker: secure_auth_service
  save_global:
    auth_token: ${{authenticate_user.data.data.token}}  # Safe token only
```

**Guidelines:**
- Never include passwords in workflow
- Never include API keys in workflow
- Never include credit card numbers
- Use tokens/references instead
- Sensitive data should stay in worker, not workflow state

---

#### 🔓 Smell 2: Authorization Not Checked

```yaml
# BAD: No authorization check
delete_user:
  type: worker
  sourceRef:
    body:
      user_id: ${{startEvent.Body.user_id}}  # Any user can delete any user!
  worker: middleware
```

**Refactoring:**

✅ **Add authorization step:**

```yaml
# GOOD: Verify permission first
check_delete_permission:
  type: worker
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      action: "delete_user"
      resource_id: ${{startEvent.Body.user_id}}
      requester_id: ${{startEvent.Header.X-User-Id}}
  targetRef:
    - verify_permission
  worker: authorization_service

verify_permission:
  type: conditional
  condition:
    variableCheck:
      - ${{check_delete_permission.data.data.authorized}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 403
        save_global:
          status_code: 403
          error_code: "FORBIDDEN"
          message: "Insufficient permissions"
  defaultTasK: delete_user

delete_user:
  type: worker
  sourceRef:
    body:
      user_id: ${{startEvent.Body.user_id}}
  worker: middleware
```

---

### Maintainability Checklist

**Before committing a workflow, check:**

- [ ] **Single Responsibility**: Workflow does one thing well
- [ ] **Error Handling**: Every worker has conditional check
- [ ] **Clear Naming**: Task names explain purpose
- [ ] **Minimal Globals**: Only save what's needed multiple times
- [ ] **No Secrets**: No passwords/keys in workflow state
- [ ] **Authorization**: Permissions checked for sensitive operations
- [ ] **Documentation**: Comments explain complex logic
- [ ] **Testable**: Can test with sample data
- [ ] **Observable**: Sufficient logging/tracing points
- [ ] **No Magic Numbers**: Status codes explained
- [ ] **Reasonable Length**: < 20 tasks ideally, < 30 maximum
- [ ] **No Duplication**: Repeated logic extracted

---

### Refactoring Decision Tree

```
Is workflow hard to understand?
├─ YES: Too many tasks (>30)?
│  └─ Split into multiple workflows
└─ NO: Continue

Are there repeated patterns?
├─ YES: Same logic in 3+ places?
│  ├─ Extract to separate workflow
│  └─ Or combine in worker
└─ NO: Continue

Are there deep conditional nests (>3 levels)?
├─ YES: Flatten with early exits
└─ NO: Continue

Do workers fail silently?
├─ YES: Add error handling conditionals
└─ NO: Continue

Are there 10+ global variables?
├─ YES: Reduce, namespace, or document
└─ NO: Continue

Is workflow slow (>30s)?
├─ YES: Identify bottlenecks
│  ├─ Can parallelize? → Refactor worker
│  └─ N+1 pattern? → Batch operations
└─ NO: Continue

Does workflow handle secrets?
├─ YES: Remove from workflow state
│  └─ Use tokens/references
└─ NO: Continue

Ship it! ✅
```

---

### When to Refactor

**Immediate (Before Merge):**
- Security issues (secrets in state)
- No error handling
- Authorization bypassed

**Soon (Technical Debt):**
- Duplicated logic (3+ times)
- God workflow (>30 tasks)
- Performance issues (>30s execution)
- Deep nesting (>3 levels)

**Eventually (Nice to Have):**
- Variable proliferation (>15 globals)
- Unclear naming
- Missing comments
- Could be more elegant

---

## Additional Resources

- **API Documentation**: See `API_DOCUMENTATION.md`
- **Migration Guide**: See `MIGRATION_JAEGER_TO_TEMPO.md`
- **Sample Workflows**: Check `sample/` directory
- **Database Schema**: See `CLAUDE.md` - Database Schema section
- **Postman Collection**: `CES-Orchestrator-API.postman_collection.json`

## Getting Help

1. Check `workflow_histories` table for execution details
2. Review `workflow_states` table for worker-level data
3. Enable telemetry and view traces in Grafana Tempo
4. Check RabbitMQ queue for stuck messages
5. Review application logs for errors
