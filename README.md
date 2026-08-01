# ikuai-tools-service

> iKuai 路由器统一管理服务：多路由器管理、防火墙/分流规则编排、定时同步、Prometheus 监控、Web 管理界面。基于 [ikuai-api](../ikuai-api) v4-only SDK（仅支持 iKuai v4 + 个人 API 令牌）。

## 特性

- **多路由器管理**: 一套服务管理多台 iKuai，路由器配置持久化到 DB，通过 API 增删改查热生效（无需重启）
- **RESTful 多实例 API**: 所有资源按 `/api/v1/ikuai/:router_id/...` 寻址，每台路由器资源独立
- **鉴权与审计**: API Key 鉴权（`Authorization: Bearer` / `X-API-Key`），所有写操作记录到审计表
- **分流自动化**: custom_isp / stream_domain 定时同步，按路由器隔离执行，执行历史持久化（变更/失败/耗时）
- **可观测性**: 结构化日志（zap + 请求 ID 链路）、内嵌 Prometheus exporter（带 router 标签）、真实健康探针
- **Web 管理界面**: 内嵌 `/ui`，路由器切换、规则查看、手动触发同步、路由器增删
- **全栈技术**: CloudWeGo Hertz + GORM（SQLite/MySQL/PostgreSQL）+ Redis + gocron + zap

## 快速开始

```bash
# 1. 准备配置
cp configs/config.yaml.example configs/config.yaml
# 编辑 configs/config.yaml：设置 auth.api_key，ikuai.token（或留空后续通过 API 添加路由器）

# 2. 运行（SQLite 免外部依赖，表会自动迁移）
make run
# 或：CONFIG_PATH=configs/config.yaml go run ./cmd/server

# 3. 访问
#   Web 界面:  http://<host>:9997/ui
#   健康:      http://<host>:9997/ready
#   指标:      http://<host>:9100/metrics
```

### 添加一台路由器

```bash
# 通过 API 添加路由器（设了 auth.api_key 时带 Bearer）
curl -X POST http://localhost:9997/api/v1/routers \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"name":"office","base_url":"https://192.168.1.1","token":"<router-api-token>"}'

# 之后所有操作带上 router_id
curl -H "Authorization: Bearer <your-api-key>" \
  http://localhost:9997/api/v1/ikuai/office/system/status
```

> 个人 API 令牌在路由器管理界面：**系统设置 → 个人令牌**。

## 主要 API

> 以下路径均省略 `/api/v1/ikuai/:router_id` 前缀。除标注只读外，均为完整 CRUD。

| 分组 | 路径 | 说明 |
|------|------|------|
| **路由器管理** | `/api/v1/routers` | CRUD（不区分 router_id） |
| **系统监控** | `monitoring/{interfaces-traffic,downstream,switch,network,app-traffic-summary,clients-traffic-summary}` | 流量/拓扑（只读） |
| **系统状态** | `system/{status,interfaces,devices}` | CPU/内存/接口/设备（只读） |
| **系统运维** | `system/{backup,upgrade,reboot,remote-access,disks,web-admin}` | 备份恢复/固件升级/定时重启/远程访问/磁盘/账号 |
| **日志** | `log/{notice,system,auth,dhcp,pppoe,arp,ddns,web-activity,wireless}` | 9 类设备日志（只读） |
| **防火墙** | `firewall/{acl,dnat,ip-group,ipv6-group,mac-objects,custom-isp,stream-domain,stream-ipport,conn-limit,mac-rules}` | 规则增删改查 |
| **对象组** | `objects/{mac,domain,port,protocol,time}` | 对象组 CRUD（ACL/分流规则引用底座） |
| **网络** | `network/{wan,lan,dhcp/*,dns/static,route/static,dmz,nat,qos/ip,qos/mac,vlan}` | 网络配置 |
| **VPN** | `vpn/{pptp,l2tp}` | VPN 客户端 |
| **多WAN/分流** | `routing/{load-balance,app-protocols}` | 负载均衡/应用分流 |
| **同步任务** | `sync/{status,custom-isp,stream-domain}` | 分流任务状态/手动触发 |
| 鉴权 | `/api/v1/auth/login` | API Key 换 token |
| 健康 | `/health` `/live` `/ready` `/ping` `/version` | 探针（public） |

覆盖 SDK 151 个端点中的 ~90 个，横跨全部 12 个分组。

---

## 项目结构

```
ikuai-tools-service/
├── cmd/server/                    # 程序入口 + bootstrap（路由注册、中间件、依赖装配）
├── configs/                       # config.yaml(.example)
├── gen/                           # Hz 生成的 HTTP 路由/模型（health/common）— 勿手改
├── idl/                           # Hertz IDL（proto）定义
├── internal/
│   ├── ikuai/                     # 多路由器连接 Registry（核心）
│   ├── job/                       # 定时同步任务（custom_isp/stream_domain）+ 执行历史
│   ├── app/router/                # 路由器 CRUD service
│   ├── conf/                      # 配置结构与校验
│   ├── pkg/                       # logger / resp / errors / ctxkey
│   ├── repo/                      # db（GORM 模型/DAO）、redis
│   └── transport/http/            # handler（业务）+ middleware（auth/audit/cors/logger/recovery）
├── Dockerfile                     # 从 monorepo 根构建（见下）
└── Makefile
```

### 分层架构

```
HTTP 请求 → middleware（Recovery→Logger→CORS→Auth→Audit）
         → handler（按 :router_id 解析 Manager）
         → ikuai.Registry（取对应路由器的 *Client）
         → ikuai-api SDK（v4 REST / ActionCall 逃生舱）
         → 路由器

job（定时）→ Registry.Names() 遍历每台路由器 → SyncCustomISP/SyncStreamDomain → job_runs 落库
```

## 安全

- **鉴权**：`auth.api_key` 设置后，所有非 public 请求需带 `Authorization: Bearer <key>` 或 `X-API-Key: <key>`。**release 模式下空 key 会拒绝启动**；debug 模式允许无鉴权（仅限可信内网）。
- **审计**：所有写操作（POST/PUT/PATCH/DELETE）记录到 `audit_logs`（actor/method/path/router_id/status/req_id/ip）。
- **CORS**：`server.cors_origins` 配置允许的来源白名单，默认 `*`（可信内网）；公网部署应收紧。
- 路由器 token 明文存库（`routers.token`，API 响应已脱敏为 `has_token`），请保护数据库访问。

## 配置

完整配置项见 `configs/config.yaml.example`（含逐项注释）。关键项：

| 项 | 说明 |
|----|------|
| `server.mode` | `release` 强制鉴权；`debug` 允许无鉴权 |
| `server.cors_origins` | CORS 白名单，默认 `*` |
| `auth.api_key` | API Key，release 模式必填 |
| `ikuai.*` | 默认路由器（legacy；推荐用 `/api/v1/routers` API 管理） |
| `metrics.*` | 内嵌 Prometheus exporter（独立 9100 端口） |
| `jobs.*` | 定时同步任务（custom_isp / stream_domain） |

## 部署

### Docker

Dockerfile 从 **monorepo 根** 构建（服务通过本地 `replace` 依赖同级的 `ikuai-api/` 和 `ikuai_exporter/`）：

```bash
# 在包含 ikuai-api/ ikuai_exporter/ ikuai-tools-service/ 的目录下
docker build -t ikuai-tools-service -f ikuai-tools-service/Dockerfile .

# 运行（挂载真实配置 + 数据库卷）
docker run -d -p 9997:9997 -p 9100:9100 \
  -v $PWD/config.yaml:/app/configs/config.yaml \
  -v ikuai-data:/app/data \
  -e CONFIG_PATH=/app/configs/config.yaml \
  ikuai-tools-service
```

> 默认 SQLite 数据库落在 `/app`；如用 MySQL/PostgreSQL，在 config 里配 `database.*`。

## 技术栈

- [CloudWeGo Hertz](https://github.com/cloudwego/hertz) — HTTP 框架
- [ikuai-api](../ikuai-api) — iKuai v4 SDK
- [GORM](https://gorm.io/) — ORM（SQLite/MySQL/PostgreSQL）
- [gocron](https://github.com/go-co-op/gocron) — 定时任务
- [zap](https://github.com/uber-go/zap) — 结构化日志
- [Prometheus client_golang](https://github.com/prometheus/client_golang) — 指标

## 代码生成（IDL 驱动）

health/common 的路由由 Hz 从 `idl/` 生成到 `gen/`（已提交，开箱即用）。新增 HTTP 接口时：

```bash
make gen-http-update IDL=<proto>   # 更新；详见 Makefile
```

> 业务 handler（`internal/transport/http/handler/ikuai_*.go`）和路由注册（`bootstrap.go`）是手写的，不受代码生成影响。

---

## License

MIT
