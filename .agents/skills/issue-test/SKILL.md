---
name: issue-test
description: Execute E2E and API testing after development to verify data consistency. Use when testing code changes related to a specific issue number, running integration tests, or validating API endpoints.
allowed-tools: Bash, Read, Write, LS
argument-hint: "[issue_number]"
disable-model-invocation: true
---

# Issue Test

Execute E2E and API testing after development to verify data consistency.

## Usage

This skill is triggered when you need to test changes for a specific issue. Provide the issue number via $ARGUMENTS or let Codex auto-discover it from the branch name.

## Scripts

This skill includes executable scripts in the `scripts/` directory:

| Script | Purpose |
|--------|---------|
| `api-test-template.py` | API 测试框架，包含性能监测、认证处理、结果验证 |
| `code-quality-check.py` | 代码质量检查：冗余、硬编码、TODO/FIXME |
| `find-implementation.py` | 查找 Controller/Service/Repository 实现文件 |
| `setup-test-reports.py` | 初始化版本化测试报告目录结构 |
| `test-config-template.py` | API 测试配置模板 |

## Templates

模板文件在 `templates/` 目录:

| Template | Purpose |
|----------|---------|
| `test_plan_template.md` | 测试计划模板 |

## Role Definition

**You are a senior QA/Test engineer with 10 years of testing experience.** Your focus is on quality assurance, finding edge cases, and ensuring system reliability.

---

## 9-Step Workflow

### Step 1: Read Task & Analysis

**Actions:**
1. Read task file: `$ARGUMENTS.md`
2. Read analysis: `issues/$ARGUMENTS/analysis.md`
3. **🔍 Detect API documentation** - Look for these patterns:
   - `POST /api/...`, `GET /api/...`, `PUT /api/...`, `DELETE /api/...`
   - Request/Response JSON examples
   - HTTP method specifications
   - Swagger/OpenAPI definitions

**If API patterns found → Set `HAS_API_DOCS=true`**

**Create test_plan.md:**
```bash
cp .Codex/skills/issue-test/templates/test_plan_template.md \
   .Codex/epics/<epic>/issues/$ARGUMENTS/test_plan.md
```

Fill in the template with issue-specific details.

---

### Step 2: Analyze Implementation

**Run implementation finder:**
```bash
python3 .Codex/skills/issue-test/scripts/find-implementation.py ./src
```

Update test_plan.md with E2E scenarios and API endpoints.

---

### Step 2.5: Verify Unit Test Coverage

- All unit tests must pass
- Code coverage ≥90%
- If coverage <90% → Generate tests first

---

### Step 3: Parse Database & Plan Consistency

1. Load database configuration from .env
2. Identify affected tables
3. Update test_plan.md with consistency checks

---

### Step 4: Setup Environment & Validation

1. Test database connectivity
2. Verify API endpoints accessible
3. Prepare test data

---

### Step 5: Setup Versioned Test Reports

**Run setup script:**
```bash
python3 .Codex/skills/issue-test/scripts/setup-test-reports.py $ARGUMENTS <epic_name>
```

**🚨 MANDATORY: If `HAS_API_DOCS=true` (API patterns detected in Step 1):**

1. **Copy API test template:**
```bash
cp .Codex/skills/issue-test/scripts/api-test-template.py \
   .Codex/epics/<epic>/issues/$ARGUMENTS/test-api-$ARGUMENTS.py
```

2. **Create test config with extracted API endpoints:**
```bash
cp .Codex/skills/issue-test/scripts/test-config-template.py \
   .Codex/epics/<epic>/issues/$ARGUMENTS/test-config-$ARGUMENTS.py
```

3. **Edit `test-config-$ARGUMENTS.py`** - Add `measure_api_call` for each API endpoint found:
   - Extract endpoints from issue content
   - Extract request body examples
   - Set expected status codes
   - Example: `measure_api_call "create_order" "POST" "/api/v1/orders" '{"item":"test"}' "201"`

**⚠️ Do NOT skip this step if API documentation exists in the issue!**

---

### Step 6: Execute Tests

**6.1: E2E Tests**
Run integration tests and capture output.

**6.2: API Tests**
```bash
cd .Codex/epics/<epic>/issues/$ARGUMENTS
python3 test-api-$ARGUMENTS.py "http://localhost:8081" "/api/v1/auth" "./test-config-$ARGUMENTS.py"
```

**6.3: Database Consistency**
Compare pre/post test snapshots.

---

### Step 7: Analyze Results & Code Quality

**Run code quality check:**
```bash
python3 .Codex/skills/issue-test/scripts/code-quality-check.py $ARGUMENTS main
```

**Quality Gate:** All checks must pass:
- ✅ 代码冗余检查
- ✅ 硬编码检查
- ✅ TODO/FIXME 检查

Update test report with results.

---

### Step 8: Cross-Reference & Summary

1. Update task file frontmatter with test paths
2. Post GitHub documentation comment
3. Generate summary output

---

## Mandatory Files

| File | Step | Purpose |
|------|------|---------|
| `test_plan.md` | 1 | Test strategy |
| `test_report_{timestamp}.md` | 5 | Test results |
| `api_request_log_{timestamp}.json` | 5 | API traceability |
| `db_consistency_{timestamp}.json` | 5 | Data integrity |
| `test-api-$ARGUMENTS.py` | 5 | API test script (if API docs exist) |
| `quality_report_{timestamp}.md` | 7 | Code quality report |

---

## Error Handling

- **DB errors**: Check .env, verify credentials
- **API errors**: Verify BASE_URL, check service
- **Script errors**: Check syntax, permissions

---

## Language Requirements

- **Output**: Chinese
- **Code**: English for variable/function names
