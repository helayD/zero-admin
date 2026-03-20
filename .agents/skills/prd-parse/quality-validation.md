# 质量验证清单详细参考

本文件定义 prd-parse Step 4 质量验证的完整检查清单。

---

## Step 0: Enhanced Feature Mining Validation

- [ ] All mined features documented and tracked
- [ ] Functional variants, extensions, boundary cases listed
- [ ] All functional dependencies identified
- [ ] **All P0 features explicitly covered in Epic**
- [ ] **All high-value features have implementation tasks**
- [ ] Feature-value assessment reflected in prioritization
- [ ] If any P0/high-value missing, STOP and correct

---

## Step 1: Feature Inventory Comparison

- [ ] Checklist of EVERY PRD feature created
- [ ] Each feature verified in epic (Technical Approach or Tasks)
- [ ] Mark: ✅ covered or ❌ missing
- [ ] If ANY ❌, STOP and add before proceeding

---

## Step 2: Section-by-Section PRD Review

- [ ] PRD "User Stories" → all covered in epic
- [ ] PRD "Functional Requirements" → all mapped to tasks
- [ ] PRD "Non-Functional Requirements" → in Architecture/Infrastructure
- [ ] PRD "Success Criteria" → technical acceptance criteria
- [ ] PRD "Constraints" → in Dependencies/Implementation Strategy

---

## Step 3: Cross-Reference Validation

- [ ] Every PRD bullet has corresponding epic item
- [ ] Every user story has implementing tasks
- [ ] Every API/integration in PRD in technical approach
- [ ] Every data model/entity has database/backend tasks

---

## Step 4: Technical Detail Validation

Execute checks and output diff report:
- [ ] API count: count(PRD APIs) vs count(Epic APIs)
- [ ] CRUD completeness: verify Create/Read/Update/Delete for each entity
- [ ] Field coverage: diff(API request/response fields, DDL table fields), output missing
- [ ] Index coverage: verify each query condition has corresponding index
- [ ] Foreign key relations: verify table relationships reflected in DDL
- [ ] Error code completeness: verify each API exception has error code

**Output format:**
```
| Check Item | PRD Count | Epic Count | Gap |
|------------|-----------|------------|-----|
| APIs       | X         | Y          | Missing: xxx |
| Tables     | X         | Y          | Missing: xxx |
| Fields     | X         | Y          | Missing: xxx |
```
**BLOCKING:** If any gap > 0, STOP and complete before proceeding

---

## Quality Checklist

### PRD Alignment
- [ ] All requirements addressed
- [ ] User stories map to components
- [ ] Success criteria translated
- [ ] Non-functional requirements considered

### Architecture Quality
- [ ] Mermaid graph TB diagram exists
- [ ] All external systems shown
- [ ] Technology choices justified
- [ ] Architecture decisions explain trade-offs
- [ ] Security and data flows documented

### Task Breakdown Quality
- [ ] All implementation areas covered
- [ ] Clear scope and deliverables
- [ ] **Explicit PRD Coverage references**
- [ ] Appropriate sizing (1-3 days each)
- [ ] Logical dependencies
- [ ] Parallelization identified
- [ ] Critical path documented
- [ ] **All PRD requirements covered by tasks**

### Implementation Strategy Quality
- [ ] Phases clearly defined
- [ ] Risk mitigation included
- [ ] Testing approach covers all levels
- [ ] Deployment addressed

### Dependencies Quality
- [ ] External services listed with protocols
- [ ] Internal dependencies identified
- [ ] Prerequisites documented
- [ ] Blockers highlighted

### Success Criteria Quality
- [ ] Benchmarks quantified
- [ ] Quality gates measurable
- [ ] Acceptance criteria testable
- [ ] Non-functional included

---

## Coverage Report Format

```
✅ PRD Coverage Review Complete

📋 Enhanced Feature Coverage:
- Total PRD Features (Explicit): X
- Mined Features (Variants/Extensions): X
- Boundary Cases: X
- **Total Features**: X
- Covered in Epic: X
- **Coverage Rate: 100%** ✅

🎯 Priority Distribution:
- P0 Core: X (X%)
- P1 Important: X (X%)
- P2 Value-Add: X (X%)

🔍 Section Coverage:
- User Stories: X/X ✅
- Functional Requirements: X/X ✅
- Feature Variants: X/X ✅
- Dependencies: X/X ✅
- Boundary Cases: X/X ✅
- Value Features: X/X ✅

⚠️ If ANY < 100%, LIST missing items and STOP
```

---

## Feature Value Assessment

For each feature, assess:
- **Business value**: revenue/cost/competitive advantage
- **User value**: problem-solving/efficiency/experience
- **Strategic value**: core business/long-term goals

Priority assignment:
- **P0 Core**: Must-have, blocks launch
- **P1 Important**: High-value, should-have
- **P2 Value-Add**: Nice-to-have, can defer
