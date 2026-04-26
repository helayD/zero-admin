package digital_card_asset

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/svc"
)

type transferRegisterPageData struct {
	Mobile          string
	AssetInstanceID string
}

var transferRegisterPage = template.Must(template.New("digital-card-register").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">
  <title>注册接收卡片</title>
  <style>
    body{margin:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#f6f7f4;color:#161616}
    main{min-height:100vh;display:flex;align-items:center;justify-content:center;padding:24px;box-sizing:border-box}
    section{width:100%;max-width:420px;background:#fff;border:1px solid #e4e7e0;border-radius:18px;padding:24px;box-sizing:border-box;box-shadow:0 18px 40px rgba(15,23,42,.08)}
    h1{margin:0 0 10px;font-size:24px;line-height:1.25}
    p{margin:0 0 18px;color:#525866;line-height:1.65}
    label{display:block;margin:12px 0 6px;font-size:13px;color:#525866}
    input{width:100%;box-sizing:border-box;border:1px solid #d8ddd4;border-radius:12px;padding:13px 12px;font-size:16px;background:#fff}
    button{width:100%;margin-top:18px;border:0;border-radius:12px;background:#161616;color:white;padding:14px 12px;font-size:16px;font-weight:700}
    .tip{margin-top:14px;font-size:13px;color:#7d8592}
  </style>
</head>
<body>
  <main>
    <section>
      <h1>注册后接收卡片</h1>
      <p>好友向你转赠了一张数字卡片。请先完成注册，注册后可登录查看并接收。</p>
      <form id="registerForm">
        <label>手机号</label>
        <input name="mobile" value="{{.Mobile}}" maxlength="11" inputmode="numeric" required>
        <label>昵称</label>
        <input name="nickname" maxlength="24" placeholder="请输入昵称" required>
        <label>密码</label>
        <input name="password" type="password" minlength="6" placeholder="至少 6 位" required>
        <label>确认密码</label>
        <input name="confirmPassword" type="password" minlength="6" placeholder="再次输入密码" required>
        <button type="submit">注册并接收</button>
      </form>
      <p class="tip" id="registerTip">注册完成后，请回到转赠页面继续接收。</p>
    </section>
  </main>
  <script>
    const form = document.getElementById('registerForm');
    const tip = document.getElementById('registerTip');
    form.addEventListener('submit', async (event) => {
      event.preventDefault();
      const data = Object.fromEntries(new FormData(form).entries());
      data.source = 1;
      if (data.password !== data.confirmPassword) {
        tip.textContent = '两次密码不一致';
        return;
      }
      tip.textContent = '正在注册...';
      try {
        const response = await fetch('/api/member/register', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(data)
        });
        const body = await response.json();
        tip.textContent = body && body.message ? body.message : '注册完成，请回到转赠页面继续接收';
      } catch (err) {
        tip.textContent = '注册失败，请稍后重试';
      }
    });
  </script>
</body>
</html>`))

func DigitalCardRegisterHintHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = transferRegisterPage.Execute(w, transferRegisterPageData{
			Mobile:          strings.TrimSpace(r.URL.Query().Get("mobile")),
			AssetInstanceID: strings.TrimSpace(r.URL.Query().Get("assetInstanceId")),
		})
	}
}
