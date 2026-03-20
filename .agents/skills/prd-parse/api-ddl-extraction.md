# API / DDL 提取规范参考

本文件定义 prd-parse 中 API 和 DDL 数据模型的提取标准。

---

## API 提取

### 识别来源
1. 扫描 PRD 中的 "Interface Design"、"API List"、"System Interactions" 等 section
2. 识别动词："call"、"request"、"fetch"、"submit"、"sync"
3. 从 User Story 反向推导：User Action → Frontend Event → Backend API

### 结构化输出

```yaml
- endpoint: /api/v1/xxx
  method: POST
  description: Feature description
  request:
    headers: [Authorization: Bearer xxx]
    params: {field: type, required: bool, validation: rule}
  response:
    success: {code: 200, data: {schema}}
    errors: [{code: 40001, message: "Business error description"}]
  auth: JWT/OAuth/API-Key
  rate_limit: 100/min
  idempotent: true/false
```

### 附加规格
- 统一错误码（业务码 + HTTP 状态码）
- 认证方式文档
- 限流策略
- API 版本策略（v1/v2 兼容性）
- Mock 数据定义（前端联调用）

### 完整性检查
- 提取 PRD 中 **所有** API
- 创建完整 API 清单：endpoint、method、params
- 验证每个功能都有 API 覆盖
- 检查 CRUD 完整性、验证、错误处理
- 映射 API 到微服务
- **阻断规则:** 如果缺少关键 API，STOP

---

## DDL 提取

### 识别数据实体
1. 扫描 PRD 中的 "Data Model"、"Entity Relationships"、"Field Specifications" 等 section
2. 识别名词："user"、"order"、"record"、"config" → 映射到表
3. 从 API request/response 字段反向推导表字段

### 结构化输出

```sql
CREATE TABLE table_name (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Primary key',
  -- Business fields (extract from PRD)
  field_name TYPE [NOT NULL] [DEFAULT x] COMMENT 'Field description',
  -- Audit fields (mandatory)
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created time',
  updated_at DATETIME ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated time',
  deleted_at DATETIME NULL COMMENT 'Soft delete time',
  -- Index design
  INDEX idx_xxx (field1, field2) COMMENT 'Index purpose'
) COMMENT='Table description';
```

### 附加规格
- 索引策略：主键、唯一、复合、覆盖索引
- 分片策略：按时间/ID/租户（如适用）
- 字段约束：NOT NULL、DEFAULT、UNIQUE
- 枚举值：字典表或字段注释
- 大字段处理：TEXT/BLOB 存储方案
- 数据迁移：新字段兼容计划

### 完整性检查
- 提取 PRD 中 **所有** 数据表
- 创建完整表清单：字段、类型、约束
- 验证每个功能都有数据模型覆盖
- 检查字段完整性、索引、约束
- 映射表关系和外键
- **阻断规则:** 如果缺少关键表，STOP
