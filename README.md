# nsk

NodeSeek 论坛的 Agent CLI。一次性子命令，stdout 默认瘦 JSON。无 TUI，无摸鱼 REPL。

站点是闭源自研 Vue SSR，不是 Discourse / Flarum / NodeBB。登录用 cookie `pjwt`。

完整规格见 [docs/design.md](docs/design.md)。领域词与路径见 [CONTEXT.md](CONTEXT.md)。语言、工程基线和运行时依赖见 [decisions/0001-technology-stack.md](decisions/0001-technology-stack.md)。

## 状态

#1 骨架与 XDG 配置已落地。#2 实现了 `nsk whoami` / `nsk cookie` 与 Chrome 124 warmup。列表/帖解析与 `nsk structure` 仍在后续票。

## Agent 用法

```bash
nsk whoami
nsk whoami --text
nsk cookie
nsk cookie --only
```

把 Cookie-Editor JSON 或单行 `pjwt=...` 写入 `$XDG_STATE_HOME/nsk/cookie.json`。导入剥掉 `cf_*`；warmup 成功才把本栈 jar 写回（0600）。Cloudflare 挑战或未登录不会覆盖该文件。

后续票才会提供：

```bash
nsk                     # 论坛结构：板块、分页、命令树
nsk cats
nsk list tech
nsk post 355740 --all
nsk search vps
nsk reply 355740 --body ./body.md
```

默认 JSON。`--text` 改人读。进度和错误在 stderr。

## 配置

`$XDG_CONFIG_HOME/nsk/config.toml`（默认 `~/.config/nsk/config.toml`）

Cookie 在 `$XDG_STATE_HOME/nsk/cookie.json`。示例：`config.toml.example`。

仅本机：只写 `[account]`，把 `pjwt` 放进 cookie 文件。持 cookie 的机器可再写 `[server]` 跑 `nsk server`；其它机器只写 `[client]`。

## 命令

实现后入口是 `just`（模板合同：`just check` 只读，`just audit` 单独上网）：

```bash
just check      # 质量门禁
just build      # 编译本地二进制
```
