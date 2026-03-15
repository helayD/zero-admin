---
project_name: zero-admin
date: 2026-03-11
status: complete
type: brownfield
parts_count: 4
---

# Zero-Admin 项目文档

## 项目概览

Zero-Admin 是一个基于 go-zero 框架的企业级后台管理系统，支持完整的电商功能。

## 文档结构

### 主要部分

| 部分 | 类型 | 描述 |
|------|------|------|
| [项目概览](project-overview.md) | 文档 | 项目整体架构和技术栈 |
| [后端服务](backend/) | 文档 | Go 微服务后端文档 |
| [前端](frontend/) | 文档 | React + Ant Design Pro 前端 |
| [移动端](mobile/) | 文档 | Flutter 电商移动端 |
| [API 参考](api/) | 文档 | API 接口文档 |

### 微服务模块

| 模块 | 描述 |
|------|------|
| sys | 系统管理（用户、角色、菜单、部门） |
| ums | 会员管理 |
| pms | 商品管理 |
| oms | 订单管理 |
| sms | 营销管理 |
| cms | 内容管理 |
| search | 搜索服务 |

---

## 技术栈

### 后端
- Go 1.25
- go-zero 1.9.3
- GORM + MySQL
- MongoDB
- Redis
- gRPC
- Elasticsearch

### 前端
- React 17
- Ant Design Pro 5.2
- Umi 3.5
- TypeScript

### 移动端
- Flutter

---

_Last Updated: 2026-03-11_
