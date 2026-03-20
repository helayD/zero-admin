# 验证清单详细参考

本文件包含 SKILL.md 中未详述的验证规则和示例。

---

## Effort Estimate 检查

**必需字段**：
- **Size**: XS (2-4h) | S (4-8h) | M (1-2d) | L (2-3d) | XL (3-5d)
- **Hours**: Range estimate
- **Risk**: Low | Medium | High

---

## API 规格示例（完整模板）

```markdown
### API Specification

**Endpoint**: `POST /api/v1/users/login`

**Request**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| email | string | yes | User email |
| password | string | yes | Password |

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Success Response** (200):
```json
{
  "token": "eyJhbG...",
  "user": { "id": 1, "email": "user@example.com" }
}
```

**Error Response** (401):
```json
{
  "error": "INVALID_CREDENTIALS",
  "message": "Email or password incorrect"
}
```
```

---

## Flutter/Dart 任务检查项

必须包含：
- **Widget Structure**: StatelessWidget vs StatefulWidget, widget tree hierarchy
- **State Management**: Provider/Bloc/Riverpod pattern with state classes
- **Navigation**: Route names, arguments, MaterialPageRoute config
- **Pubspec Dependencies**: Required packages with version constraints (e.g., `flutter_bloc: ^8.1.0`)
- **Asset References**: Images, fonts, config files in `assets/` with pubspec.yaml registration
- **Platform Specifics**: iOS/Android-specific code using `Platform.isAndroid` checks if needed

**示例 Flutter Widget Spec**:
```dart
// File: lib/src/features/auth/presentation/login_screen.dart
class LoginScreen extends StatefulWidget {
  final VoidCallback onLoginSuccess;
  const LoginScreen({Key? key, required this.onLoginSuccess}) : super(key: key);
}

// State Management (Bloc)
class LoginBloc extends Bloc<LoginEvent, LoginState> {
  final AuthRepository authRepository;
  // ...
}
```

---

## Error Handling

### File Reading Errors
- **Missing file**: Report as critical error, mark task as INVALID
- **Permission denied**: Skip with warning, require manual check
- **Encoding issues**: Attempt UTF-8 decode, fallback to ASCII

### Parsing Errors
- **Invalid YAML frontmatter**: Report line number, suggest fix
- **Malformed JSON in examples**: Non-critical warning, suggest correction
- **Corrupted markdown**: Attempt to parse sections individually

### Validation Failures
- **Timeout (>30s per task)**: Skip remaining checks, report partial results
- **Circular dependencies**: Detect using graph traversal, report cycle path
- **Missing Epic source**: Cannot auto-fix, mark items as `[NEEDS_EPIC_INPUT]`

### Recovery Strategy
```
IF file_read_error:
  → Log error, continue with next task
IF parse_error:
  → Report error with line number, continue validation
IF auto_fix_fails:
  → Mark field as [NEEDS_INPUT], add TODO comment
IF all_tasks_fail:
  → Generate diagnostic report, suggest Epic review
```

---

## Key Principle

> **After validation passes, every task file should be immediately actionable by any developer without asking "what does this mean?" or "where are the API specs?"**

The goal: Complete technical specifications in every task, zero ambiguity, zero external dependencies.

---

## Self-Containment Invalid Patterns

```
❌ "See Epic.md for API details"
❌ "Refer to PRD section 3.2"
❌ "Implementation similar to existing module"
❌ "Parameters TBD"
❌ "Details to be discussed"
❌ "Check design doc for specs"
```

---

## Acceptance Criteria 示例

```markdown
## Acceptance Criteria

### Feature Acceptance
- [ ] User can login with email/password
- [ ] JWT token expires after 24h
- [ ] Failed login shows error message

### Quality Acceptance
- [ ] Unit test coverage ≥ 90%
- [ ] API response time < 200ms
- [ ] No TypeScript errors
```

---

## Definition of Done

```markdown
- [ ] Code complete (no TODO/FIXME)
- [ ] Tests pass (unit + integration)
- [ ] Docs updated
- [ ] Deployment verified
```
