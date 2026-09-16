# nsk

NodeSeek 论坛的 Agent CLI。一份配置描述这台机器的角色：自己登录论坛，或连到另一台持有登录的机器。

主用户是 Agent：一次性子命令，stdout 默认瘦 JSON。无 TUI，无摸鱼 REPL。无参数 = `nsk structure`。

## Language

**Account**:
当前这台机器用来访问 NodeSeek 的论坛身份（`pjwt` cookie，可选 username 提示）。
_Avoid_: auth, session, user, login

**Server**:
持有 Account、替其它机器访问论坛的那台 `nsk` 进程。
_Avoid_: listen, gateway, daemon, remote

**Client**:
不持有 Account、只连接 Server 的那台 `nsk` 进程。
_Avoid_: remote, thin, slave

**Forum**:
www.nodeseek.com 站点本身（闭源自研 Vue SSR，不是 Discourse）。
_Avoid_: upstream, backend, API

**Post / Floor**:
帖与楼层。不用 Discourse 的 Topic / PostStream。

## Paths / Config / Env

**Config**:
`$XDG_CONFIG_HOME/nsk/config.toml`（默认 `~/.config/nsk/config.toml`）。`nsk --config PATH` 可覆盖（对 `nsk server` 同样有效）。不查找 cwd、`~/.nsk/` 或其它目录。TOML 用 `pelletier/go-toml/v2`。一次性命令默认 stdout 瘦 JSON，`--text` 改人读。

**State**:
不在配置目录。Cookie：`$XDG_STATE_HOME/nsk/cookie.json`（默认 `~/.local/state/nsk/cookie.json`）。`[account].cookie` 相对路径相对 state 目录，绝对路径仍可用。无 CLI history 文件。

**Env**:
`NSK_*` 覆盖同名 TOML 字段：`NSK_USERNAME`、`NSK_SERVER_ADDR`、`NSK_SERVER_TOKEN`、`NSK_CLIENT_URL`、`NSK_CLIENT_TOKEN`。无密码环境变量，无 cookie 路径环境变量。

**Config shapes**:
`[server]` 与 `[client]` 互斥（文件或环境变量两边都设则报错）。Server 机器：`[account]` + `[server]`，运行 `nsk server`。Client 机器：只写 `[client]`，不拷 cookie。仅本机：只写 `[account]`。`[server] addr = "HOST:PORT"`；`nsk server --addr HOST:PORT`。空 addr 时默认绑定 `127.0.0.1:9200`。非 loopback 必须有 token；拒绝 `0.0.0.0` 与 `::`。token 可写在 0600 toml；HTTP Authorization Bearer 不变。

## Cookie

v1 只导入 `pjwt`（Cookie-Editor JSON 或单行 `pjwt=...`）。导入剥掉 `cf_*`。warmup 用 Chrome 124 TLS 自拿 CF cookie。warmup 失败不覆盖 `CookieFile()`。`nsk cookie` 导出本进程 jar；`nsk cookie --only` 打 `pjwt=`。`[client]` 机器上 `nsk cookie` fatal。

## Agent surface

```
nsk                     # SiteStructure
nsk cats
nsk list [slug] [--page N]
nsk post <id> [--page N|--all]
nsk whoami
nsk user <id>
nsk search <q> [--page N]
nsk notify
nsk reply <id> --body FILE
nsk server [--addr HOST:PORT] [--token TOKEN]
```

## Stack

Read `decisions/0001-technology-stack.md` before adding a runtime dependency or changing the CLI parser, HTTP client, HTML parser, or quality baseline.
