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

// H5 领取页：朋友扫码后在浏览器（含微信内置浏览器）打开的落地页。
// 页面职责：
//  1. 调用 /api/digitalCard/validateClaimToken 预览卡片信息
//  2. 提示登录/注册（未注册用户跳到 H5 注册页 /h5/digitalCard/register）
//  3. 登录后调用 /api/digitalCard/claim 完成领取
//
// 安全说明：此页面不使用服务端注入 token（仅通过 query 回填到 JS），所有业务校验在后端完成。
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
    .card{display:flex;gap:14px;padding:14px;background:#fafaf7;border-radius:12px;margin:16px 0}
    .card img{width:96px;height:128px;border-radius:8px;object-fit:cover;background:#eee}
    .card .meta{flex:1;display:flex;flex-direction:column;justify-content:center}
    .card .meta h2{margin:0 0 6px;font-size:17px}
    .card .meta span{color:#7d8592;font-size:13px;line-height:1.5}
    .btn{width:100%;box-sizing:border-box;display:block;text-align:center;border:0;border-radius:12px;padding:14px 12px;font-size:16px;font-weight:700;cursor:pointer;text-decoration:none;margin-top:10px}
    .btn-primary{background:#161616;color:#fff}
    .btn-outline{background:#fff;color:#161616;border:1px solid #d8ddd4}
    .tip{margin-top:12px;font-size:13px;color:#7d8592;min-height:20px}
    .label{display:block;margin:12px 0 6px;font-size:13px;color:#525866}
    input{width:100%;box-sizing:border-box;border:1px solid #d8ddd4;border-radius:12px;padding:12px;font-size:16px;background:#fff}
    .error{color:#b42318}
  </style>
</head>
<body>
  <main>
    <section>
      <h1 id="title">正在校验分享链接...</h1>
      <p id="desc">请稍候，正在读取提货卡信息。</p>
      <div class="card" id="cardBox" style="display:none">
        <img id="cardImg" alt="卡面">
        <div class="meta">
          <h2 id="cardName">-</h2>
          <span id="cardSender">-</span>
          <span id="cardExpire">-</span>
        </div>
      </div>
      <div id="loginBlock" style="display:none">
        <label class="label">手机号</label>
        <input id="mobile" maxlength="11" inputmode="numeric" placeholder="请输入已注册的手机号">
        <label class="label">密码</label>
        <input id="password" type="password" minlength="6" placeholder="请输入密码">
        <button class="btn btn-primary" id="loginBtn">登录并领取</button>
        <a class="btn btn-outline" id="registerLink" href="#">没有账号？去注册</a>
      </div>
      <div id="successBlock" style="display:none">
        <p>领取成功！卡片已转入你的账户。</p>
        <a class="btn btn-outline" href="javascript:void(0)" onclick="window.close()">关闭</a>
      </div>
      <p class="tip" id="tip"></p>
    </section>
  </main>
  <script>
    var TOKEN = {{.Token}};
    var titleEl = document.getElementById('title');
    var descEl = document.getElementById('desc');
    var cardBox = document.getElementById('cardBox');
    var cardImg = document.getElementById('cardImg');
    var cardName = document.getElementById('cardName');
    var cardSender = document.getElementById('cardSender');
    var cardExpire = document.getElementById('cardExpire');
    var loginBlock = document.getElementById('loginBlock');
    var successBlock = document.getElementById('successBlock');
    var tip = document.getElementById('tip');
    var loginBtn = document.getElementById('loginBtn');
    var registerLink = document.getElementById('registerLink');

    function setTip(msg, isError) {
      tip.textContent = msg || '';
      tip.className = isError ? 'tip error' : 'tip';
    }

    function validateToken() {
      if (!TOKEN) {
        titleEl.textContent = '链接无效';
        descEl.textContent = '请联系好友重新分享。';
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
          return;
        }
        var info = body.data.token || {};
        titleEl.textContent = '朋友送你一张提货卡';
        descEl.textContent = '请登录领取，没有账号可先完成注册。';
        cardName.textContent = info.templateName || '提货卡';
        cardSender.textContent = '分享人：' + (info.senderName || '朋友');
        cardExpire.textContent = '有效期至：' + (info.expireAt || '-');
        if (info.cardFaceImage) { cardImg.src = info.cardFaceImage; }
        cardBox.style.display = 'flex';
        loginBlock.style.display = 'block';
        registerLink.href = '/h5/digitalCard/register?redirect=' + encodeURIComponent(location.href);
      }).catch(function(){
        titleEl.textContent = '网络异常';
        descEl.textContent = '请检查网络后刷新重试。';
      });
    }

    function login() {
      var mobile = (document.getElementById('mobile').value || '').trim();
      var password = document.getElementById('password').value || '';
      if (!/^1[3-9]\d{9}$/.test(mobile)) { setTip('请输入有效的手机号', true); return; }
      if (password.length < 6) { setTip('请输入不少于 6 位的密码', true); return; }
      setTip('正在登录...');
      fetch('/api/member/login', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({mobile: mobile, password: password})
      }).then(function(r){return r.json()}).then(function(body){
        if (!body || body.code !== 0 || !body.data) {
          setTip((body && body.message) || '登录失败，请检查手机号或密码', true);
          return;
        }
        var token = body.data.token || body.data.accessToken;
        if (!token) { setTip('登录信息异常', true); return; }
        claim(token);
      }).catch(function(){ setTip('网络异常，请稍后重试', true); });
    }

    function claim(authToken) {
      setTip('正在领取...');
      fetch('/api/digitalCard/claim', {
        method: 'POST',
        headers: {'Content-Type': 'application/json', 'Authorization': 'Bearer ' + authToken},
        body: JSON.stringify({token: TOKEN, requestId: 'h5-' + Date.now()})
      }).then(function(r){return r.json()}).then(function(body){
        if (!body || body.code !== 0) {
          setTip((body && body.message) || '领取失败，请稍后重试', true);
          return;
        }
        cardBox.style.display = 'none';
        loginBlock.style.display = 'none';
        successBlock.style.display = 'block';
        titleEl.textContent = '领取成功';
        descEl.textContent = '卡片已转入你的账户，打开 App 即可查看。';
        setTip('');
      }).catch(function(){ setTip('网络异常，请稍后重试', true); });
    }

    loginBtn.addEventListener('click', login);
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
