package digital_card_asset

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/svc"
)

type claimHintPageData struct {
	Token string
}

// H5 领取页（Story 10.7 闭环修复 2026-05-12）
//
// UX 设计系统（来自 ui-ux-pro-max ：mobile gift claim friendly premium warm）：
//   - 模式：Feature-Rich Showcase（hero + 卡片预览 + 表单 + CTA 重复）
//   - 风格：Liquid Glass（暖色渐变 hero + 磨砂玻璃卡片）
//   - 色板：礼物红 #DC2626 / 金 #D97706 / 粉 #EC4899 / 暖背景 #FFF1F2
//   - 字体：Calistoga（标题，温暖人文）+ Inter（正文）
//
// 流程：
//  1. 加载 → /api/digitalCard/validateClaimToken 预览卡片信息（含接收人手机号掩码）
//  2. 一步式表单：手机号 + 验证码（mock=123456，自动填入）
//  3. → /api/digitalCard/claimByMobile
//     - 命中接收人手机号 → 自动注册或登录 → 转赠 → 「领取成功 + 下载 App」
//     - 不命中           → 「领取失败 + 下载 App 注册」
//  4. App 下载链接来自 /api/config/appDownload（后台 sys_system_config）
//
// 监管约束：
//   - 不显示任何 chainType/chainStatus/tokenId/区块链相关字段
//   - 仅展示：模板名、卡面、有效期、接收人 mask、发放/合规状态
//
// 安全说明：所有业务校验仍在后端完成；HTML 仅是 UI 包装。
var claimHintPage = template.Must(template.New("digital-card-claim").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">
  <meta name="theme-color" content="#DC2626">
  <title>朋友送你一张提货卡</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Calistoga&family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    *,*::before,*::after{box-sizing:border-box}
    :root{
      --primary:#DC2626;--primary-dark:#991B1B;--accent:#EC4899;--gold:#D97706;
      --bg:#FFF1F2;--surface:#FFFFFF;--ink:#0F172A;--ink-muted:#475569;--ink-soft:#94A3B8;
      --border:#FAE4E4;--success:#059669;--warn:#A46500;--danger:#B42318;
      --radius-lg:18px;--radius-md:14px;--radius-sm:10px;
      --shadow-card:0 24px 60px -20px rgba(220,38,38,.18),0 8px 24px -12px rgba(15,23,42,.08);
      --font-display:'Calistoga',serif;
      --font-body:'Inter',-apple-system,BlinkMacSystemFont,'PingFang SC','Microsoft YaHei',sans-serif;
    }
    html,body{margin:0;padding:0;font-family:var(--font-body);background:var(--bg);color:var(--ink);-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}
    body{min-height:100vh;background:
      radial-gradient(1200px 600px at 50% -100px,#FFE2E5 0%,transparent 60%),
      radial-gradient(900px 500px at 100% 200px,#FEF3C7 0%,transparent 50%),
      var(--bg);
      background-attachment:fixed}

    main{max-width:480px;margin:0 auto;padding:0 20px 32px;min-height:100vh;display:flex;flex-direction:column}

    /* Hero */
    .hero{padding:32px 4px 20px;text-align:center;position:relative}
    .hero .badge{display:inline-flex;align-items:center;gap:6px;padding:6px 12px;font-size:12px;color:var(--primary);background:rgba(220,38,38,.08);border:1px solid rgba(220,38,38,.16);border-radius:999px;font-weight:600;letter-spacing:.4px}
    .hero h1{font-family:var(--font-display);font-size:30px;font-weight:400;line-height:1.25;margin:14px 0 8px;color:var(--ink)}
    .hero p{font-size:14px;line-height:1.6;color:var(--ink-muted);margin:0 12px}

    /* Card Preview */
    .card{display:flex;gap:16px;padding:18px;border-radius:var(--radius-lg);background:rgba(255,255,255,.78);backdrop-filter:saturate(180%) blur(20px);-webkit-backdrop-filter:saturate(180%) blur(20px);border:1px solid var(--border);box-shadow:var(--shadow-card);margin-top:18px;position:relative;overflow:hidden}
    .card::before{content:'';position:absolute;top:0;left:0;right:0;height:3px;background:linear-gradient(90deg,var(--primary) 0%,var(--gold) 50%,var(--accent) 100%)}
    .card .face{width:104px;height:140px;border-radius:var(--radius-md);background:linear-gradient(135deg,#FCE7C8 0%,#F4CBA0 100%);display:flex;align-items:center;justify-content:center;color:#8B6F3A;font-size:12px;flex-shrink:0;overflow:hidden;text-align:center;padding:8px;box-shadow:0 8px 20px -10px rgba(180,120,40,.4)}
    .card .face img{width:100%;height:100%;border-radius:var(--radius-md);object-fit:cover}
    .card .face .ph-icon{width:28px;height:28px;opacity:.5;margin-bottom:6px}
    .card .meta{flex:1;display:flex;flex-direction:column;justify-content:space-between;min-width:0}
    .card .meta h2{font-family:var(--font-display);font-size:18px;font-weight:400;margin:0 0 8px;color:var(--ink);line-height:1.3;word-break:break-all;overflow-wrap:anywhere;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
    .card .meta-row{display:flex;align-items:center;gap:6px;font-size:12px;color:var(--ink-muted);line-height:1.5;margin-bottom:4px}
    .card .meta-row svg{width:14px;height:14px;color:var(--ink-soft);flex-shrink:0}

    /* Target hint chip */
    .target-hint{margin-top:14px;padding:12px 14px;background:linear-gradient(135deg,rgba(217,119,6,.08) 0%,rgba(220,38,38,.06) 100%);border:1px solid rgba(217,119,6,.2);color:var(--warn);border-radius:var(--radius-md);font-size:13px;line-height:1.5;display:flex;align-items:flex-start;gap:8px}
    .target-hint svg{width:16px;height:16px;flex-shrink:0;margin-top:2px}
    .target-hint strong{color:var(--primary);font-weight:700;letter-spacing:.5px}

    /* Form */
    .form{margin-top:18px;padding:20px;background:rgba(255,255,255,.85);backdrop-filter:saturate(180%) blur(20px);-webkit-backdrop-filter:saturate(180%) blur(20px);border:1px solid var(--border);border-radius:var(--radius-lg);box-shadow:0 12px 36px -16px rgba(15,23,42,.08)}
    .form-row{margin-bottom:14px}
    .form-row:last-child{margin-bottom:0}
    .label{display:block;font-size:13px;color:var(--ink-muted);font-weight:500;margin-bottom:6px}
    .input{width:100%;height:48px;padding:0 14px;font-size:16px;font-family:var(--font-body);color:var(--ink);background:#fff;border:1.5px solid var(--border);border-radius:var(--radius-md);outline:none;transition:border-color .2s ease,box-shadow .2s ease;-webkit-appearance:none}
    .input:focus{border-color:var(--primary);box-shadow:0 0 0 3px rgba(220,38,38,.12)}
    .input-row{display:flex;gap:8px;align-items:stretch}
    .input-row .input{flex:1}
    .code-btn{flex-shrink:0;padding:0 16px;font-size:13px;font-weight:600;color:var(--primary);background:#fff;border:1.5px solid var(--primary);border-radius:var(--radius-md);cursor:pointer;font-family:var(--font-body);white-space:nowrap;transition:background .2s,color .2s}
    .code-btn:hover:not(:disabled){background:var(--primary);color:#fff}
    .code-btn:disabled{color:var(--ink-soft);border-color:var(--border);cursor:not-allowed;background:#fafafa}

    /* Buttons */
    .btn{width:100%;height:52px;display:flex;align-items:center;justify-content:center;gap:8px;font-size:16px;font-weight:700;font-family:var(--font-body);border-radius:var(--radius-md);border:0;cursor:pointer;text-decoration:none;letter-spacing:.3px;transition:transform .15s ease,box-shadow .2s ease,background .2s ease,color .2s ease}
    .btn:active{transform:translateY(1px)}
    .btn-primary{background:var(--ink);color:#fff;box-shadow:0 8px 20px -10px rgba(15,23,42,.4)}
    .btn-primary:hover:not(:disabled){background:var(--primary);box-shadow:0 12px 28px -12px rgba(220,38,38,.5)}
    .btn-primary:disabled{background:var(--ink-soft);cursor:not-allowed;box-shadow:none}
    .btn-outline{background:#fff;color:var(--ink);border:1.5px solid var(--border)}
    .btn-outline:hover{border-color:var(--ink);background:#fafafa}
    .btn svg{width:18px;height:18px}
    .form .btn-primary{margin-top:6px}

    /* Tip */
    .tip{margin-top:12px;font-size:13px;color:var(--ink-muted);line-height:1.5;min-height:18px;padding:0 4px;text-align:center}
    .tip.error{color:var(--danger);font-weight:500}
    .tip.success{color:var(--success);font-weight:500}

    /* Success / Fail blocks */
    .result-block{margin-top:18px;padding:24px 20px;background:rgba(255,255,255,.85);backdrop-filter:saturate(180%) blur(20px);-webkit-backdrop-filter:saturate(180%) blur(20px);border:1px solid var(--border);border-radius:var(--radius-lg);box-shadow:var(--shadow-card);text-align:center}
    .result-block .icon-wrap{width:64px;height:64px;margin:0 auto 14px;border-radius:50%;display:flex;align-items:center;justify-content:center;background:linear-gradient(135deg,var(--primary) 0%,var(--accent) 100%)}
    .result-block .icon-wrap.fail{background:linear-gradient(135deg,#fecaca 0%,#fca5a5 100%)}
    .result-block .icon-wrap svg{width:32px;height:32px;color:#fff}
    .result-block h3{font-family:var(--font-display);font-size:22px;font-weight:400;margin:0 0 8px;color:var(--ink)}
    .result-block p{font-size:14px;line-height:1.65;color:var(--ink-muted);margin:0 8px}

    /* App download block */
    .app-block{margin-top:18px;padding:20px;background:linear-gradient(135deg,#0F172A 0%,#1E293B 100%);border-radius:var(--radius-lg);color:#fff;box-shadow:var(--shadow-card);position:relative;overflow:hidden}
    .app-block::before{content:'';position:absolute;width:200px;height:200px;background:radial-gradient(circle,rgba(236,72,153,.3) 0%,transparent 70%);top:-80px;right:-60px;border-radius:50%}
    .app-block .app-header{display:flex;align-items:center;gap:12px;margin-bottom:14px;position:relative;z-index:1}
    .app-block .app-logo{width:44px;height:44px;border-radius:12px;background:linear-gradient(135deg,var(--primary) 0%,var(--gold) 100%);display:flex;align-items:center;justify-content:center;flex-shrink:0;box-shadow:0 6px 16px -6px rgba(220,38,38,.5)}
    .app-block .app-logo svg{width:22px;height:22px;color:#fff}
    .app-block .app-info{flex:1}
    .app-block .app-name{font-family:var(--font-display);font-size:18px;font-weight:400;line-height:1.2;margin:0}
    .app-block .app-tag{font-size:12px;color:rgba(255,255,255,.7);margin-top:2px;line-height:1.4}
    .app-block .app-buttons{display:grid;grid-template-columns:1fr 1fr;gap:10px;position:relative;z-index:1}
    .app-block .app-btn{display:flex;align-items:center;justify-content:center;gap:6px;height:44px;border-radius:var(--radius-sm);background:rgba(255,255,255,.12);border:1px solid rgba(255,255,255,.16);color:#fff;font-size:13px;font-weight:600;text-decoration:none;backdrop-filter:blur(8px);transition:background .2s}
    .app-block .app-btn:hover{background:rgba(255,255,255,.2)}
    .app-block .app-btn svg{width:16px;height:16px}

    /* Footer */
    .footer{margin-top:auto;padding:18px 4px 0;font-size:11px;color:var(--ink-soft);text-align:center;line-height:1.6}
    .footer a{color:var(--ink-muted);text-decoration:none}

    /* Loading skeleton */
    .skeleton{display:inline-block;height:1em;background:linear-gradient(90deg,#f3f4f6 25%,#e5e7eb 50%,#f3f4f6 75%);background-size:200% 100%;animation:skel 1.4s ease-in-out infinite;border-radius:4px}
    @keyframes skel{0%{background-position:200% 0}100%{background-position:-200% 0}}

    @media (prefers-reduced-motion: reduce){
      *,*::before,*::after{animation-duration:.01ms !important;transition-duration:.01ms !important}
    }
  </style>
</head>
<body>
  <main>
    <header class="hero">
      <span class="badge">
        <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path d="M10 18a8 8 0 100-16 8 8 0 000 16zm-2-7.586l1.293-1.293a1 1 0 011.414 0L12 10.414V8a1 1 0 112 0v5a1 1 0 01-1 1H8a1 1 0 110-2h2.586L8.293 9.707a1 1 0 010-1.414z"/></svg>
        九克城提货卡
      </span>
      <h1 id="title">朋友送你一张提货卡</h1>
      <p id="desc">正在为你准备...</p>
    </header>

    <section class="card" id="cardBox" style="display:none">
      <div class="face" id="cardFaceWrap">
        <svg class="ph-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <rect x="3" y="6" width="18" height="14" rx="2"/>
          <path d="M3 12h18M9 6V4a2 2 0 012-2h2a2 2 0 012 2v2"/>
        </svg>
      </div>
      <div class="meta">
        <h2 id="cardName">提货卡</h2>
        <div>
          <div class="meta-row" id="cardExpireRow">
            <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path d="M10 18a8 8 0 100-16 8 8 0 000 16zm.75-13a.75.75 0 00-1.5 0v5c0 .2.08.39.22.53l3 3a.75.75 0 101.06-1.06L10.75 9.69V5z"/></svg>
            <span id="cardExpire">-</span>
          </div>
          <div class="meta-row" id="cardSenderRow">
            <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"/></svg>
            <span id="cardSender">分享人：朋友</span>
          </div>
        </div>
      </div>
    </section>

    <div id="targetHint" class="target-hint" style="display:none">
      <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9 9a1 1 0 012 0v4a1 1 0 11-2 0V9zm1-4a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"/></svg>
      <div>此卡片仅限手机号 <strong id="targetMaskedText"></strong> 领取</div>
    </div>

    <form id="claimBlock" class="form" style="display:none" autocomplete="off" novalidate>
      <div class="form-row">
        <label class="label" for="mobile">手机号</label>
        <input id="mobile" class="input" maxlength="11" inputmode="numeric" autocomplete="tel" placeholder="请输入接收卡片的手机号">
      </div>
      <div class="form-row">
        <label class="label" for="code">验证码</label>
        <div class="input-row">
          <input id="code" class="input" maxlength="6" inputmode="numeric" autocomplete="one-time-code" placeholder="6 位数字">
          <button type="button" class="code-btn" id="codeBtn">获取验证码</button>
        </div>
      </div>
      <button type="button" class="btn btn-primary" id="claimBtn">
        <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path d="M5 9V7a5 5 0 0110 0v2h1a1 1 0 011 1v8a1 1 0 01-1 1H4a1 1 0 01-1-1v-8a1 1 0 011-1h1zm2 0h6V7a3 3 0 10-6 0v2z"/></svg>
        领取卡片
      </button>
      <p class="tip" id="tip"></p>
    </form>

    <section id="successBlock" class="result-block" style="display:none">
      <div class="icon-wrap" id="resultIconWrap">
        <svg id="resultIcon" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/></svg>
      </div>
      <h3 id="resultTitle">领取成功</h3>
      <p id="resultDesc">卡片已发放至你的账户。</p>
    </section>

    <section id="appBlock" class="app-block" style="display:none" aria-label="下载 App">
      <div class="app-header">
        <div class="app-logo">
          <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path d="M3 4a2 2 0 012-2h10a2 2 0 012 2v12a2 2 0 01-2 2H5a2 2 0 01-2-2V4zm5 4a1 1 0 100 2h4a1 1 0 100-2H8z"/></svg>
        </div>
        <div class="app-info">
          <div class="app-name" id="appName">下载 App</div>
          <div class="app-tag" id="appTag">查看你的提货卡，畅享专属权益</div>
        </div>
      </div>
      <div class="app-buttons">
        <a class="app-btn" id="androidLink" href="#" target="_blank" rel="noopener" aria-label="下载 Android 版">
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M17.523 15.342a1.07 1.07 0 110 2.142 1.07 1.07 0 010-2.142m-11.046 0a1.07 1.07 0 110 2.142 1.07 1.07 0 010-2.142m11.43-6.246l2.142-3.7a.448.448 0 00-.776-.448l-2.166 3.748A13.4 13.4 0 0012 7.27c-1.873 0-3.643.412-5.107 1.426L4.727 4.948a.448.448 0 00-.776.448l2.142 3.7C2.412 11.2 0 14.94 0 19.286h24c0-4.346-2.412-8.087-6.093-10.19"/></svg>
          Android
        </a>
        <a class="app-btn" id="iosLink" href="#" target="_blank" rel="noopener" aria-label="下载 iOS 版">
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/></svg>
          iOS
        </a>
      </div>
    </section>

    <p class="footer">
      由 <strong>九克城</strong> 提供 · 仅限指定接收人领取
    </p>
  </main>

  <script>
    var TOKEN = {{.Token}};
    var $ = function(id){ return document.getElementById(id); };

    var titleEl = $('title');
    var descEl = $('desc');
    var cardBox = $('cardBox');
    var cardFaceWrap = $('cardFaceWrap');
    var cardName = $('cardName');
    var cardSender = $('cardSender');
    var cardExpire = $('cardExpire');
    var targetHint = $('targetHint');
    var targetMaskedText = $('targetMaskedText');
    var claimBlock = $('claimBlock');
    var successBlock = $('successBlock');
    var resultIconWrap = $('resultIconWrap');
    var resultIcon = $('resultIcon');
    var resultTitle = $('resultTitle');
    var resultDesc = $('resultDesc');
    var appBlock = $('appBlock');
    var appName = $('appName');
    var appTag = $('appTag');
    var androidLink = $('androidLink');
    var iosLink = $('iosLink');
    var tip = $('tip');
    var codeBtn = $('codeBtn');
    var claimBtn = $('claimBtn');
    var mobileInput = $('mobile');
    var codeInput = $('code');

    var APP_CONFIG = null;

    var ICON_CHECK = '<svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/></svg>';
    var ICON_X = '<svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/></svg>';

    function setTip(msg, kind){
      tip.textContent = msg || '';
      tip.className = 'tip' + (kind ? ' ' + kind : '');
    }
    function isValidMobile(m){ return /^1[3-9]\d{9}$/.test(m); }

    function setCardFace(url){
      if (url) {
        var img = new Image();
        img.src = url;
        img.alt = '卡面';
        img.onload = function(){
          cardFaceWrap.innerHTML = '';
          cardFaceWrap.appendChild(img);
        };
        img.onerror = function(){
          cardFaceWrap.innerHTML = '<span style="font-size:11px;opacity:.6">卡面缺失</span>';
        };
      }
    }

    function showAppBlock(cfg){
      if (!cfg) return;
      APP_CONFIG = cfg;
      if (cfg.appName) appName.textContent = cfg.appName;
      if (cfg.tagline) appTag.textContent = cfg.tagline;
      if (cfg.androidUrl) {
        androidLink.href = cfg.androidUrl;
      } else {
        androidLink.style.display = 'none';
      }
      if (cfg.iosUrl) {
        iosLink.href = cfg.iosUrl;
      } else {
        iosLink.style.display = 'none';
      }
      appBlock.style.display = 'block';
    }

    function loadAppConfig(){
      return fetch('/api/config/appDownload').then(function(r){return r.json()}).then(function(b){
        if (b && b.code === 0 && b.data) return b.data;
        return null;
      }).catch(function(){ return null; });
    }

    function validateToken() {
      if (!TOKEN) {
        titleEl.textContent = '链接无效';
        descEl.textContent = '请联系好友重新分享。';
        loadAppConfig().then(showAppBlock);
        return;
      }
      fetch('/api/digitalCard/validateClaimToken', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({token: TOKEN})
      }).then(function(r){return r.json()}).then(function(body){
        if (!body || !body.data || body.data.valid !== true) {
          titleEl.textContent = '链接已失效';
          descEl.textContent = (body && body.data && body.data.failureReason) || '该分享链接已过期、被吊销或已被领取。';
          loadAppConfig().then(showAppBlock);
          return;
        }
        var info = body.data.token || {};
        var masked = body.data.targetMobileMasked || '';
        titleEl.textContent = '朋友送你一张提货卡';
        descEl.textContent = '请输入手机号 + 验证码完成领取';
        cardName.textContent = info.templateName || '提货卡';
        cardExpire.textContent = '有效期至 ' + (info.expireAt || '-');
        cardSender.textContent = '分享人：朋友';
        setCardFace(info.cardFaceImage);
        if (masked) {
          targetMaskedText.textContent = masked;
          targetHint.style.display = 'flex';
        }
        cardBox.style.display = 'flex';
        claimBlock.style.display = 'block';
        loadAppConfig().then(showAppBlock);
      }).catch(function(){
        titleEl.textContent = '网络异常';
        descEl.textContent = '请检查网络后刷新重试。';
      });
    }

    function startCountdown(){
      var n = 60;
      codeBtn.disabled = true;
      codeBtn.textContent = n + 's 后重发';
      var timer = setInterval(function(){
        n -= 1;
        if (n <= 0) {
          clearInterval(timer);
          codeBtn.disabled = false;
          codeBtn.textContent = '重新获取';
        } else {
          codeBtn.textContent = n + 's 后重发';
        }
      }, 1000);
    }

    function sendCode(){
      var mobile = (mobileInput.value || '').trim();
      if (!isValidMobile(mobile)) { setTip('请先输入正确的手机号', 'error'); return; }
      setTip('正在发送验证码...');
      fetch('/api/digitalCard/sendClaimVerifyCode', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({token: TOKEN, mobile: mobile})
      }).then(function(r){return r.json()}).then(function(body){
        if (!body || body.code !== 0) { setTip((body && body.message) || '验证码下发失败', 'error'); return; }
        startCountdown();
        var mock = body.data && body.data.mockCode;
        if (mock) {
          codeInput.value = mock;
          setTip('演示阶段：验证码 ' + mock + ' 已自动填入', 'success');
        } else {
          setTip('验证码已发送至 ' + mobile, 'success');
        }
      }).catch(function(){ setTip('网络异常，请稍后重试', 'error'); });
    }

    function showResult(success, title, desc){
      cardBox.style.display = 'none';
      claimBlock.style.display = 'none';
      targetHint.style.display = 'none';
      titleEl.textContent = success ? '领取成功' : '领取失败';
      descEl.textContent = success ? '卡片已发放至你的账户' : '请确认手机号与朋友分享时一致';
      resultTitle.textContent = title;
      resultDesc.textContent = desc;
      resultIconWrap.className = 'icon-wrap' + (success ? '' : ' fail');
      resultIcon.outerHTML = (success ? ICON_CHECK : ICON_X).replace('<svg', '<svg id="resultIcon"');
      successBlock.style.display = 'block';
    }

    function claim(){
      var mobile = (mobileInput.value || '').trim();
      var code = (codeInput.value || '').trim();
      if (!isValidMobile(mobile)) { setTip('请输入有效的手机号', 'error'); return; }
      if (code.length !== 6) { setTip('请输入 6 位验证码', 'error'); return; }
      claimBtn.disabled = true;
      setTip('正在领取...');
      fetch('/api/digitalCard/claimByMobile', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({token: TOKEN, mobile: mobile, verifyCode: code, requestId: 'h5-' + Date.now()})
      }).then(function(r){return r.json()}).then(function(body){
        claimBtn.disabled = false;
        if (!body || body.code !== 0 || !body.data) {
          setTip((body && body.message) || '领取失败，请稍后重试', 'error');
          return;
        }
        if (body.data.appDownload) showAppBlock(body.data.appDownload);
        if (body.data.success) {
          var card = body.data.card;
          var name = card && card.templateName ? card.templateName : '提货卡';
          showResult(true, '已领取「' + name + '」', '打开 App 即可查看你的卡片详情');
          setTip('');
        } else {
          var msg = body.data.failureReason || '此卡片仅限指定接收人领取';
          showResult(false, msg, '建议先注册成为' + (APP_CONFIG && APP_CONFIG.appName ? APP_CONFIG.appName : '九克城') + '会员');
          setTip('');
        }
      }).catch(function(){ claimBtn.disabled = false; setTip('网络异常，请稍后重试', 'error'); });
    }

    codeBtn.addEventListener('click', sendCode);
    claimBtn.addEventListener('click', claim);
    mobileInput.addEventListener('input', function(){
      this.value = this.value.replace(/\D/g, '').slice(0, 11);
    });
    codeInput.addEventListener('input', function(){
      this.value = this.value.replace(/\D/g, '').slice(0, 6);
    });
    validateToken();
  </script>
</body>
</html>`))

// DigitalCardClaimHintHandler 处理 GET /h5/digital-card/claim?token=xxx
// 用于朋友扫描二维码或点击分享链接时的 H5 落地页。
func DigitalCardClaimHintHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = claimHintPage.Execute(w, claimHintPageData{
			Token: strings.TrimSpace(r.URL.Query().Get("token")),
		})
	}
}
