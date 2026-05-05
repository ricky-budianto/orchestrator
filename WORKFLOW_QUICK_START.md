# Workflow Quick Start Guide

Fast reference for creating common workflow patterns. For detailed documentation, see `WORKFLOW_GUIDE.md`.

## 5-Minute Workflow Creation

### Step 1: Create YAML File

```yaml
name: my_first_workflow
requestType: my_first_workflow
path: /api/ibridge/v1/my-workflow
method: POST
startEvent:
  targetRef: validate_input

serviceTask:
  validate_input:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        email: ${{startEvent.Body.email}}
    targetRef:
      - check_validation
    worker: middleware

  check_validation:
    type: conditional
    condition:
      variableCheck:
        - ${{validate_input.data.data.valid}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 400
          save_global:
            status_code: 400
            message: "Invalid input"
    defaultTasK: process_data

  process_data:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        email: ${{startEvent.Body.email}}
    targetRef:
      - ibridge-reply
    worker: middleware

  ibridge-reply:
    endParsing: Separated
    sourceRef:
      result: ${{process_data.data}}
      error_message: ${{ibridge.global.message}}
    type: endEvent
    taskRef: ibridge-reply
```

### Step 2: Base64 Encode

```bash
base64 -i my_workflow.yaml
# Or in Python:
# python3 -c "import base64; print(base64.b64encode(open('my_workflow.yaml', 'rb').read()).decode())"
```

### Step 3: Create in Database

```bash
curl -X POST 'http://localhost:8081/api/ibridge/v1/workflow_configuration' \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "my_first_workflow",
    "name": "My First Workflow",
    "path": "/api/ibridge/v1/my-workflow",
    "method": "POST",
    "description": "My first workflow",
    "configuration": "BASE64_ENCODED_YAML_HERE"
  }'
```

### Step 4: Test Workflow

**Option A: Dynamic Route (No restart needed)**
```bash
curl -X POST 'http://localhost:8081/api/ibridge/v2/orchestrate/my_first_workflow' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com"
  }'
```

**Option B: Static Route (Requires restart)**
```bash
# Restart service first
curl -X POST 'http://localhost:8081/api/ibridge/v1/my-workflow' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com"
  }'
```

---

## Common Patterns (Copy & Paste)

### Pattern 1: Simple API Call

```yaml
call_api:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      user_id: ${{startEvent.Body.user_id}}
      action: "process"
  targetRef:
    - ibridge-reply
  worker: middleware
```

### Pattern 2: Validation with Error Response

```yaml
validate_data:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body: ${{startEvent.Body}}
  targetRef:
    - check_valid
  worker: middleware

check_valid:
  type: conditional
  condition:
    variableCheck:
      - ${{validate_data.data.data.valid}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 400
        save_global:
          status_code: 400
          message: "Validation failed"
          error_code: "VAL_001"
          errors: ${{validate_data.data.data.errors}}
  defaultTasK: next_step
```

### Pattern 3: Multi-Step Chain

```yaml
step_1:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body: ${{startEvent.Body}}
  targetRef:
    - step_2
  worker: middleware
  save_global:
    user_id: ${{step_1.data.data.id}}

step_2:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      user_id: ${{ibridge.global.user_id}}
      data: ${{step_1.data.data}}
  targetRef:
    - step_3
  worker: middleware

step_3:
  type: worker
  sourceRefParsing: ${{ibridge.httprequest}}
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      user_id: ${{ibridge.global.user_id}}
      previous: ${{step_2.data}}
  targetRef:
    - ibridge-reply
  worker: middleware
```

### Pattern 4: Boolean Check

```yaml
check_flag:
  type: conditional
  condition:
    variableCheck:
      - ${{previous_task.data.data.enabled}}
    value:
      - case:
          - true
        target: handle_enabled
      - case:
          - false
        target: handle_disabled
  defaultTasK: handle_unknown
```

### Pattern 5: String Match

```yaml
route_by_type:
  type: conditional
  condition:
    variableCheck:
      - ${{previous_task.data.data.type}}
    value:
      - case:
          - "PREMIUM"
        target: handle_premium
      - case:
          - "STANDARD"
        target: handle_standard
      - case:
          - "BASIC"
        target: handle_basic
  defaultTasK: handle_unknown_tier
```

### Pattern 6: Multi-Field Validation

```yaml
check_requirements:
  type: conditional
  condition:
    variableCheck:
      - ${{check_age.data.data.eligible}}
      - ${{check_income.data.data.sufficient}}
      - ${{check_credit.data.data.approved}}
    value:
      - case:
          - true
          - true
          - true
        target: approve
  defaultTasK: reject
```

### Pattern 7: Error with Details

```yaml
error_condition:
  type: conditional
  condition:
    variableCheck:
      - ${{task.data.success}}
    value:
      - case:
          - false
        target: ibridge-reply
        status_code: 422
        save_global:
          status_code: 422
          message: "Processing failed"
          error_code: "PROC_001"
          error_details:
            reason: ${{task.data.error.reason}}
            field: ${{task.data.error.field}}
            timestamp: ${{ibridge.timestamp}}
  defaultTasK: continue
```

### Pattern 8: Success Response

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    success: true
    data:
      id: ${{create_record.data.data.id}}
      status: ${{create_record.data.data.status}}
      created_at: ${{create_record.data.data.created_at}}
    metadata:
      workflow_id: ${{ibridge.global.workflow_id}}
      timestamp: ${{ibridge.timestamp}}
  type: endEvent
  taskRef: ibridge-reply
```

### Pattern 9: Error Response

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    success: false
    error:
      code: ${{ibridge.global.error_code}}
      message: ${{ibridge.global.message}}
      details: ${{ibridge.global.error_details}}
    metadata:
      timestamp: ${{ibridge.timestamp}}
  type: endEvent
  taskRef: ibridge-reply
```

### Pattern 10: Conditional Response

```yaml
ibridge-reply:
  endParsing: Separated
  sourceRef:
    # Success data (may be null on error)
    result: ${{final_task.data}}

    # Error data (may be null on success)
    error:
      code: ${{ibridge.global.error_code}}
      message: ${{ibridge.global.message}}

    # Always present
    status_code: ${{ibridge.global.status_code}}
    timestamp: ${{ibridge.timestamp}}
  type: endEvent
  taskRef: ibridge-reply
```

---

## Variable Reference Cheat Sheet

### Request Data
```yaml
${{startEvent.Body.field_name}}               # Request body field
${{startEvent.Body.nested.field}}             # Nested field
${{startEvent.Header.Authorization}}          # Header value
${{startEvent.Header.X-Custom-Header}}        # Custom header
${{startEvent.Params.id}}                     # URL parameter
${{startEvent.QueryParams.filter}}            # Query string
${{startEvent.Authorization}}                 # Auth header shortcut
```

### Worker Results
```yaml
${{worker_name.data}}                         # Full response
${{worker_name.data.data}}                    # Common data field
${{worker_name.data.data.field}}              # Specific field
${{worker_name.data.nested.deep.field}}       # Deep nesting
```

### Global Variables
```yaml
${{ibridge.global.user_id}}                   # Saved global
${{ibridge.global.status_code}}               # HTTP status
${{ibridge.global.error_code}}                # Error code
${{ibridge.global.message}}                   # Error message
${{ibridge.timestamp}}                        # Current time
```

---

## Complete Minimal Example

```yaml
name: minimal_example
requestType: minimal_example
path: /api/ibridge/v1/minimal
method: POST
startEvent:
  targetRef: process

serviceTask:
  process:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      body: ${{startEvent.Body}}
    targetRef:
      - ibridge-reply
    worker: middleware

  ibridge-reply:
    endParsing: Separated
    sourceRef:
      result: ${{process.data}}
    type: endEvent
    taskRef: ibridge-reply
```

---

## Complete Realistic Example

```yaml
name: user_registration
requestType: user_registration
path: /api/ibridge/v1/register-user
method: POST
startEvent:
  targetRef: validate_email

serviceTask:
  # Step 1: Validate Email
  validate_email:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        email: ${{startEvent.Body.email}}
    targetRef:
      - check_email_valid
    worker: middleware

  check_email_valid:
    type: conditional
    condition:
      variableCheck:
        - ${{validate_email.data.data.valid}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 422
          save_global:
            status_code: 422
            error_code: "EMAIL_INVALID"
            message: "Email validation failed"
    defaultTasK: validate_phone

  # Step 2: Validate Phone
  validate_phone:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        phone: ${{startEvent.Body.phone}}
    targetRef:
      - check_phone_valid
    worker: middleware

  check_phone_valid:
    type: conditional
    condition:
      variableCheck:
        - ${{validate_phone.data.data.valid}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 422
          save_global:
            status_code: 422
            error_code: "PHONE_INVALID"
            message: "Phone validation failed"
    defaultTasK: create_user

  # Step 3: Create User
  create_user:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        email: ${{startEvent.Body.email}}
        phone: ${{startEvent.Body.phone}}
        full_name: ${{startEvent.Body.full_name}}
        email_verified: ${{validate_email.data.data.verified}}
        phone_verified: ${{validate_phone.data.data.verified}}
    targetRef:
      - check_user_created
    worker: middleware

  check_user_created:
    type: conditional
    condition:
      variableCheck:
        - ${{create_user.data.success}}
      value:
        - case:
            - false
          target: ibridge-reply
          status_code: 500
          save_global:
            status_code: 500
            error_code: "USER_CREATE_FAILED"
            message: "Failed to create user"
    defaultTasK: send_welcome_email

  # Step 4: Send Welcome Email
  send_welcome_email:
    type: worker
    sourceRefParsing: ${{ibridge.httprequest}}
    sourceRef:
      authorization: ${{startEvent.Authorization}}
      body:
        user_id: ${{create_user.data.data.id}}
        email: ${{startEvent.Body.email}}
        full_name: ${{startEvent.Body.full_name}}
    targetRef:
      - ibridge-reply
    worker: middleware

  # Final Response
  ibridge-reply:
    endParsing: Separated
    sourceRef:
      success: true
      user:
        id: ${{create_user.data.data.id}}
        email: ${{startEvent.Body.email}}
        phone: ${{startEvent.Body.phone}}
        full_name: ${{startEvent.Body.full_name}}
      verification:
        email: ${{validate_email.data}}
        phone: ${{validate_phone.data}}
      notification:
        welcome_email_sent: ${{send_welcome_email.data.success}}
      error:
        code: ${{ibridge.global.error_code}}
        message: ${{ibridge.global.message}}
      timestamp: ${{ibridge.timestamp}}
    type: endEvent
    taskRef: ibridge-reply
```

---

## Testing Your Workflow

### 1. Check Database Entry

```sql
SELECT id, name, path, method, status
FROM workflow_configurations
WHERE id = 'your_workflow_id';
```

### 2. Test via V2 API (Instant)

```bash
curl -X POST 'http://localhost:8081/api/ibridge/v2/orchestrate/your_workflow_id' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your_token' \
  -d '{
    "field1": "value1",
    "field2": "value2"
  }'
```

### 3. Check Execution History

```sql
SELECT id, workflow_configuration_id, status, created_at
FROM workflow_histories
WHERE workflow_configuration_id = 'your_workflow_id'
ORDER BY created_at DESC
LIMIT 10;
```

### 4. Debug Worker States

```sql
SELECT id, workflow_id, state
FROM workflow_states
WHERE workflow_id = 'your_workflow_history_id';
```

### 5. View in Grafana

1. Open http://localhost:3001
2. Go to Explore → Select Tempo
3. Search by workflow ID or time range
4. View trace spans for each worker

---

## Common Mistakes to Avoid

❌ **Using `/v2` in workflow path**
```yaml
path: /api/ibridge/v2/my-workflow  # WRONG - reserved!
```
✅ **Use `/v1` or other path**
```yaml
path: /api/ibridge/v1/my-workflow  # Correct
```

---

❌ **Lowercase 'k' in defaultTask**
```yaml
defaultTask: next_step  # WRONG - won't work!
```
✅ **Capital 'K' in defaultTasK**
```yaml
defaultTasK: next_step  # Correct
```

---

❌ **Missing conditional check after worker**
```yaml
some_worker:
  targetRef:
    - next_worker  # WRONG - no error handling!
```
✅ **Add conditional check**
```yaml
some_worker:
  targetRef:
    - check_worker_result  # Correct

check_worker_result:
  type: conditional
  # ... error handling ...
```

---

❌ **Wrong variable scope**
```yaml
sourceRef:
  body:
    field: ${{Body.field}}  # WRONG - missing scope!
```
✅ **Include scope**
```yaml
sourceRef:
  body:
    field: ${{startEvent.Body.field}}  # Correct
```

---

❌ **Forgetting save_global for status code**
```yaml
condition:
  value:
    - case:
        - false
      target: ibridge-reply
      status_code: 400  # Only sets in conditional, not global!
```
✅ **Save to global**
```yaml
condition:
  value:
    - case:
        - false
      target: ibridge-reply
      status_code: 400
      save_global:
        status_code: 400  # Now available in ibridge-reply
```

---

---

## Code Quality Checklist

Before deploying your workflow, verify:

### ✅ Must Have (Critical)
- [ ] **Error handling** after every worker
- [ ] **No secrets** in workflow state (passwords, API keys, tokens)
- [ ] **Authorization checks** for sensitive operations
- [ ] **Status codes** set correctly in error paths
- [ ] **Capital K** in `defaultTasK` (not `defaultTask`)

### ⚠️ Should Have (Important)
- [ ] **Task names** are descriptive and clear
- [ ] **< 30 tasks** total (split if larger)
- [ ] **< 10 global variables** (reduce if more)
- [ ] **No duplicate code** (same logic 3+ times)
- [ ] **Comments** for complex logic

### 💡 Nice to Have (Quality)
- [ ] **< 20 tasks** ideally
- [ ] **Linear flow** (avoid deep nesting)
- [ ] **Single responsibility** (one domain per workflow)
- [ ] **Performance optimized** (batch operations, avoid N+1)

---

## Anti-Patterns Quick Reference

### 🚫 Don't Do This

```yaml
# ❌ Silent failures (no error handling)
worker1:
  targetRef: [worker2]  # What if worker1 fails?

# ❌ Secrets in workflow state
auth:
  sourceRef:
    body:
      password: ${{startEvent.Body.password}}  # LOGGED!

# ❌ Using /v2 in path
path: /api/ibridge/v2/my-workflow  # RESERVED!

# ❌ Lowercase 'k' in defaultTask
defaultTask: next_step  # WON'T WORK!

# ❌ Magic variables
sourceRef:
  body:
    x: ${{ibridge.global.xyz}}  # What is xyz?

# ❌ No authorization check
delete_user:
  sourceRef:
    body:
      user_id: ${{startEvent.Body.user_id}}  # Anyone can delete!
```

### ✅ Do This Instead

```yaml
# ✅ Error handling
worker1:
  targetRef: [check_worker1_result]

check_worker1_result:
  type: conditional
  condition:
    variableCheck: [${{worker1.data.success}}]
    value:
      - case: [false]
        target: ibridge-reply
        status_code: 500
        save_global:
          error_code: "WORKER1_FAILED"
  defaultTasK: worker2

# ✅ No secrets in state
auth:
  sourceRef:
    body:
      username: ${{startEvent.Body.username}}
      # Password in Authorization header only
  save_global:
    auth_token: ${{auth.data.data.token}}  # Safe token

# ✅ Use /v1 or other path
path: /api/ibridge/v1/my-workflow

# ✅ Capital 'K'
defaultTasK: next_step

# ✅ Descriptive variables
# Set in authenticate_user task
save_global:
  authenticated_user_id: ${{auth.data.data.user_id}}

# Later use
sourceRef:
  body:
    user_id: ${{ibridge.global.authenticated_user_id}}

# ✅ Authorization check
check_permission:
  type: worker
  sourceRef:
    authorization: ${{startEvent.Authorization}}
    body:
      action: "delete_user"
      resource_id: ${{startEvent.Body.user_id}}
  targetRef: [verify_permission]

verify_permission:
  type: conditional
  condition:
    variableCheck: [${{check_permission.data.data.authorized}}]
    value:
      - case: [false]
        target: ibridge-reply
        status_code: 403
  defaultTasK: delete_user
```

---

## Refactoring Indicators

**Refactor IMMEDIATELY if:**
- 🔴 Secrets (passwords/keys) in workflow state
- 🔴 No error handling anywhere
- 🔴 Missing authorization for sensitive ops

**Refactor SOON if:**
- 🟡 > 30 tasks in workflow
- 🟡 Same logic duplicated 3+ times
- 🟡 Workflow takes > 30 seconds
- 🟡 > 15 global variables
- 🟡 Deep conditional nesting (> 3 levels)

**Refactor EVENTUALLY if:**
- 🟢 Unclear task naming
- 🟢 Missing comments on complex logic
- 🟢 Could be more elegant

**See `WORKFLOW_GUIDE.md` §11 for detailed refactoring patterns.**

---

## Next Steps

- Read full documentation: `WORKFLOW_GUIDE.md`
- Review anti-patterns & refactoring: `WORKFLOW_GUIDE.md` - Section 11
- View sample workflows: `sample/` directory
- API documentation: `API_DOCUMENTATION.md`
- Migration guide: `MIGRATION_JAEGER_TO_TEMPO.md`
- Architecture: `CLAUDE.md`
