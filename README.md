# nsk

NodeSeek 论坛的 Agent CLI。一次性子命令，stdout 默认瘦 JSON。无 TUI，无摸鱼 REPL。

站点是闭源自研 Vue SSR，不是 Discourse / Flarum / NodeBB。登录用 cookie `pjwt`。

完整规格见 [docs/design.md](docs/design.md)。领域词与路径见 [CONTEXT.md](CONTEXT.md)。语言、工程基线和运行时依赖见 [decisions/0001-technology-stack.md](decisions/0001-technology-stack.md)。

## 状态

v1 已落地：一次性 Agent CLI、cookie Account、`nsk server` / `[client]` remote Forum。无 `nsk reply`。tag `v*` 走 GoReleaser（linux/darwin/windows × amd64/arm64，`CGO_ENABLED=0`）。

## 安装

从 [GitHub Releases](https://github.com/solanab/nsk/releases) 下载对应 `nsk_*_<os>_<arch>.tar.gz`，解压出 `nsk`（Windows 为 `nsk.exe`）放到 `PATH`。

本地构建：

```bash
just install
just build          # dist/nsk
```

校验：`nsk --version`。

## Cookie 与 whoami

Agent 默认 stdout 瘦 JSON。不要开 TUI / MCP / REPL。一次子命令一个资源。

1. 浏览器登录 <https://www.nodeseek.com>。
2. 用 Cookie-Editor 导出 JSON 数组，或复制单行 `pjwt=...`。
3. 写入 `$XDG_STATE_HOME/nsk/cookie.json`（默认 `~/.local/state/nsk/cookie.json`），权限 `0600`。
4. 仅本机可不写 config。需要提示用户名时放 `$XDG_CONFIG_HOME/nsk/config.toml`（默认 `~/.config/nsk/config.toml`）：

```toml
[account]
username = "your_username"
```

5. 跑通登录：

```bash
nsk whoami
```

成功则 stdout 是 `UserInfo` JSON（`id` / `name` / 可选 `chicken` / `level`）。失败在 stderr：缺文件会指出路径；Cloudflare 或过期 cookie 不会覆盖 `cookie.json`。导入剥掉 `cf_*`；warmup 用 Chrome 124 自拿 CF cookie。

不要把 cookie 拷到其它机器。持 cookie 的机器可加 `[server]` 跑 `nsk server`；其它机器只写 `[client]` `url` / `token`，然后照常 `nsk whoami`。`nsk cookie` 在 Client 机 fatal。

v1 无密码登录、无签到、无鸡腿任务、无用户名查找（`nsk user` 只要数字 id）。

## Agent 用法

```bash
nsk                     # 论坛结构：板块、分页、命令树
nsk structure
nsk cats
nsk list
nsk list tech
nsk list --page 2
nsk post 355740
nsk post 355740 --page 2
nsk post 355740 --all
nsk post 355740/4
nsk post 355740 -o
nsk search vps
nsk search vps --page 2
nsk whoami
nsk whoami --text
nsk user 42
nsk user 42 --text
nsk notify
nsk notify --text
nsk cookie
nsk cookie --only
nsk server
nsk server --addr 127.0.0.1:9200
```

把 Cookie-Editor JSON 或单行 `pjwt=...` 写入 `$XDG_STATE_HOME/nsk/cookie.json`。导入剥掉 `cf_*`；warmup 成功才把本栈 jar 写回（0600）。Cloudflare 挑战或未登录不会覆盖该文件。搜索失败不会降级到公开最新列表。

v1 不做 `nsk reply`（见 [docs/forum-write.md](docs/forum-write.md)）。

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
