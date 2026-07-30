package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// WebUI serves the embedded management interface.
// @router /ui [GET]
func WebUI(ctx context.Context, c *app.RequestContext) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.SetStatusCode(consts.StatusOK)
	c.WriteString(webUIHTML) //nolint:errcheck
}

const webUIHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>iKuai Tools</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#f0f2f5;color:#222}
.nav{background:#1677ff;color:#fff;padding:0 24px;display:flex;align-items:center;height:52px;gap:24px}
.nav h1{font-size:18px;font-weight:600;flex:1}
.nav a{color:rgba(255,255,255,.85);text-decoration:none;font-size:14px;padding:6px 12px;border-radius:4px;cursor:pointer}
.nav a:hover,.nav a.active{background:rgba(255,255,255,.2);color:#fff}
.container{max-width:1200px;margin:24px auto;padding:0 16px}
.card{background:#fff;border-radius:8px;padding:20px;margin-bottom:20px;box-shadow:0 1px 3px rgba(0,0,0,.08)}
.card h2{font-size:16px;margin-bottom:16px;color:#1677ff}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:16px;margin-bottom:24px}
.stat{background:#fff;border-radius:8px;padding:20px;text-align:center;box-shadow:0 1px 3px rgba(0,0,0,.08)}
.stat .val{font-size:32px;font-weight:700;color:#1677ff}
.stat .label{font-size:13px;color:#888;margin-top:4px}
table{width:100%;border-collapse:collapse;font-size:13px}
th{background:#fafafa;padding:10px 12px;text-align:left;border-bottom:1px solid #eee;font-weight:500;color:#555}
td{padding:10px 12px;border-bottom:1px solid #f5f5f5}
tr:hover td{background:#fafafa}
.btn{display:inline-flex;align-items:center;gap:4px;padding:6px 14px;border:none;border-radius:4px;cursor:pointer;font-size:13px;transition:.15s}
.btn-primary{background:#1677ff;color:#fff}.btn-primary:hover{background:#0958d9}
.btn-danger{background:#ff4d4f;color:#fff}.btn-danger:hover{background:#d9363e}
.btn-default{background:#fff;border:1px solid #d9d9d9}.btn-default:hover{background:#f5f5f5}
.form-row{display:flex;gap:12px;flex-wrap:wrap;margin-bottom:12px;align-items:flex-end}
.form-group{display:flex;flex-direction:column;gap:4px;min-width:180px}
.form-group label{font-size:12px;color:#666;font-weight:500}
input,select,textarea{padding:7px 10px;border:1px solid #d9d9d9;border-radius:4px;font-size:13px;outline:none;width:100%}
input:focus,select:focus,textarea:focus{border-color:#1677ff;box-shadow:0 0 0 2px rgba(22,119,255,.1)}
textarea{min-height:100px;resize:vertical;font-family:monospace}
.badge{display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:500}
.badge-ok{background:#f6ffed;color:#52c41a;border:1px solid #b7eb8f}
.badge-err{background:#fff2f0;color:#ff4d4f;border:1px solid #ffccc7}
.badge-run{background:#e6f4ff;color:#1677ff;border:1px solid #91caff}
.tab-bar{display:flex;gap:2px;margin-bottom:16px;border-bottom:1px solid #eee}
.tab{padding:8px 16px;cursor:pointer;border-radius:4px 4px 0 0;font-size:14px;color:#666}
.tab.active{background:#1677ff;color:#fff}
.tab:hover:not(.active){background:#f0f2f5}
.msg{padding:10px 14px;border-radius:6px;margin-bottom:12px;font-size:13px}
.msg-ok{background:#f6ffed;border:1px solid #b7eb8f;color:#389e0d}
.msg-err{background:#fff2f0;border:1px solid #ffccc7;color:#a8071a}
.loading{color:#999;font-size:13px;padding:20px;text-align:center}
#page-status,#page-firewall,#page-sync{display:none}
#page-status.show,#page-firewall.show,#page-sync.show{display:block}
pre{font-size:12px;background:#f6f8fa;padding:12px;border-radius:4px;overflow-x:auto}
</style>
</head>
<body>
<div class="nav">
  <h1>iKuai Tools</h1>
  <select id="router-select" onchange="onRouterChange()" style="padding:4px 8px;border-radius:4px;border:none;font-size:13px;max-width:180px"></select>
  <a onclick="showPage('status')" id="tab-status">系统状态</a>
  <a onclick="showPage('firewall')" id="tab-firewall">防火墙规则</a>
  <a onclick="showPage('sync')" id="tab-sync">同步任务</a>
  <a onclick="showPage('routers')" id="tab-routers">路由器管理</a>
</div>
<div class="container">

<!-- 系统状态页 -->
<div id="page-status">
  <div id="stats-grid" class="grid"></div>
  <div class="card">
    <h2>接口状态</h2>
    <div id="iface-table"><div class="loading">加载中...</div></div>
  </div>
  <div class="card">
    <h2>在线设备</h2>
    <div id="device-table"><div class="loading">加载中...</div></div>
  </div>
</div>

<!-- 防火墙规则页 -->
<div id="page-firewall">
  <div class="tab-bar">
    <div class="tab active" onclick="showFwTab('custom-isp')">Custom ISP</div>
    <div class="tab" onclick="showFwTab('stream-domain')">域名分流</div>
    <div class="tab" onclick="showFwTab('stream-ipport')">端口分流</div>
    <div class="tab" onclick="showFwTab('ip-group')">IP分组</div>
    <div class="tab" onclick="showFwTab('ipv6-group')">IPv6分组</div>
    <div class="tab" onclick="showFwTab('acl')">ACL</div>
    <div class="tab" onclick="showFwTab('dnat')">端口映射</div>
    <div class="tab" onclick="showFwTab('conn-limit')">连接数限制</div>
  </div>
  <div id="fw-content"><div class="loading">选择规则类型</div></div>
</div>

<!-- 同步任务页 -->
<div id="page-sync">
  <div class="card">
    <h2>任务状态</h2>
    <button class="btn btn-default" onclick="loadSyncStatus()">刷新</button>
    <div id="sync-status-table" style="margin-top:12px"></div>
  </div>
  <div class="card">
    <h2>手动触发 - Custom ISP</h2>
    <div id="msg-cisp"></div>
    <div class="form-row">
      <div class="form-group"><label>规则名称</label><input id="cisp-name" placeholder="my-isp"></div>
      <div class="form-group"><label>GH代理 (可选)</label><input id="cisp-proxy" placeholder="https://ghproxy.com"></div>
    </div>
    <div class="form-group" style="margin-bottom:12px"><label>URL列表 (每行一个)</label><textarea id="cisp-urls" placeholder="https://example.com/ip-list.txt"></textarea></div>
    <button class="btn btn-primary" onclick="triggerCustomISP()">立即同步</button>
  </div>
  <div class="card">
    <h2>手动触发 - 域名分流</h2>
    <div id="msg-sd"></div>
    <div class="form-row">
      <div class="form-group"><label>接口 (逗号分隔)</label><input id="sd-iface" placeholder="wan1,wan2"></div>
      <div class="form-group"><label>备注</label><input id="sd-comment" placeholder="ikuai-tools"></div>
      <div class="form-group"><label>来源地址 (可选)</label><input id="sd-src" placeholder=""></div>
      <div class="form-group"><label>GH代理 (可选)</label><input id="sd-proxy" placeholder="https://ghproxy.com"></div>
    </div>
    <div class="form-group" style="margin-bottom:12px"><label>URL列表 (每行一个)</label><textarea id="sd-urls" placeholder="https://example.com/domain-list.txt"></textarea></div>
    <button class="btn btn-primary" onclick="triggerStreamDomain()">立即同步</button>
  </div>
</div>

<!-- 路由器管理页 -->
<div id="page-routers">
  <div class="card">
    <h2>已管理的路由器</h2>
    <button class="btn btn-default" onclick="loadRoutersPage()">刷新</button>
    <div id="routers-table" style="margin-top:12px"></div>
  </div>
  <div class="card">
    <h2>添加 / 编辑路由器</h2>
    <div id="msg-router"></div>
    <div class="form-row">
      <div class="form-group"><label>名称（唯一标识）</label><input id="rt-name" placeholder="office"></div>
      <div class="form-group"><label>地址</label><input id="rt-url" placeholder="https://192.168.1.1"></div>
      <div class="form-group"><label>Token</label><input id="rt-token" placeholder="个人API令牌"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>备注</label><input id="rt-comment" placeholder=""></div>
    </div>
    <button class="btn btn-primary" onclick="addRouter()">添加</button>
  </div>
</div>

</div>

<script>
let currentRouter = localStorage.getItem('ikuai-router') || '';
let apiKey = localStorage.getItem('ikuai-apikey') || '';
let currentFwTab = '';

function apiBase() { return '/api/v1/ikuai/' + encodeURIComponent(currentRouter); }
function authHeaders(extra={}) {
  const h = {'Content-Type':'application/json', ...extra};
  if (apiKey) h['Authorization'] = 'Bearer ' + apiKey;
  return h;
}

async function api(path, opts={}) {
  const r = await fetch(apiBase()+path, {headers:authHeaders(), ...opts});
  return r.json();
}

async function loadRouters() {
  try {
    const r = await fetch('/api/v1/routers', {headers:authHeaders()});
    const j = await r.json();
    const sel = document.getElementById('router-select');
    sel.innerHTML = '';
    const list = (j.data||[]);
    if (list.length === 0) {
      sel.innerHTML = '<option value="">（请先添加路由器）</option>';
      return;
    }
    list.forEach(rt => {
      const o = document.createElement('option');
      o.value = rt.name; o.textContent = rt.name;
      sel.appendChild(o);
    });
    if (!currentRouter || !list.find(rt=>rt.name===currentRouter)) {
      currentRouter = list[0].name;
      localStorage.setItem('ikuai-router', currentRouter);
    }
    sel.value = currentRouter;
  } catch(e) { console.error('load routers', e); }
}

function onRouterChange() {
  currentRouter = document.getElementById('router-select').value;
  localStorage.setItem('ikuai-router', currentRouter);
  // refresh the active page
  const active = document.querySelector('.nav a.active');
  if (active) active.click();
}

function showPage(name) {
  ['status','firewall','sync','routers'].forEach(p=>{
    const pg = document.getElementById('page-'+p);
    if (pg) pg.classList.remove('show');
    const tb = document.getElementById('tab-'+p);
    if (tb) tb.classList.remove('active');
  });
  document.getElementById('page-'+name).classList.add('show');
  document.getElementById('tab-'+name).classList.add('active');
  if(name==='status') loadStatus();
  if(name==='sync') loadSyncStatus();
  if(name==='routers') loadRoutersPage();
  if(name==='firewall' && !currentFwTab) showFwTab('custom-isp');
}

// boot: ensure API key, then load routers
(async function(){
  if (!apiKey) {
    apiKey = prompt('请输入 API Key（在 auth.api_key 配置中设置，留空则无鉴权）') || '';
    if (apiKey) localStorage.setItem('ikuai-apikey', apiKey);
  }
  await loadRouters();
  showPage('status');
})();

// ── Status ──────────────────────────────────────────────────────────────────
async function loadStatus() {
  try {
    const s = await api('/system/status');
    if(!s.data) return;
    const d = s.data;
    const grid = document.getElementById('stats-grid');
    const uptime = d.Uptime ? Math.floor(d.Uptime/86400)+'d '+Math.floor((d.Uptime%86400)/3600)+'h' : '--';
    const cpu = Array.isArray(d.CPU) ? d.CPU.map((c,i)=>'核'+i+':'+c).join(' ') : '--';
    const mem = d.Memory ? Math.round((d.Memory.Total-d.Memory.Available)/d.Memory.Total*100)+'%' : '--';
    grid.innerHTML = stat('运行时间', uptime)+stat('CPU', cpu)+stat('内存使用', mem)+stat('版本', d.VerInfo&&d.VerInfo.VerString||'--');
  } catch(e){ console.error(e); }

  try {
    const r = await api('/system/interfaces');
    document.getElementById('iface-table').innerHTML = ifaceTable(r.data);
  } catch(e){}

  try {
    const r = await api('/system/devices');
    document.getElementById('device-table').innerHTML = deviceTable(r.data);
  } catch(e){}
}

function stat(label, val) {
  return '<div class="stat"><div class="val">'+escH(String(val||'--'))+'</div><div class="label">'+escH(label)+'</div></div>';
}

function ifaceTable(data) {
  const streams = (data&&data.IFaceStream)||[];
  if(!streams.length) return '<div class="loading">无数据</div>';
  return '<table><tr><th>接口</th><th>IP</th><th>上传速度</th><th>下载速度</th><th>连接数</th></tr>'+
    streams.map(r=>'<tr><td>'+escH(r.Interface||r.Comment||'')+'</td><td>'+escH(r.IPAddr||'')+'</td>'+
      '<td>'+fmtSpeed(r.Upload)+'</td><td>'+fmtSpeed(r.Download)+'</td><td>'+escH(String(r.ConnectNum||0))+'</td></tr>').join('')+'</table>';
}

function deviceTable(data) {
  const items = data||[];
  if(!items.length) return '<div class="loading">无设备</div>';
  return '<table><tr><th>主机名</th><th>IP</th><th>MAC</th><th>上传</th><th>下载</th></tr>'+
    items.map(r=>'<tr><td>'+escH(r.Hostname||r.Comment||'')+'</td><td>'+escH(r.IPAddr||'')+'</td>'+
      '<td>'+escH(r.Mac||'')+'</td><td>'+fmtSpeed(r.Upload)+'</td><td>'+fmtSpeed(r.Download)+'</td></tr>').join('')+'</table>';
}

// ── Firewall ─────────────────────────────────────────────────────────────────
function showFwTab(name) {
  currentFwTab = name;
  document.querySelectorAll('.tab-bar .tab').forEach((t,i)=>{
    const tabs=['custom-isp','stream-domain','stream-ipport','ip-group','ipv6-group','acl','dnat','conn-limit'];
    t.classList.toggle('active', tabs[i]===name);
  });
  loadFwTab(name);
}

async function loadFwTab(name) {
  document.getElementById('fw-content').innerHTML='<div class="loading">加载中...</div>';
  try {
    const r = await api('/firewall/'+name);
    const items = r.data||[];
    document.getElementById('fw-content').innerHTML = renderFwTable(name, items);
  } catch(e) {
    document.getElementById('fw-content').innerHTML='<div class="msg msg-err">加载失败: '+escH(e.message)+'</div>';
  }
}

function renderFwTable(name, items) {
  if(!items.length) return '<div class="card"><div class="loading">暂无数据</div></div>';
  const configs = {
    'custom-isp':    {cols:['ID','名称','备注','IP数量'],       row:r=>[r.id,r.name,r.comment,(r.ipgroup||'').split(',').length]},
    'stream-domain': {cols:['ID','接口','域名数','来源地址','备注'],row:r=>[r.id,r.interface,(r.domain||'').split(',').length,r.src_addr,r.comment]},
    'stream-ipport': {cols:['ID','接口','协议','源IP','目标IP','端口','备注'],row:r=>[r.id,r.interface,r.protocol,r.src_addr,r.dst_addr,r.dst_port||r.src_port,r.comment]},
    'ip-group':      {cols:['ID','名称','IP数量','备注'],       row:r=>[r.id,r.group_name,(r.group_value||'').split(',').length,r.comment]},
    'ipv6-group':    {cols:['ID','名称','IP数量','备注'],       row:r=>[r.id,r.group_name,(r.group_value||'').split(',').length,r.comment]},
    'acl':           {cols:['ID','来源','目标','协议','动作','备注'],row:r=>[r.id,r.src_addr,r.dst_addr,r.protocol,r.action,r.comment]},
    'dnat':          {cols:['ID','接口','协议','外网端口','内网地址','内网端口','备注'],row:r=>[r.id,r.interface,r.protocol,r.wan_port,r.lan_addr,r.lan_port,r.comment]},
    'conn-limit':    {cols:['ID','来源','连接数','备注'],        row:r=>[r.id,r.src_addr,r.conn_num,r.comment]},
  };
  const cfg = configs[name]||{cols:['数据'],row:r=>[JSON.stringify(r)]};
  return '<div class="card"><h2>共 '+items.length+' 条规则 <button class="btn btn-default" style="float:right" onclick="loadFwTab(\''+name+'\')">刷新</button></h2>'+
    '<table><tr>'+cfg.cols.map(h=>'<th>'+escH(h)+'</th>').join('')+'<th>操作</th></tr>'+
    items.map(r=>{
      const vals = cfg.row(r);
      const id = r.id;
      return '<tr>'+vals.map(v=>'<td>'+escH(String(v==null?'':v))+'</td>').join('')+
        '<td><button class="btn btn-danger" onclick="deleteRule(\''+name+'\','+id+')">删除</button></td></tr>';
    }).join('')+'</table></div>';
}

async function deleteRule(tab, id) {
  if(!confirm('确定删除 ID='+id+' 的规则?')) return;
  const r = await api('/firewall/'+tab+'/'+id, {method:'DELETE'});
  if(r.code===0) loadFwTab(tab);
  else alert('删除失败: '+(r.message||''));
}

// ── Sync ─────────────────────────────────────────────────────────────────────
async function loadSyncStatus() {
  try {
    const r = await api('/sync/status');
    const items = r.data||[];
    if(!items.length) {
      document.getElementById('sync-status-table').innerHTML='<div class="loading">暂无任务记录</div>';
      return;
    }
    document.getElementById('sync-status-table').innerHTML=
      '<table><tr><th>任务名</th><th>状态</th><th>上次执行</th><th>执行次数</th><th>最后错误</th></tr>'+
      items.map(s=>'<tr><td>'+escH(s.name)+'</td>'+
        '<td>'+badge(s)+'</td>'+
        '<td>'+(s.last_run_at?new Date(s.last_run_at).toLocaleString():'从未')+'</td>'+
        '<td>'+s.run_count+'</td>'+
        '<td><span title="'+escH(s.last_error||'')+'">'+escH((s.last_error||'').slice(0,60))+'</span></td></tr>').join('')+'</table>';
  } catch(e) {
    document.getElementById('sync-status-table').innerHTML='<div class="msg msg-err">加载失败</div>';
  }
}

function badge(s) {
  if(s.running) return '<span class="badge badge-run">运行中</span>';
  if(s.run_count===0) return '<span class="badge">待运行</span>';
  return s.success?'<span class="badge badge-ok">成功</span>':'<span class="badge badge-err">失败</span>';
}

async function triggerCustomISP() {
  const name=document.getElementById('cisp-name').value.trim();
  const urls=document.getElementById('cisp-urls').value.trim().split('\n').filter(Boolean);
  const proxy=document.getElementById('cisp-proxy').value.trim();
  if(!name||!urls.length){showMsg('msg-cisp','名称和URL不能为空','err');return;}
  showMsg('msg-cisp','同步中...','ok');
  const r=await api('/sync/custom-isp',{method:'POST',body:JSON.stringify({name,url:urls,gh_proxy:proxy})});
  showMsg('msg-cisp',r.code===0?'已触发同步，请稍后查看任务状态':(r.message||'失败'),r.code===0?'ok':'err');
}

async function triggerStreamDomain() {
  const ifaces=document.getElementById('sd-iface').value.trim().split(',').map(s=>s.trim()).filter(Boolean);
  const urls=document.getElementById('sd-urls').value.trim().split('\n').filter(Boolean);
  const comment=document.getElementById('sd-comment').value.trim()||'ikuai-tools';
  const src=document.getElementById('sd-src').value.trim();
  const proxy=document.getElementById('sd-proxy').value.trim();
  if(!ifaces.length||!urls.length){showMsg('msg-sd','接口和URL不能为空','err');return;}
  showMsg('msg-sd','同步中...','ok');
  const r=await api('/sync/stream-domain',{method:'POST',body:JSON.stringify({interface:ifaces,url:urls,comment,src_addr:src,gh_proxy:proxy})});
  showMsg('msg-sd',r.code===0?'已触发同步，请稍后查看任务状态':(r.message||'失败'),r.code===0?'ok':'err');
}

function showMsg(id, text, type) {
  const el=document.getElementById(id);
  el.innerHTML='<div class="msg msg-'+type+'">'+escH(text)+'</div>';
}

// ── Routers management ───────────────────────────────────────────────────────
async function loadRoutersPage() {
  try {
    const r = await fetch('/api/v1/routers', {headers:authHeaders()});
    const j = await r.json();
    const list = j.data || [];
    const tbl = document.getElementById('routers-table');
    if (!list.length) { tbl.innerHTML = '<div class="loading">暂无路由器</div>'; return; }
    tbl.innerHTML = '<table><thead><tr><th>名称</th><th>地址</th><th>Token</th><th>状态</th><th>操作</th></tr></thead><tbody>' +
      list.map(rt => '<tr><td>'+escH(rt.name)+'</td><td>'+escH(rt.base_url)+'</td><td>'+(rt.has_token?'已设置':'未设置')+'</td><td>'+(rt.status===1?'<span class="badge badge-ok">启用</span>':'<span class="badge badge-err">停用</span>')+'</td>'+
      '<td><button class="btn btn-danger" onclick="delRouter(\''+escH(rt.name)+'\')">删除</button></td></tr>').join('') +
      '</tbody></table>';
    // also refresh the top selector
    loadRouters();
  } catch(e) { console.error('loadRoutersPage', e); }
}

async function addRouter() {
  const body = {
    name: document.getElementById('rt-name').value.trim(),
    base_url: document.getElementById('rt-url').value.trim(),
    token: document.getElementById('rt-token').value.trim(),
    comment: document.getElementById('rt-comment').value.trim(),
  };
  if (!body.name || !body.base_url || !body.token) {
    document.getElementById('msg-router').innerHTML = '<div class="msg msg-err">名称、地址、Token 均必填</div>';
    return;
  }
  try {
    const r = await fetch('/api/v1/routers', {method:'POST', headers:authHeaders(), body:JSON.stringify(body)});
    const j = await r.json();
    if (j.code === 0) {
      document.getElementById('msg-router').innerHTML = '<div class="msg msg-ok">已添加并连接</div>';
      ['rt-name','rt-url','rt-token','rt-comment'].forEach(id=>document.getElementById(id).value='');
      loadRoutersPage();
    } else {
      document.getElementById('msg-router').innerHTML = '<div class="msg msg-err">'+escH(j.message||'添加失败')+'</div>';
    }
  } catch(e) {
    document.getElementById('msg-router').innerHTML = '<div class="msg msg-err">'+escH(e.message)+'</div>';
  }
}

async function delRouter(name) {
  if (!confirm('确认删除路由器 '+name+'？')) return;
  await fetch('/api/v1/routers/'+encodeURIComponent(name), {method:'DELETE', headers:authHeaders()});
  loadRoutersPage();
}

// ── Helpers ──────────────────────────────────────────────────────────────────
function escH(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
function fmtSpeed(v) { if(!v)return '0 B/s'; if(v>1048576)return (v/1048576).toFixed(1)+' MB/s'; if(v>1024)return (v/1024).toFixed(1)+' KB/s'; return v+' B/s'; }
</script>
</body>
</html>`
