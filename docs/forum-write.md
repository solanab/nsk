# NodeSeek 写接口

v1 **不做** `Reply`。不把 `Reply` 放进 `Forum`，不提供 `nsk reply`，Server 不挂 replies 路由。

本文件是写接口门闩。门闩要求钉死**一条** POST：URL、method、headers（cookie、content-type、origin/referer、CSRF）、body
schema、成功 JSON → `Floor`、失败文本 → 错误。猜出来的契约不算钉死。

## 判定

**拒绝。** 第三方客户端对 URL 有共识，但对 body、CSRF、成功体没有单一可复现契约。本仓库夹具也钉不住 CSRF 或成功 JSON。

## 有共识的部分

第三方实现一致指向：

| 项           | 观察                                                    |
| ------------ | ------------------------------------------------------- |
| URL          | `POST https://www.nodeseek.com/api/content/new-comment` |
| Cookie       | 需要 `pjwt`                                             |
| Content-Type | `application/json`                                      |
| Referer      | 多数带 `https://www.nodeseek.com/post-{id}-1`           |
| Origin       | 部分带 `https://www.nodeseek.com`                       |

这**不够**构成 v1 契约。缺 CSRF、body、成功 JSON → `Floor`。

## 钉不住的部分

### Body

至少两套互相冲突的 JSON：

1. `{ "content": "<markdown>", "mode": "new-comment", "postId": <int> }` （fluxseek、wz-android、KonglingDaDa
   userscript、nodeseek-comment-preview）
2. `{ "content": "<markdown>", "post_id": <int> }` （tanaer/nodeseeksignin）

没有本仓库夹具证明哪一套是站点当前唯一接受的 schema。v1 不猜。

### CSRF

第三方各走各的，且本仓库 HTML 夹具对不上：

- 随机 16 位字母数字，header `csrf-token`（fluxseek 回帖路径 `skipCsrf`、seekmate、comment-preview、wz-android）
- 页面 `meta[name="csrf-token"]` 或 HTML 里的 `csrfToken`，header `Csrf-Token`（KonglingDaDa、tanaer）
- 另有 `x-csrf-challenge: simple-token` + 缓存的 `X-CSRF-Token`（fluxseek 拦截器文档；回帖路径却 `skipCsrf: true`）
- comment-preview 还加 `x-dynamic-sign`（SHA-1 of method/url/UA/body）

本仓库夹具：

- `window.__config__` 只有 `user` / `site`，没有 CSRF
- `post-703863-1.html` 没有 `meta[name="csrf-token"]`

无法钉死「从哪读、用哪个 header 名、值是否必须来自页面」。

### 成功 JSON → Floor

没有一份保存下来的成功响应能映射到 `Floor`（`floor` / `author` / `created_at` / `markdown` / `reply_to`）。

观察到的成功形态：

- `{ "success": true, "redirectHash": "#<floor>" }`（fluxseek：再 GET 帖页拼楼层）
- `{ "success": false, "message": "..." }` 当失败
- 部分实现把非 JSON / 空 body 也当成功

`redirectHash` 不是 `Floor`。再 GET 帖页不是「唯一 POST」。失败文本候选是 `message`，但没有夹具钉死字段名与 HTTP
状态的组合。

## 来源（只读，未对站点 POST）

- [tyrad/nodeseek CommentSubmissionAutomationScript](https://github.com/tyrad/nodeseek/blob/1465e3a46832424ceb3c0047b8713bc29195a7c7/nodeseek/AppRuntime/Web/AutomationScripts/CommentSubmissionAutomationScript.swift)
  — 只拦截 `/api/content/new-comment`，走页面按钮，不钉 body
- [yreidev/fluxseek `_posts.dart`](https://github.com/yreidev/fluxseek/blob/46da249c1ff722b0de41a46e048fff1cd5ef9be1/lib/services/nodeseek/client_parts/_posts.dart)
  与
  [CSRF 文档](https://github.com/yreidev/fluxseek/blob/46da249c1ff722b0de41a46e048fff1cd5ef9be1/.trellis/spec/network/cookies-and-cloudflare.md)
- [everythink98/wz-android actionRequest.ts](https://github.com/everythink98/wz-android/blob/9304443206fe0d39db52aead39a80a88233d490c/src/sources/nodeseek/actionRequest.ts)
- [moxuun/nodeseek-comment-preview action-api.js](https://github.com/moxuun/nodeseek-comment-preview/blob/52bb842eb539e41bcd82af670cb61bbd0332952b/src/nodeseek/action-api.js)
- [KonglingDaDa/NodeSeekB userscript](https://github.com/KonglingDaDa/NodeSeekB/blob/dbc09be380f4ae74f0b3702c1dc69bc85d856303/nodeseek-auto-reply.user.js)
- [tanaer/nodeseeksignin worker](https://github.com/tanaer/nodeseeksignin/blob/4378bc29f919aec4f6e1c740e6328e0a6bf13910/worker/src/nodeseek.ts)
- 本仓库 `internal/client/testdata/html/{homepage-user.html,post-703863-1.html}`

没有对本机 `pjwt` 做 live POST。写操作不以第三方脚本代替夹具。

## v1 后果

- `Forum` 不含 `Reply`
- CLI 无 `reply`；`SiteStructure.commands` 不含该名
- `nsk server` 不提供 `POST /api/v1/posts/{id}/replies`
- GitHub `#10` 关闭为 wontfix

以后若有**本仓库**保存的成功/失败 JSON 夹具，并能钉死单一 body 与 CSRF，再开新票，不复活本文件里的猜测。
