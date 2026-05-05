# Workflow Sample Library

This directory contains comprehensive workflow examples demonstrating all features and patterns of the CES-iBridge Orchestrator system.

## 📚 Sample Files Overview

### 🎯 For Beginners

#### 1. **sample.yaml** - Simple Flow Example
- **Purpose**: Basic workflow with email and income validation
- **Concepts**: Worker tasks, conditionals, variable substitution
- **When to use**: Learning the basics, quick reference
- **Complexity**: ⭐ Beginner

#### 2. **user_verification_flow.yaml** - Realistic Business Flow
- **Purpose**: Complete user verification process with KYC
- **Concepts**: Multi-step validation, risk assessment, error handling
- **When to use**: Understanding production-ready workflows
- **Complexity**: ⭐⭐ Intermediate

### 🚀 For Advanced Users

#### 3. **complete_feature_demo.yaml** - All Features Showcase
- **Purpose**: Demonstrates EVERY workflow feature available
- **Concepts Covered**:
  - ✅ All task types (worker, conditional, endEvent)
  - ✅ All variable scopes (startEvent, worker, ibridge, global)
  - ✅ All comparison types (boolean, string, numeric, multiple)
  - ✅ Parallel execution pattern
  - ✅ Retry/fallback logic
  - ✅ Global variables and status codes
  - ✅ Data enrichment patterns
  - ✅ Complex conditional branching
- **When to use**: Reference for implementing complex features
- **Complexity**: ⭐⭐⭐⭐ Advanced
- **Lines**: 400+ with extensive inline documentation

#### 4. **endparsing_examples.yaml** - Response Formatting Guide
- **Purpose**: Demonstrates all endParsing types and their outputs
- **Concepts**:
  - `Separated`: Structured response with separate fields
  - `Combined`: Flat merged response
  - `raw`: Unprocessed response
  - Dynamic response format selection
- **When to use**: Deciding how to format API responses
- **Complexity**: ⭐⭐ Intermediate

#### 5. **edge_cases_and_patterns.yaml** - Production Patterns & Gotchas
- **Purpose**: Real-world patterns and common pitfalls
- **Patterns Covered**:
  - 🔄 Circuit Breaker (service fallback)
  - 📊 Saga Pattern (distributed transactions with compensation)
  - 🎯 Aggregation (fan-out/fan-in)
  - ⚡ Early Termination (fail fast validation)
  - 🔀 Dynamic Worker Selection
- **Gotchas Covered**:
  - ⚠️ `defaultTasK` uppercase K
  - ⚠️ Variable scope visibility
  - ⚠️ Numeric comparison format
  - ⚠️ Multiple conditions AND logic
  - ⚠️ Empty response handling
- **When to use**: Building production systems, debugging issues
- **Complexity**: ⭐⭐⭐⭐⭐ Expert

## 🎓 Learning Path

### Step 1: Understand the Basics
```
1. Read sample.yaml
2. Study the basic worker → conditional → endEvent flow
3. Understand variable substitution: ${{startEvent.Body.field}}
```

### Step 2: Learn Realistic Patterns
```
1. Read user_verification_flow.yaml
2. Study multi-step validation patterns
3. Understand error handling with save_global
```

### Step 3: Master All Features
```
1. Read complete_feature_demo.yaml section by section
2. Focus on sections relevant to your use case:
   - Parallel execution (lines 150-250)
   - Retry/fallback (lines 300-380)
   - Global variables (throughout)
```

### Step 4: Understand Response Formatting
```
1. Read endparsing_examples.yaml
2. Compare the three endParsing types
3. Choose the right format for your API
```

### Step 5: Avoid Common Pitfalls
```
1. Read edge_cases_and_patterns.yaml
2. Study the "Gotchas" section carefully
3. Reference production patterns when building complex flows
```

## 🔍 Quick Reference

### Find Examples By Feature

| Feature | File | Section/Lines |
|---------|------|---------------|
| **Basic worker task** | sample.yaml | Lines 13-23 |
| **Conditional with boolean** | sample.yaml | Lines 24-38 |
| **String comparison** | complete_feature_demo.yaml | Lines 90-120 |
| **Numeric comparison** | complete_feature_demo.yaml | Lines 130-150 |
| **Multiple conditions (AND)** | complete_feature_demo.yaml | Lines 250-280 |
| **Parallel execution** | complete_feature_demo.yaml | Lines 150-250 |
| **Retry/fallback pattern** | edge_cases_and_patterns.yaml | Circuit Breaker |
| **Global variables** | complete_feature_demo.yaml | Throughout |
| **Error handling** | user_verification_flow.yaml | Lines 29-43 |
| **endParsing: Separated** | endparsing_examples.yaml | Example 1 |
| **endParsing: Combined** | endparsing_examples.yaml | Example 2 |
| **endParsing: raw** | endparsing_examples.yaml | Example 3 |
| **Saga pattern** | edge_cases_and_patterns.yaml | Pattern 3 |
| **Circuit breaker** | edge_cases_and_patterns.yaml | Pattern 2 |
| **Aggregation (fan-out/fan-in)** | edge_cases_and_patterns.yaml | Pattern 4 |

## 🎯 Use Cases

### I want to...

#### Build a simple validation flow
→ Start with **sample.yaml**

#### Implement user onboarding with KYC
→ Study **user_verification_flow.yaml**

#### Call multiple services in parallel
→ See **complete_feature_demo.yaml** (Parallel Execution section)

#### Handle service failures gracefully
→ See **edge_cases_and_patterns.yaml** (Circuit Breaker pattern)

#### Implement a multi-step transaction with rollback
→ See **edge_cases_and_patterns.yaml** (Saga pattern)

#### Format my API response properly
→ See **endparsing_examples.yaml**

#### Understand variable scopes
→ See **complete_feature_demo.yaml** (extensively documented)

#### Debug a workflow issue
→ Check **edge_cases_and_patterns.yaml** (Gotchas section)

## 📋 Variable Reference Cheat Sheet

### Available Variable Scopes

```yaml
# 1. Start Event Variables
${{startEvent.Authorization}}               # Auth header
${{startEvent.Header.X-Custom-Header}}      # Any header
${{startEvent.Body.field_name}}             # Request body field
${{startEvent.QueryParam.param_name}}       # Query parameter

# 2. Worker Result Variables
${{worker_name.data.field}}                 # Worker response data
${{worker_name.data.data.nested_field}}     # Nested data

# 3. iBridge Special Variables
${{ibridge.httprequest}}                    # HTTP request template
${{ibridge.timestamp}}                      # Current timestamp
${{ibridge.global.variable_name}}           # Global variable

# 4. Global Variables (set via save_global)
save_global:
  custom_field: "value"
# Access: ${{ibridge.global.custom_field}}
```

## ⚠️ Common Gotchas

### 1. defaultTasK has uppercase K
```yaml
# ✅ CORRECT
defaultTasK: next_task

# ❌ WRONG
defaultTask: next_task  # Won't work!
```

### 2. Numeric comparisons need operator strings
```yaml
# ✅ CORRECT
- case:
  - ">1000"

# ❌ WRONG
- case:
  - 1000  # Won't work!
```

### 3. Multiple conditions create AND logic
```yaml
variableCheck:
  - ${{worker.data.field1}}  # Must match
  - ${{worker.data.field2}}  # AND must match
value:
  - case:
    - true   # field1 must be true
    - true   # AND field2 must be true
```

### 4. Variables from skipped tasks are null
```yaml
# If condition_route sends to route_a_handler,
# then route_b_handler never executes
sourceRef:
  result_a: ${{route_a_handler.data}}  # ✅ Available
  result_b: ${{route_b_handler.data}}  # ⚠️ Will be null!
```

## 📖 Additional Resources

- **[WORKFLOW_GUIDE.md](../WORKFLOW_GUIDE.md)** - Complete workflow documentation (2,124 lines)
- **[WORKFLOW_QUICK_START.md](../WORKFLOW_QUICK_START.md)** - Quick reference guide (827 lines)
- **[CLAUDE.md](../CLAUDE.md)** - Project architecture and commands

## 🔧 Testing Your Workflow

### 1. Validate YAML Syntax
```bash
# Check YAML is valid
yamllint sample/your_workflow.yaml
```

### 2. Load into Database
```bash
# Using the API (v2 for dynamic routing)
curl -X POST http://localhost:8081/api/ibridge/v2/workflow-config \
  -H "Content-Type: application/json" \
  -d @sample/your_workflow.yaml
```

### 3. Test the Workflow
```bash
# Call the workflow endpoint
curl -X POST http://localhost:8081/api/ibridge/v2/orchestrate/{workflow_id} \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test123",
    "amount": 1000
  }'
```

### 4. Check Workflow History
```bash
# Query workflow execution history
SELECT * FROM workflow_history
WHERE workflow_configuration_id = 'your_workflow_id'
ORDER BY created_at DESC LIMIT 10;
```

## 🎨 Workflow Naming Conventions

### Task Names
- Use **snake_case**: `check_user_status`, `validate_input`
- Be descriptive: `condition_email_validity` not `cond1`
- Include action: `send_notification` not `notification`

### Worker Names
- Match RabbitMQ queue names exactly
- Common workers: `middleware`, `payment_worker`, `notification_worker`

### Variables in save_global
- Use **snake_case**: `status_code`, `error_message`
- Be specific: `payment_error` not `error`
- Include context: `primary_service_failed` not `failed`

## 🚀 Next Steps

1. ✅ Choose a sample file based on your needs
2. ✅ Copy and modify it for your use case
3. ✅ Test thoroughly with real data
4. ✅ Review gotchas and edge cases
5. ✅ Monitor in production

## 💡 Tips for Success

1. **Start Simple**: Begin with sample.yaml, then add complexity
2. **Use Comments**: Document complex logic inline
3. **Test Error Paths**: Don't just test the happy path
4. **Handle Failures**: Always have error handling with save_global
5. **Structure Responses**: Use Separated endParsing for complex responses
6. **Monitor Performance**: Check workflow_history for slow tasks

## 📞 Need Help?

- Check the [Troubleshooting](../WORKFLOW_GUIDE.md#troubleshooting) section
- Review [Code Smells & Refactoring](../WORKFLOW_GUIDE.md#workflow-code-smells--refactoring)
- See [Best Practices](../WORKFLOW_GUIDE.md#best-practices)

---

**Last Updated**: 2025-10-05
**Total Sample Workflows**: 5
**Total Examples**: 30+
**Coverage**: 100% of workflow features
