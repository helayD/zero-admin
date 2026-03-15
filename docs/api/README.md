# API 参考文档

## 概述

Zero-Admin 提供两套 API：
- **admin-api**: 管理端 API
- **front-api**: 用户端 API

## Base URL

| 环境 | admin-api | front-api |
|------|-----------|-----------|
| 开发 | http://localhost:8000 | http://localhost:9999 |
| 生产 | 根据部署配置 | 根据部署配置 |

## 认证

使用 JWT Token 认证：

```
Authorization: Bearer <token>
```

## API 模块

### 1. 系统管理 (sys)

| 模块 | 描述 |
|------|------|
| /sys/user | 用户管理 |
| /sys/role | 角色管理 |
| /sys/menu | 菜单管理 |
| /sys/dept | 部门管理 |
| /sys/post | 岗位管理 |
| /sys/dict/type | 字典类型 |
| /sys/dict/item | 字典数据 |

### 2. 会员管理 (ums)

| 模块 | 描述 |
|------|------|
| /ums/member | 会员管理 |
| /ums/address | 收货地址 |
| /ums/growth | 成长值 |
| /ums/points | 积分 |

### 3. 商品管理 (pms)

| 模块 | 描述 |
|------|------|
| /pms/product | 商品管理 |
| /pms/brand | 品牌管理 |
| /pms/category | 分类管理 |
| /pms/attribute | 属性管理 |
| /pms/spec | 规格管理 |

### 4. 订单管理 (oms)

| 模块 | 描述 |
|------|------|
| /oms/order | 订单管理 |
| /oms/cart | 购物车 |
| /oms/return | 退货管理 |

### 5. 营销管理 (sms)

| 模块 | 描述 |
|------|------|
| /sms/coupon | 优惠券 |
| /sms/seckill | 秒杀 |
| /sms/flash | 限时折扣 |

### 6. 内容管理 (cms)

| 模块 | 描述 |
|------|------|
| /cms/subject | 专题管理 |
| /cms/area | 优选专区 |

## 通用响应格式

### 成功
```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

### 失败
```json
{
  "code": 1,
  "msg": "error message",
  "data": null
}
```

## 分页参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| pageNum | 页码 | 1 |
| pageSize | 每页数量 | 20 |

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 失败 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |
