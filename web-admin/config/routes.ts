export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        path: '/user',
        routes: [
          {
            name: 'login',
            path: '/user/login',
            component: './user/login',
          },
        ],
      },
      {
        component: './404',
      },
    ],
  },
  {
    path: '/welcome',
    name: 'welcome',
    icon: 'smile',
    component: './Welcome',
  },
  {
    path: '/admin',
    name: 'admin',
    icon: 'crown',
    access: 'canAdmin',
    component: './Admin',
    routes: [
      {
        path: '/admin/sub-page',
        name: 'sub-page',
        icon: 'smile',
        component: './Welcome',
      },
      {
        component: './404',
      },
    ],
  },

  {
    path: '/',
    redirect: '/welcome',
  },
  {
    path: '/system',
    name: '系统管理',
    icon: 'crown',
    routes: [
      {
        name: '用户列表',
        icon: 'table',
        path: '/system/user/list',
        component: './system/user',
      },
      {
        name: '角色列表',
        icon: 'table',
        path: '/system/role/list',
        component: './system/role',
      },
      {
        name: '菜单列表',
        icon: 'table',
        path: '/system/menu/list',
        component: './system/menu',
      },
      {
        name: '机构列表',
        icon: 'table',
        path: '/system/dept/list',
        component: './system/dept',
      },
      {
        name: '字典列表',
        icon: 'table',
        path: '/system/dict/list',
        component: './system/dict',
      },
      {
        name: '岗位管理',
        icon: 'table',
        path: '/system/post/list',
        component: './system/post',
      },
      {
        name: '通知管理',
        icon: 'table',
        path: '/system/notice/list',
        component: './system/notice',
      },
      {
        name: '租户治理',
        icon: 'table',
        path: '/system/tenant/list',
        component: './system/tenant',
      },
      {
        name: '商户治理',
        icon: 'table',
        path: '/system/merchant/list',
        component: './system/merchant',
      },
      {
        name: '模板治理',
        icon: 'table',
        path: '/system/channelIntegrationTemplate/list',
        component: './system/channel_integration_template',
      },
      {
        name: '系统配置',
        icon: 'setting',
        path: '/system/systemConfig',
        component: './system/system_config',
      },
      {
        path: '/system/channel-integration-template/list',
        redirect: '/system/channelIntegrationTemplate/list',
        hideInMenu: true,
      },
      {
        path: '/system/system-config',
        redirect: '/system/systemConfig',
        hideInMenu: true,
      },
      // {
      //   name: '参数管理',
      //   icon: 'table',
      //   path: '/system/param/list',
      //   component: './system/param',
      // },
    ],
  },
  {
    path: '/log',
    name: '日志管理',
    icon: 'crown',
    routes: [
      {
        name: '登录日志',
        icon: 'table',
        path: '/log/loginLog/list',
        component: './log/LoginLog',
      },
      {
        name: '操作日志',
        icon: 'table',
        path: '/log/sysLog/list',
        component: './log/OperateLog',
      },
      {
        name: '治理审计中心',
        icon: 'table',
        path: '/log/auditCenter/list',
        component: './log/AuditCenter',
      },
    ],
  },
  {
    path: '/ums',
    name: '会员管理',
    icon: 'crown',
    routes: [
      {
        name: '会员列表',
        icon: 'table',
        path: '/ums/member/list',
        component: './ums/MemberInfo',
      },
      {
        name: '会员等级',
        icon: 'table',
        path: '/ums/memberLevel/list',
        component: './ums/MemberLevel',
      },
      // {
      //   name: '会员地址',
      //   icon: 'table',
      //   path: '/ums/memberAddress/list',
      //   component: './ums/member_address',
      // },
      // {
      //   name: '登录记录',
      //   icon: 'table',
      //   path: '/ums/memberLoginLog/list',
      //   component: './ums/member_login_log',
      // },
      {
        name: '积分消费设置',
        icon: 'table',
        path: '/ums/integrationConsumeSetting/list',
        component: './ums/MemberConsumeSetting',
      },
      {
        name: '成长规则',
        icon: 'table',
        path: '/ums/memberRule/list',
        component: './ums/MemberRuleSetting',
      },
      {
        name: '会员标签',
        icon: 'table',
        path: '/ums/memberTag/list',
        component: './ums/MemberTag',
      },
      {
        name: '会员任务',
        icon: 'table',
        path: '/ums/memberTask/list',
        component: './ums/MemberTask',
      },
      {
        name: '会员任务',
        icon: 'table',
        path: '/ums/statistics/list',
        component: './ums/MemberStatisticsInfo',
      },
    ],
  },
  {
    path: '/pms',
    name: '商品管理',
    icon: 'crown',
    routes: [
      {
        name: '商品列表',
        icon: 'table',
        path: '/pms/product/list',
        component: './pms/product',
      },
      {
        name: '商品分类',
        icon: 'table',
        path: '/pms/productCategory/list',
        component: './pms/ProductCategory',
      },
      {
        name: '品牌管理',
        icon: 'table',
        path: '/pms/productBrand/list',
        component: './pms/ProductBrand',
      },
      {
        name: '属性分组',
        icon: 'table',
        path: '/pms/attributeGroup/list',
        component: './pms/ProductAttributeGroup',
      },

      {
        name: '商品属性',
        icon: 'table',
        path: '/pms/attribute/list',
        component: './pms/ProductAttribute',
      },
      {
        name: '属性规格',
        icon: 'table',
        path: '/pms/productSpec/list',
        component: './pms/ProductSpec',
      },

      {
        name: '属性规格值',
        icon: 'table',
        path: '/pms/productSpecValue/list',
        component: './pms/ProductSpecValue',
      },
      {
        name: '商品spu',
        icon: 'table',
        path: '/pms/productSpu/list',
        component: './pms/ProductSpu',
      },

      {
        name: '商品sku',
        icon: 'table',
        path: '/pms/productSku/list',
        component: './pms/ProductSku',
      },
    ],
  },
  {
    path: '/oms',
    name: '订单管理',
    icon: 'crown',
    routes: [
      {
        name: '订单列表',
        icon: 'table',
        path: '/oms/order/list',
        component: './oms/order',
      },
      {
        name: '订单设置',
        icon: 'table',
        path: '/oms/orderSetting/list',
        component: './oms/OrderSetting',
      },
      {
        name: '退货申请',
        icon: 'table',
        path: '/oms/orderReturnApply/list',
        component: './oms/order_return_apply',
      },
      {
        name: '退货原因',
        icon: 'table',
        path: '/oms/orderReturnReason/list',
        component: './oms/OrderReturnReason',
      },
      {
        name: '公司地址',
        icon: 'table',
        path: '/oms/companyAddress/list',
        component: './oms/CompanyAddress',
      },
    ],
  },
  {
    path: '/sms',
    name: '营销管理',
    icon: 'crown',
    routes: [
      {
        path: '/sms/flashPromotion/list',
        redirect: '/sms/seckillActivity/list',
      },
      {
        name: '品牌推荐',
        icon: 'table',
        path: '/sms/homeBrand/list',
        component: './sms/HomeBrand',
      },
      {
        name: '新品推荐',
        icon: 'table',
        path: '/sms/homeNewProduct/list',
        component: './sms/HomeNewProduct',
      },
      {
        name: '人气推荐',
        icon: 'table',
        path: '/sms/homeRecommendProduct/list',
        component: './sms/HomeRecommendProduct',
      },
      {
        name: '专题推荐',
        icon: 'table',
        path: '/sms/homeRecommendSubject/list',
        component: './sms/HomeRecommendSubject',
      },
      {
        name: '广告列表',
        icon: 'table',
        path: '/sms/homeAdvertise/list',
        component: './sms/HomeAdvertise',
      },
      {
        name: '运营数据看板',
        icon: 'table',
        path: '/sms/operateDashboard/list',
        component: './sms/OperateDashboard',
      },
      {
        name: '优惠券',
        icon: 'table',
        path: '/sms/Coupon/list',
        component: './sms/Coupon',
      },
      {
        name: '秒杀活动',
        icon: 'table',
        path: '/sms/seckillActivity/list',
        component: './sms/SeckillActivity',
      },
      {
        name: '抽卡活动',
        icon: 'table',
        path: '/sms/digitalCardActivity/list',
        component: './sms/DigitalCardActivity',
      },
      {
        name: '提货卡链路',
        icon: 'table',
        path: '/sms/digitalCardChain/list',
        component: './sms/DigitalCardChainMonitor',
      },
      {
        name: '提货卡资产',
        icon: 'table',
        path: '/sms/digitalCardAsset/list',
        component: './sms/DigitalCardAssetWorkbench',
      },
      {
        // Story 10.7 Task 2.8 / S3+S4 — 后台提货单管理
        name: '提货单管理',
        icon: 'gift',
        path: '/sms/digitalCardRedemptionOrder/list',
        component: './sms/DigitalCardRedemptionOrder',
      },
      {
        // Story 10.7 Task 8.8 / S5+S6 — 后台分享凭证管理
        name: '分享凭证管理',
        icon: 'share-alt',
        path: '/sms/digitalCardClaimToken/list',
        component: './sms/DigitalCardClaimToken',
      },
      {
        // Story 10.7 Task 9.4 — 后台合规审计跨资产检索
        name: '转赠审计记录',
        icon: 'audit',
        path: '/sms/digitalCardTransferAudit/list',
        component: './sms/DigitalCardTransferAudit',
      },
      {
        name: '实体卡履约',
        icon: 'table',
        path: '/sms/digitalCardPhysicalFulfillment/list',
        component: './sms/DigitalCardPhysicalFulfillmentWorkbench',
      },
      {
        name: '商品发卡规则',
        icon: 'table',
        path: '/sms/productFulfillmentRule/list',
        component: './sms/ProductFulfillmentRule',
      },
      {
        name: '卡片模板',
        icon: 'table',
        path: '/sms/cardTemplate/list',
        component: './sms/CardTemplate',
      },
    ],
  },
  {
    path: '/cms',
    name: '内容管理',
    icon: 'crown',
    routes: [
      {
        name: '商品专题',
        icon: 'table',
        path: '/cms/subject/list',
        component: './cms/Subject',
      },
      {
        name: '专题分类',
        icon: 'table',
        path: '/cms/subjectCategory/list',
        component: './cms/SubjectCategory',
      },
      {
        name: '商品优选',
        icon: 'table',
        path: '/cms/preferredArea/list',
        component: './cms/PreferredArea',
      },
    ],
  },
  {
    component: './404',
  },
];
