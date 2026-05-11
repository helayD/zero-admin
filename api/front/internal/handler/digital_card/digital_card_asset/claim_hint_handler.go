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

// H5 领取页（Story 10.7 闭环修复 2026-05-12 重写）：
//
// 流程：
//  1. 加载时调 /api/digitalCard/validateClaimToken 预览卡片信息（含接收人手机号掩码）
//  2. 一步式：用户输入「手机号 + 验证码（mock=123456）」 → /api/digitalCard/claimByMobile
//     - 命中接收人手机号 → 自动注册或登录 → 转赠成功 → 显示「领取成功 + 下载 App」
//     - 不命中           → 显示「领取失败 + 下载 App 注册」
//  3. App 下载链接来自 /api/config/appDownload（后台 sys_system_config 配置）
//
// 安全说明：所有业务校验仍在后端完成；HTML 仅是 UI 包装。
var claimHintPage = template.Must(template.New("digital-card-claim").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">
  <title>领取提货卡</title>
  <style>
    body{margin:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#f6f7f4;color:#161616}
    main{min-height:100vh;display:flex;align-items:center;justify-content:center;padding:24px;box-sizing:border-box}
    section{width:100%;max-width:420px;background:#fff;border:1px solid #e4e7e0;border-radius:18px;padding:24px;box-sizing:border-box;box-shadow:0 18px 40px rgba(15,23,42,.08)}
    h1{margin:0 0 12px;font-size:22px;line-height:1.3}
    p{margin:0 0 12px;color:#525866;line-height:1.65;font-size:14px}
    .card{display:flex;gap:14px;padding:14px;background:#fafaf7;border-radius:12px;margin:16px 0;align-items:center}
    .card .face{width:96px;height:128px;border-radius:8px;background:linear-gradient(135deg,#f5e6c8 0%,#e6c9a0 100%);display:flex;align-items:center;justify-content:center;color:#8b6f3a;font-size:13px;flex-shrink:0;overflow:hidden;text-align:center;padding:6px;box-sizing:border-box}
    .card .face img{width:100%;height:100%;border-radius:8px;object-fit:cover}
    .card .meta{flex:1;display:flex;flex-direction:column;justify-content:center;gap:6px}
    .card .meta h2{margin:0;font-size:17px}
    .card .meta span{color:#7d8592;font-size:13px;line-height:1.5}
    .target-hint{padding:10px 12px;background:#fff8e6;border:1px solid #ffd591;color:#a46500;border-radius:10px;font-size:13px;margin-bottom:12px;line-height:1.5}
    .btn{width:100%;box-sizing:border-box;display:block;text-align:center;border:0;border-radius:12px;padding:14px 12px;font-size:16px;font-weight:700;cursor:pointer;text-decoration:none;margin-top:10px}
    .btn-primary{background:#161616;color:#fff}
    .btn-primary:disabled{background:#9ca3a0;cursor:not-allowed}
    .btn-outline{background:#fff;color:#161616;border:1px solid #d8ddd4}
    .btn-warn{background:#fff5f5;color:#b42318;border:1px solid #ffccc7}
    .tip{margin-top:12px;font-size:13px;color:#7d8592;min-height:20px;line-height:1.5}
    .label{display:block;margin:12px 0 6px;font-size:13px;color:#525866}
    .row{display:flex;gap:8px;align-items:stretch}
    .row input{flex:1}
    .row .code-btn{flex-shrink:0;border:1px solid #d8ddd4;background:#fff;border-radius:12px;padding:0 14px;font-size:14px;color:#161616;cursor:pointer;white-space:nowrap}
    .row .code-btn:disabled{color:#9ca3a0;cursor:not-allowed;background:#f6f7f4}
    input{width:100%;box-sizing:border-box;border:1px solid #d8ddd4;border-radius:12px;padding:12px;font-size:16px;background:#fff}
    .error{color:#b42318}
    .success{color:#1a7f37}
    .app-block{margin-top:18px;padding:14px;background:#f0f5ff;border:1px solid #c8d8ff;border-radius:12px;text-align:center}
    .app-block .app-name{font-size:15px;font-weight:700;margin-bottom:4px}
    .app-block .app-tag{font-size:13px;color:#525866;margin-bottom:10px}
    .footer{margin-top:18px;font-size:11px;color:#9ca3a0;text-align:center}
  </style>
</head>
<body>
  <main>
    <section>
      <h1 id="title">正在校验分享链接...</h1>
      <p id="desc">请稍候，正在读取提货卡信息。</p>

      <div class="card" id="cardBox" style="display:none">
        <div class="face" id="cardFaceWrap"><span id="cardFaceFallback">卡面</span></div>
        <div class="meta">
          <h2 id="cardName">-</h2>
          <span id="cardExpire">-</span>
          <span id="cardSender">-</span>
        </div>
      </div>

      <div id="targetHint" class="target-hint" style="display:none"></div>

      <div id="claimBlock" style="display:none">
        <label class="label">手机号</label>
        <input id="mobile" maxlength="11" inputmode="numeric" placeholder="请输入接收卡片的手机号">
        <label class="label">验证码</label>
        <div class="row">
          <input id="code" maxlength="6" inputmode="numeric" placeholder="请输入 6 位验证码">
          <button class="code-btn" id="codeBtn" type="button">获取验证码</button>
        </div>
        <button class="btn btn-primary" id="claimBtn">领取卡片</button>
      </div>

      <div id="successBlock" style="display:none">
        <p class="success" id="successMsg">领取成功！卡片已转入你的账户。</p>
      </div>

      <div id="appBlock" class="app-block" style="display:none">
        <div class="app-name" id="appName">下载 App</div>
        <div class="app-tag" id="appTag">下载 App，查看你的提货卡</div>
        <a class="btn btn-outline" id="androidLink" href="#" target="_blank" rel="noopener">Android 下载</a>
        <a class="btn btn-outline" id="iosLink" href="#" target="_blank" rel="noopener">iOS 下载</a>
      </div>

      <p class="tip" id="tip"></p>
      <p class="footer">本页面由九克城提供 · 仅限指定接收人领取</p>
    </section>
  </main>
  <script>
    var TOKEN = {{.Token}};
    var titleEl = document.getElementById('title');
    var descEl = document.getElementById('desc');
    var cardBox = document.getElementById('cardBox');
    var cardFaceWrap = document.getElementById('cardFaceWrap');
    var cardFaceFallback = document.getElementById('cardFaceFallback');
    var cardName = document.getElementById('cardName');
    var cardSender = document.getElementById('cardSender');
    var cardExpire = document.getElementById('cardExpire');
    var targetHint = document.getElementById('targetHint');
    var claimBlock = document.getElementById('claimBlock');
    var successBlock = document.getElementById('successBlock');
    var successMsg = document.getElementById('successMsg');
    var appBlock = document.getElementById('appBlock');
    var appName = document.getElementById('appName');
    var appTag = document.getElementById('appTag');
    var androidLink = document.getElementById('androidLink');
    var iosLink = document.getElementById('iosLink');
    var tip = document.getElementById('tip');
    var codeBtn = document.getElementById('codeBtn');
    var claimBtn = document.getElementById('claimBtn');
    var mobileInput = document.getElementById('mobile');
    var codeInput = document.getElementById('code');

    var APP_CONFIG = null;

    function setTip(msg, isError) {
      tip.textContent = msg || '';
      tip.className = isError ? 'tip error' : 'tip';
    }
    function isValidMobile(m){ return /^1[3-9]\d{9}$/.test(m); }
    function setCardFace(url){
      if (url) {
        var img = document.createElement('img');
        img.src = url;
        img.alt = '卡面';
        img.onerror = function(){ cardFaceWrap.innerHTML = '<span>卡面缺失</span>'; };
        cardFaceWrap.innerHTML = '';
        cardFaceWrap.appendChild(img);
      } else {
        cardFaceWrap.innerHTML = '<span>卡面缺失</span>';
      }
    }
    function showAppBlock(cfg){
      if (!cfg) return;
      APP_CONFIG = cfg;
      if (cfg.appName) appName.textContent = cfg.appName;
      if (cfg.tagline) appTag.textContent = cfg.tagline;
      if (cfg.androidUrl) androidLink.href = cfg.androidUrl; else androidLink.style.display = 'none';
      if (cfg.iosUrl) iosLink.href = cfg.iosUrl; else iosLink.style.display = 'none';
      appBlock.style.display = 'block';
    }

    function loadAppConfig(){
      return fetch('/api/config/appDownload').then(function(r){return r.json()}).then(function(b){
        if (b && b.code === 0 && b.data) { return b.data; }
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
        descEl.textContent = '请输入手机号 + 验证码完成领取。';
        cardName.textContent = info.templateName || '提货卡';
        cardExpire.textContent = '有效期至：' + (info.expireAt || '-');
        cardSender.textContent = '分享人：朋友';
        setCardFace(info.cardFaceImage);
        if (masked) {
          targetHint.style.display = 'block';
          targetHint.textContent = '此卡片仅限手机号 ' + masked + ' 领取';
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
          codeBtn.textContent = '获取验证码';
        } else {
          codeBtn.textContent = n + 's 后重发';
        }
      }, 1000);
    }

    function sendCode(){
      var mobile = (mobileInput.value || '').trim();
      if (!isValidMobile(mobile)) { setTip('请先输入正确的手机号', true); return; }
      setTip('正在发送验证码...');
      fetch('/api/digitalCard/sendClaimVerifyCode', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({token: TOKEN, mobile: mobile})
      }).then(function(r){return r.json()}).then(function(body){
        if (!body || body.code !== 0) { setTip((body && body.message) || '验证码下发失败', true); return; }
        startCountdown();
        var mock = body.data && body.data.mockCode;
        if (mock) {
          // mock 阶段：直接把验证码填进输入框，方便联调
          codeInput.value = mock;
          setTip('演示阶段：验证码 ' + mock + ' 已自动填入', false);
        } else {
          setTip('验证码已发送至 ' + mobile, false);
        }
      }).catch(function(){ setTip('网络异常，请稍后重试', true); });
    }

    function claim(){
      var mobile = (mobileInput.value || '').trim();
      var code = (codeInput.value || '').trim();
      if (!isValidMobile(mobile)) { setTip('请输入有效的手机号', true); return; }
      if (code.length !== 6) { setTip('请输入 6 位验证码', true); return; }
      claimBtn.disabled = true;
      setTip('正在领取...');
      fetch('/api/digitalCard/claimByMobile', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({token: TOKEN, mobile: mobile, verifyCode: code, requestId: 'h5-' + Date.now()})
      }).then(function(r){return r.json()}).then(function(body){
        claimBtn.disabled = false;
        if (!body || body.code !== 0 || !body.data) {
          setTip((body && body.message) || '领取失败，请稍后重试', true);
          return;
        }
        if (body.data.appDownload) showAppBlock(body.data.appDownload);
        if (body.data.success) {
          claimBlock.style.display = 'none';
          targetHint.style.display = 'none';
          successBlock.style.display = 'block';
          titleEl.textContent = '领取成功';
          descEl.textContent = '卡片已发放至你的账户，打开 App 查看更多详情。';
          var card = body.data.card;
          if (card && card.cardFaceImage) setCardFace(card.cardFaceImage);
          successMsg.textContent = '已领取：' + (card && card.templateName ? card.templateName : '提货卡');
          setTip('');
        } else {
          // 领取失败：手机号不命中 / 已过期 / 已被领等
          claimBlock.style.display = 'none';
          targetHint.style.display = 'none';
          cardBox.style.display = 'none';
          titleEl.textContent = '领取失败';
          descEl.textContent = body.data.failureReason || '此卡片仅限指定接收人领取';
          successMsg.className = 'error';
          successBlock.style.display = 'block';
          successMsg.textContent = '建议先注册成为九克城会员';
          setTip('');
        }
      }).catch(function(){ claimBtn.disabled = false; setTip('网络异常，请稍后重试', true); });
    }

    codeBtn.addEventListener('click', sendCode);
    claimBtn.addEventListener('click', claim);
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
