# 前端文档

## 技术栈

- React 17
- Ant Design Pro 5.2.0
- Umi 3.5.0
- TypeScript 4.5.0

## 目录结构

```
web-admin/
├── src/
│   ├── components/        # 公共组件
│   ├── pages/            # 页面组件
│   │   ├── cms/         # 内容管理
│   │   ├── oms/         # 订单管理
│   │   ├── pms/         # 商品管理
│   │   ├── sms/         # 营销管理
│   │   ├── sys/         # 系统管理
│   │   └── log/         # 日志
│   ├── services/        # API 服务
│   ├── utils/          # 工具函数
│   └── global.tsx      # 全局配置
├── config/             # 配置文件
└── public/             # 静态资源
```

## 页面模块

### 系统管理 (sys)
- 用户管理
- 角色管理
- 菜单管理
- 部门管理
- 岗位管理
- 字典管理
- 地区管理
- 登录日志
- 操作日志

### 订单管理 (oms)
- 订单列表
- 订单设置
- 退货申请
- 公司地址

### 商品管理 (pms)
- 商品列表
- 商品分类
- 商品品牌
- 商品属性
- 商品规格
- SKU 管理

### 营销管理 (sms)
- 优惠券
- 秒杀活动
- 限时折扣
- 首页广告
- 品牌推荐
- 新品推荐

### 内容管理 (cms)
- 专题管理
- 专题分类
- 优选专区
- 话题管理

## 开发命令

```bash
# 开发
npm run dev

# 构建
npm run build

# 测试
npm run test

# 代码检查
npm run lint
```
