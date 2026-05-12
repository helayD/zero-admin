---
name: channelintegrationtemplateservice
description: "Skill for the Channelintegrationtemplateservice area of zero-admin. 55 symbols across 12 files."
---

# Channelintegrationtemplateservice

55 symbols | 12 files | Cohesion: 63%

## When to Use

- Working with code in `rpc/`
- Understanding how NormalizeStatus, ValidateStatusTransition, NewUpdateChannelIntegrationTemplateStatusLogic work
- Modifying channelintegrationtemplateservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/sys/internal/logic/channelintegrationtemplateservice/binding_helpers.go` | mergeImpactScopeConfigWithSummary, ensureTemplateScopeSubjectExists, queryBindingSubjects, queryTenantBindingSubjects, queryMerchantBindingSubjects (+13) |
| `rpc/sys/internal/server/channelintegrationtemplateservice/channelintegrationtemplateserviceserver.go` | UpdateChannelIntegrationTemplateStatus, QueryChannelIntegrationTemplateList, UpdateChannelIntegrationTemplate, CreateChannelIntegrationTemplate, QueryChannelIntegrationTemplateDetail |
| `rpc/sys/client/channelintegrationtemplateservice/channelintegrationtemplateservice.go` | UpdateChannelIntegrationTemplate, QueryChannelIntegrationTemplateDetail, QueryChannelIntegrationTemplateList, CreateChannelIntegrationTemplate, UpdateChannelIntegrationTemplateStatus |
| `rpc/sys/internal/logic/channelintegrationtemplateservice/channelintegrationtemplate_logic_test.go` | newChannelIntegrationTemplateTestSvc, TestQueryChannelIntegrationTemplateListSupportsAliasFilter, TestQueryChannelIntegrationTemplateListGracefullyHandlesMissingMerchantTable, TestCreateChannelIntegrationTemplateNormalizesAliasAndSecretRefs, TestUpdateChannelIntegrationTemplateStatusRejectsArchivedReenable |
| `pkg/channeltemplate/template.go` | NormalizeStatus, ValidateStatusTransition, ValidateSecretRefMap, ValidateIntentContracts |
| `rpc/sys/internal/logic/channelintegrationtemplateservice/service_helpers.go` | normalizeJSONObjectString, normalizeSecretRefConfig, normalizeIntentContractConfig, normalizeStatusFilter |
| `rpc/sys/sysclient/channel_integration_template_grpc.pb.go` | NewChannelIntegrationTemplateServiceClient, QueryChannelIntegrationTemplateList, CreateChannelIntegrationTemplate, UpdateChannelIntegrationTemplateStatus |
| `rpc/sys/internal/logic/channelintegrationtemplateservice/updatechannelintegrationtemplatestatuslogic.go` | NewUpdateChannelIntegrationTemplateStatusLogic, UpdateChannelIntegrationTemplateStatus |
| `rpc/sys/internal/logic/channelintegrationtemplateservice/querychannelintegrationtemplatelistlogic.go` | NewQueryChannelIntegrationTemplateListLogic, QueryChannelIntegrationTemplateList |
| `rpc/sys/internal/logic/channelintegrationtemplateservice/updatechannelintegrationtemplatelogic.go` | NewUpdateChannelIntegrationTemplateLogic, UpdateChannelIntegrationTemplate |

## Entry Points

Start here when exploring this area:

- **`NormalizeStatus`** (Function) — `pkg/channeltemplate/template.go:82`
- **`ValidateStatusTransition`** (Function) — `pkg/channeltemplate/template.go:91`
- **`NewUpdateChannelIntegrationTemplateStatusLogic`** (Function) — `rpc/sys/internal/logic/channelintegrationtemplateservice/updatechannelintegrationtemplatestatuslogic.go:24`
- **`ValidateSecretRefMap`** (Function) — `pkg/channeltemplate/template.go:133`
- **`ValidateIntentContracts`** (Function) — `pkg/channeltemplate/template.go:151`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NormalizeStatus` | Function | `pkg/channeltemplate/template.go` | 82 |
| `ValidateStatusTransition` | Function | `pkg/channeltemplate/template.go` | 91 |
| `NewUpdateChannelIntegrationTemplateStatusLogic` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/updatechannelintegrationtemplatestatuslogic.go` | 24 |
| `ValidateSecretRefMap` | Function | `pkg/channeltemplate/template.go` | 133 |
| `ValidateIntentContracts` | Function | `pkg/channeltemplate/template.go` | 151 |
| `NewQueryChannelIntegrationTemplateListLogic` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/querychannelintegrationtemplatelistlogic.go` | 21 |
| `NewChannelIntegrationTemplateServiceClient` | Function | `rpc/sys/sysclient/channel_integration_template_grpc.pb.go` | 30 |
| `TestQueryChannelIntegrationTemplateListSupportsAliasFilter` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/channelintegrationtemplate_logic_test.go` | 167 |
| `TestQueryChannelIntegrationTemplateListGracefullyHandlesMissingMerchantTable` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/channelintegrationtemplate_logic_test.go` | 197 |
| `NewUpdateChannelIntegrationTemplateLogic` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/updatechannelintegrationtemplatelogic.go` | 24 |
| `TestCreateChannelIntegrationTemplateNormalizesAliasAndSecretRefs` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/channelintegrationtemplate_logic_test.go` | 130 |
| `TestUpdateChannelIntegrationTemplateStatusRejectsArchivedReenable` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/channelintegrationtemplate_logic_test.go` | 238 |
| `NewCreateChannelIntegrationTemplateLogic` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/createchannelintegrationtemplatelogic.go` | 23 |
| `NewQueryChannelIntegrationTemplateDetailLogic` | Function | `rpc/sys/internal/logic/channelintegrationtemplateservice/querychannelintegrationtemplatedetaillogic.go` | 20 |
| `UpdateChannelIntegrationTemplateStatus` | Method | `rpc/sys/internal/server/channelintegrationtemplateservice/channelintegrationtemplateserviceserver.go` | 29 |
| `UpdateChannelIntegrationTemplateStatus` | Method | `rpc/sys/internal/logic/channelintegrationtemplateservice/updatechannelintegrationtemplatestatuslogic.go` | 28 |
| `QueryChannelIntegrationTemplateList` | Method | `rpc/sys/internal/server/channelintegrationtemplateservice/channelintegrationtemplateserviceserver.go` | 39 |
| `QueryChannelIntegrationTemplateList` | Method | `rpc/sys/internal/logic/channelintegrationtemplateservice/querychannelintegrationtemplatelistlogic.go` | 25 |
| `UpdateChannelIntegrationTemplate` | Method | `rpc/sys/client/channelintegrationtemplateservice/channelintegrationtemplateservice.go` | 45 |
| `QueryChannelIntegrationTemplateDetail` | Method | `rpc/sys/client/channelintegrationtemplateservice/channelintegrationtemplateservice.go` | 55 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Channeltemplate | 3 calls |
| Userservice | 2 calls |
| Drawactivityservice | 1 calls |
| Operatefunnel | 1 calls |
| Audit | 1 calls |

## How to Explore

1. `gitnexus_context({name: "NormalizeStatus"})` — see callers and callees
2. `gitnexus_query({query: "channelintegrationtemplateservice"})` — find related execution flows
3. Read key files listed above for implementation details
