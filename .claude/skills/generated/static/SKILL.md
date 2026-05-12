---
name: static
description: "Skill for the Static area of zero-admin. 1324 symbols across 8 files."
---

# Static

1324 symbols | 8 files | Cohesion: 73%

## When to Use

- Working with code in `api/`
- Understanding how remove work
- Modifying static-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `api/front/static/swagger-ui-es-bundle-core.js` | createDeepLinkPath, escapeDeepLinkPath, getExtensions, safeBuildUrl, sanitizeUrl (+244) |
| `api/admin/static/swagger-ui-es-bundle-core.js` | createDeepLinkPath, escapeDeepLinkPath, getExtensions, safeBuildUrl, sanitizeUrl (+244) |
| `api/front/static/swagger-ui.js` | useComponent, useFn, keywords_Not, keywords_If, keywords_Then (+219) |
| `api/admin/static/swagger-ui.js` | useComponent, useFn, keywords_Not, keywords_If, keywords_Then (+219) |
| `api/front/static/swagger-ui-standalone-preset.js` | Iterable, KeyedIterable, IndexedIterable, SetIterable, isIterable (+183) |
| `api/admin/static/swagger-ui-standalone-preset.js` | is_EOL, is_WHITE_SPACE, is_WS_OR_EOL, is_FLOW_INDICATOR, generateError (+183) |
| `web-admin/src/app.tsx` | cleanupOrphanAntdOverlays |
| `flutter-mall/lib/utils/shared_preferences_util.dart` | remove |

## Entry Points

Start here when exploring this area:

- **`remove`** (Method) — `flutter-mall/lib/utils/shared_preferences_util.dart:68`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `remove` | Method | `flutter-mall/lib/utils/shared_preferences_util.dart` | 68 |
| `Iterable` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `KeyedIterable` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `IndexedIterable` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `SetIterable` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `isIterable` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `isKeyed` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `isIndexed` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `isAssociative` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `Seq` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `KeyedSeq` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `IndexedSeq` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `SetSeq` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `emptySequence` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `indexedSeqFromValue` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `deepEqual` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `mergeIntoMapWith` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `mergeIntoCollectionWith` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `mergeIntoListWith` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |
| `groupByFactory` | Function | `api/front/static/swagger-ui-standalone-preset.js` | 1 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Build → Remove` | cross_community | 6 |
| `Build → Remove` | cross_community | 5 |
| `Build → Remove` | cross_community | 5 |
| `Main → Remove` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Digitalcardmint | 4 calls |

## How to Explore

1. `gitnexus_context({name: "remove"})` — see callers and callees
2. `gitnexus_query({query: "static"})` — find related execution flows
3. Read key files listed above for implementation details
