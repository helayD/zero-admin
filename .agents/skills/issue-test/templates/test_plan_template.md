---
issue: ${ISSUE_NUMBER}
created: ${DATETIME}
has_api_docs: ${HAS_API_DOCS}
---

# Issue #${ISSUE_NUMBER} Test Plan

## Test Scope
${SUMMARY}

## E2E Test Scenarios
- [ ] 场景1: ${SCENARIO_1}
- [ ] 场景2: ${SCENARIO_2}

## API Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| ${METHOD} | ${ENDPOINT} | ${DESC} |

## API Test Script
${API_SCRIPT_NOTE}

## Database Consistency Checks
| Table | Check Type | Expected |
|-------|------------|----------|
| ${TABLE} | ${CHECK_TYPE} | ${EXPECTED} |

## Test Environment
- Base URL: ${BASE_URL}
- Database: ${DATABASE}
- Test Data: ${TEST_DATA}

## Code Quality
- [ ] 代码冗余检查通过
- [ ] 硬编码检查通过
- [ ] TODO/FIXME 检查通过
