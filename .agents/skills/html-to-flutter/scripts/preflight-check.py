#!/usr/bin/env python3
"""
HTML → Flutter 迁移预检脚本
验证迁移所需的文件和环境是否就绪。

用法: python3 .claude/skills/html-to-flutter/scripts/preflight-check.py <page_name>
示例: python3 .claude/skills/html-to-flutter/scripts/preflight-check.py home
"""

import sys
import os
from pathlib import Path

# 页面→模块映射
PAGE_MODULE_MAP = {
    # P0: MVP
    "home": {"module": "home", "page": "home_page.dart", "priority": "P0", "status": "refactor"},
    "workers": {"module": "ai_workers", "page": "workers_page.dart", "priority": "P0", "status": "refactor"},
    "worker-detail": {"module": "ai_workers", "page": "worker_detail_page.dart", "priority": "P0", "status": "refactor"},
    "chat": {"module": "ai_workers", "page": "worker_chat_page.dart", "priority": "P0", "status": "reuse"},
    "briefing": {"module": "briefing", "page": "briefing_page.dart", "priority": "P0", "status": "new"},
    "dashboard": {"module": "dashboard", "page": "dashboard_page.dart", "priority": "P0", "status": "refactor"},
    "onboarding": {"module": "onboarding", "page": "onboarding_page.dart", "priority": "P0", "status": "new"},
    "track-assessment": {"module": "onboarding", "page": "track_assessment_page.dart", "priority": "P0", "status": "new"},
    # P1: 业务闭环
    "notifications": {"module": "notifications", "page": "notifications_page.dart", "priority": "P1", "status": "new"},
    "content": {"module": "content", "page": "content_page.dart", "priority": "P1", "status": "new"},
    "content-review": {"module": "content", "page": "content_review_page.dart", "priority": "P1", "status": "new"},
    "crm": {"module": "crm", "page": "crm_page.dart", "priority": "P1", "status": "refactor"},
    "projects": {"module": "projects", "page": "projects_page.dart", "priority": "P1", "status": "new"},
    "profile": {"module": "profile", "page": "profile_page.dart", "priority": "P1", "status": "extend"},
    "portfolio": {"module": "profile", "page": "portfolio_page.dart", "priority": "P1", "status": "new"},
    "credits": {"module": "finance", "page": "credits_page.dart", "priority": "P1", "status": "new"},
    "subscription": {"module": "subscription", "page": "subscription_page.dart", "priority": "P1", "status": "extend"},
    "demands": {"module": "marketplace", "page": "demands_page.dart", "priority": "P1", "status": "new"},
    "market": {"module": "marketplace", "page": "marketplace_page.dart", "priority": "P1", "status": "new"},
    "enterprise": {"module": "marketplace", "page": "enterprise_page.dart", "priority": "P1", "status": "new"},
    # P2: 高级功能
    "finance": {"module": "finance", "page": "finance_page.dart", "priority": "P2", "status": "new"},
    "escrow": {"module": "finance", "page": "escrow_page.dart", "priority": "P2", "status": "new"},
    "worker-collaboration": {"module": "ai_workers", "page": "worker_collaboration_page.dart", "priority": "P2", "status": "new"},
    "worker-memory": {"module": "ai_workers", "page": "worker_memory_page.dart", "priority": "P2", "status": "new"},
    "automation": {"module": "automation", "page": "automation_page.dart", "priority": "P2", "status": "new"},
    "coze-workflows": {"module": "automation", "page": "coze_workflows_page.dart", "priority": "P2", "status": "new"},
    "comfyui-workflows": {"module": "automation", "page": "comfyui_workflows_page.dart", "priority": "P2", "status": "new"},
    "mcp-tools": {"module": "automation", "page": "mcp_tools_page.dart", "priority": "P2", "status": "new"},
    "knowledge": {"module": "knowledge", "page": "knowledge_page.dart", "priority": "P2", "status": "transform"},
    "messaging-channels": {"module": "messaging_channels", "page": "messaging_channels_page.dart", "priority": "P2", "status": "new"},
    "brand": {"module": "profile", "page": "brand_page.dart", "priority": "P2", "status": "new"},
    # P3: 全功能
    "phone-os": {"module": "phone_os", "page": "phone_os_page.dart", "priority": "P3", "status": "extend"},
    "phone-os-edit": {"module": "phone_os", "page": "phone_os_edit_page.dart", "priority": "P3", "status": "new"},
    "phone-tasks": {"module": "phone_os", "page": "phone_tasks_page.dart", "priority": "P3", "status": "new"},
    "credit-score": {"module": "credit_score", "page": "credit_score_page.dart", "priority": "P3", "status": "new"},
    "ai-quality": {"module": "ai_quality", "page": "ai_quality_page.dart", "priority": "P3", "status": "new"},
    "comply": {"module": "ai_quality", "page": "comply_page.dart", "priority": "P3", "status": "new"},
}

STATUS_LABELS = {
    "new": "❌ 新建模块",
    "refactor": "🔄 重构现有模块",
    "extend": "🔄 扩展现有模块",
    "reuse": "🔄 复用现有组件",
    "transform": "🔄 改造现有模块",
}

REQUIRED_THEME_FILES = [
    "app_theme.dart",
    "color_schemes.dart",
    "brand_colors.dart",
    "amnt_tokens.dart",
    "design_tokens.dart",
    "typography_system.dart",
    "component_themes.dart",
    "theme_extensions.dart",
]


def find_project_root():
    """向上查找包含 flutter_client/ 的项目根目录"""
    current = Path.cwd()
    for _ in range(10):
        if (current / "flutter_client").is_dir():
            return current
        if (current / "wib").is_dir():
            return current
        parent = current.parent
        if parent == current:
            break
        current = parent
    # 如果找不到，尝试常见路径
    for candidate in [
        Path.cwd(),
        Path.cwd().parent,
        Path.cwd().parent.parent,
    ]:
        if (candidate / "flutter_client").is_dir():
            return candidate
    return Path.cwd()


def main():
    if len(sys.argv) < 2:
        print("❌ 错误: 请提供页面名称")
        print(f"用法: python3 {sys.argv[0]} <page_name>")
        print(f"\n可用页面 ({len(PAGE_MODULE_MAP)}):")
        for priority in ["P0", "P1", "P2", "P3"]:
            pages = [k for k, v in PAGE_MODULE_MAP.items() if v["priority"] == priority]
            print(f"  {priority}: {', '.join(pages)}")
        sys.exit(1)

    page_name = sys.argv[1].strip().lower()
    
    # 去掉 .html 后缀
    if page_name.endswith(".html"):
        page_name = page_name[:-5]

    root = find_project_root()
    errors = []
    warnings = []

    print(f"{'='*60}")
    print(f"  HTML → Flutter 迁移预检")
    print(f"  页面: {page_name}")
    print(f"{'='*60}")
    print()

    # 1. 检查页面是否在映射表中
    if page_name not in PAGE_MODULE_MAP:
        print(f"❌ 未知页面: {page_name}")
        print(f"\n可用页面:")
        for k in sorted(PAGE_MODULE_MAP.keys()):
            v = PAGE_MODULE_MAP[k]
            print(f"  {k:30s} → {v['module']}/{v['page']} [{v['priority']}]")
        sys.exit(1)

    mapping = PAGE_MODULE_MAP[page_name]
    module = mapping["module"]
    page_file = mapping["page"]
    priority = mapping["priority"]
    status = mapping["status"]

    print(f"PAGE_NAME     = {page_name}")
    print(f"MODULE        = {module}")
    print(f"PAGE_FILE     = {page_file}")
    print(f"PRIORITY      = {priority}")
    print(f"MODULE_STATUS = {STATUS_LABELS.get(status, status)}")
    print()

    # 2. 检查 HTML 原型文件
    html_path = root / "wib" / "ModelAi" / "app-prototype" / f"{page_name}.html"
    if html_path.exists():
        size = html_path.stat().st_size
        print(f"✅ HTML 原型: {html_path.relative_to(root)} ({size} bytes)")
    else:
        errors.append(f"HTML 原型文件不存在: {html_path.relative_to(root)}")
        print(f"❌ HTML 原型: {html_path.relative_to(root)} 不存在")

    # 3. 检查共享样式和导航
    css_path = root / "wib" / "ModelAi" / "app-prototype" / "app.css"
    nav_path = root / "wib" / "ModelAi" / "app-prototype" / "app-nav.js"
    
    if css_path.exists():
        print(f"✅ 共享样式: app.css")
    else:
        warnings.append("app.css 不存在")

    if nav_path.exists():
        print(f"✅ 共享导航: app-nav.js")
    else:
        warnings.append("app-nav.js 不存在")
    print()

    # 4. 检查 Flutter 目标路径
    flutter_base = root / "flutter_client" / "lib" / "src" / "features"
    module_path = flutter_base / module
    target_page = module_path / "presentation" / page_file

    print(f"TARGET_MODULE = flutter_client/lib/src/features/{module}/")
    print(f"TARGET_PAGE   = flutter_client/lib/src/features/{module}/presentation/{page_file}")
    
    if module_path.exists():
        item_count = sum(1 for _ in module_path.rglob("*") if _.is_file())
        print(f"✅ 目标模块已存在 ({item_count} 个文件)")
        if target_page.exists():
            warnings.append(f"目标页面已存在: {target_page.relative_to(root)}，可能需要覆盖或合并")
            print(f"⚠️  目标页面已存在: {page_file}")
    else:
        print(f"📁 目标模块不存在，将创建新模块")
    print()

    # 5. 检查 AMNT 设计系统
    theme_dir = root / "flutter_client" / "lib" / "src" / "core" / "theme"
    missing_themes = []
    for tf in REQUIRED_THEME_FILES:
        if not (theme_dir / tf).exists():
            missing_themes.append(tf)
    
    if not missing_themes:
        print(f"✅ AMNT 设计系统完整 ({len(REQUIRED_THEME_FILES)} 个核心文件)")
    else:
        errors.append(f"AMNT 设计系统文件缺失: {', '.join(missing_themes)}")
        print(f"❌ AMNT 设计系统文件缺失: {', '.join(missing_themes)}")
    print()

    # 6. 检查迁移进度文件
    progress_path = root / "flutter_client" / "docs" / "MIGRATION-PROGRESS.md"
    if progress_path.exists():
        print(f"✅ 迁移进度文件: MIGRATION-PROGRESS.md")
    else:
        warnings.append("MIGRATION-PROGRESS.md 不存在，需要创建")
        print(f"⚠️  迁移进度文件不存在")

    # 7. 检查路由配置
    routes_path = root / "flutter_client" / "lib" / "src" / "core" / "navigation" / "app_routes.dart"
    if routes_path.exists():
        print(f"✅ 路由配置: app_routes.dart")
    else:
        errors.append("app_routes.dart 不存在")
        print(f"❌ 路由配置文件不存在")
    print()

    # 8. 输出结果
    print(f"{'='*60}")
    if errors:
        print(f"❌ 预检失败 — {len(errors)} 个错误:")
        for e in errors:
            print(f"   • {e}")
        if warnings:
            print(f"\n⚠️  {len(warnings)} 个警告:")
            for w in warnings:
                print(f"   • {w}")
        print(f"{'='*60}")
        sys.exit(1)
    else:
        print(f"✅ 预检通过")
        if warnings:
            print(f"\n⚠️  {len(warnings)} 个警告:")
            for w in warnings:
                print(f"   • {w}")
        print(f"\n准备迁移:")
        print(f"  HTML: wib/ModelAi/app-prototype/{page_name}.html")
        print(f"  Flutter: flutter_client/lib/src/features/{module}/presentation/{page_file}")
        print(f"  操作: {STATUS_LABELS.get(status, status)}")
        print(f"{'='*60}")
        sys.exit(0)


if __name__ == "__main__":
    main()
