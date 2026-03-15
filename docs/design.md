# Codeup Control 设计文档

## 1. 目标与范围

### 1.0 当前实现状态

本设计文档描述的是 `Codeup Control` 的整体设计目标与后续规划，不等同于当前版本已完整实现的功能集合。

当前代码中已经可用的核心能力主要包括：

- 认证：`cuctl auth login`、`logout`、`status`
- 仓库：`cuctl repo list`、`clone`、`view`、`set-default`
- 合并请求：`cuctl pr list`、`create`、`view`、`close`、`status`

其余命令仍可能处于占位或待实现状态。面向最终用户的当前可用命令，请以 `README.md` 为准。

### 1.1 目标

仿照 GitHub 的 `gh` CLI，为阿里云效 Codeup 提供命令行工具，使在身份认证后可在终端完成：

- 代码库：列表、克隆、查看
- 合并请求（MR/PR）：创建、列表、查看、合并
- 后续可扩展：分支、评审、Issue 等

### 1.2 非目标（当前版本）

- 不做代码迁移（迁移场景使用官方 Codeup-CLI）
- 不实现 Web 控制台的全部功能，仅覆盖常用开发流程

---

## 2. 技术选型

| 项目     | 选型        | 说明 |
|----------|-------------|------|
| 语言     | Go 1.25+    | 与 gh、kubectl 等主流 CLI 一致，单二进制、易分发 |
| CLI 框架 | Cobra       | 子命令、全局/局部 flag、帮助生成 |
| 配置     | 本地 YAML   | 可选 Viper 绑定，存储 host、org、默认分支等 |
| HTTP 客户端 | 标准库 `net/http`，后续可引入 retryablehttp 等 | 调用云效 OpenAPI |

---

## 3. 项目结构

```
codeup-control/
├── main.go                 # 入口，仅调用 cmd.Execute()
├── go.mod
├── go.sum
├── .gitignore              # 忽略构建产物
├── cmd/
│   ├── root.go             # 根命令、全局 flag（--config, --debug）、--version
│   ├── helpers.go          # 公共辅助（loadClientFromConfig 等）
│   ├── auth.go             # cuctl auth login/logout/status
│   ├── repo.go             # cuctl repo list/clone/view/set-default/create
│   ├── pr.go               # cuctl pr list/create/view/merge/checkout/diff/close/status/review/comment
│   └── branch.go           # cuctl branch list/create/delete/view/protect
├── internal/
│   ├── client/             # 云效 API 客户端（HTTP + PAT）
│   ├── config/             # 配置与凭证读写（PAT 存储位置、环境变量）
│   └── codeup/             # API 模型与端点封装（按需）
└── docs/
    └── design.md           # 本设计文档
```

- 对外仅暴露 `cmd` 包；业务与 API 封装放在 `internal/`，避免被外部引用。

---

## 4. 认证设计

### 4.1 当前：个人访问令牌（PAT）

- 云效 OpenAPI 当前文档仅支持使用 **个人访问令牌（PAT）** 调用 Codeup API。
- 调用方式：请求头 `x-yunxiao-token: <PAT>`（以云效官方文档为准）。
- PAT 由用户在云效工作台创建，Codeup Control 只负责**读取与存储**，不代用户创建。

### 4.2 凭证来源（优先级从高到低）

1. 环境变量：优先使用 `CODEUP_PERSONAL_ACCESS_TOKEN`、`CODEUP_ORGANIZATION_ID`、`CODEUP_DOMAIN`；兼容旧变量 `CODEUP_TOKEN`、`YUNXIAO_*`
2. 凭证文件：`~/.config/cuctl/credentials`（与 config.yaml 分离，权限 0600，兼容读取旧的 `~/.config/codeupcl/credentials`）
3. 交互式提示：`auth login` 时提示输入并写入凭证文件（输入不回显、不写日志）

### 4.3 配置与存储

- 配置文件路径：`--config` 指定，否则默认 `~/.config/cuctl/config.yaml`（或 XDG 约定）。
- 凭证建议与配置分离：如 `~/.config/cuctl/credentials` 或仅存 token 路径，权限设为 0600。
- 不在日志、错误信息中输出 token；`--debug` 仅打印请求 URL/方法，不打印 header 中的 token。

### 4.4 后续扩展：OAuth / 飞连

- 若云效开放「OAuth 设备流」或「用企业 IdP（如飞连）换取 API 令牌」：
  - 可新增 `cuctl auth login` 的浏览器/设备码流程，将取得的 token 写入上述凭证存储。
  - PAT 仍保留为备选（环境变量或配置文件），与 OAuth 二选一使用。

---

## 5. 命令设计

### 5.1 根命令

```bash
cuctl [--config PATH] [--debug] <command> [args]
```

- 全局 flag：`--config`、`--debug`，见 `cmd/root.go`。

### 5.2 auth（认证）

| 子命令        | 说明 |
|---------------|------|
| `cuctl auth login`  | 引导配置 PAT（写入配置/凭证）；若未来支持 OAuth 则走设备码流程 |
| `cuctl auth logout` | 清除本地存储的 token，不撤销云效侧 PAT |
| `cuctl auth status` | 显示当前使用的认证来源（env/credentials）及是否有效（可调一个只读 API 校验） |

### 5.3 repo（代码库）

| 子命令           | 优先级 | 说明 | 云效 API |
|------------------|:---:|------|----------|
| `cuctl repo list`   | P0 | 列出当前用户有权限的代码库 | ListRepositories |
| `cuctl repo clone [<org>/]<repo>` | P0 | 克隆指定仓库（API 取 URL + git clone） | GetRepository |
| `cuctl repo view [<org>/]<repo>`  | P0 | 显示仓库概要（默认分支、可见性、URL 等） | GetRepository |
| `cuctl repo set-default`  | P1 | 设置当前默认仓库（写入本地配置，后续命令可省略 org/repo） | 纯本地 |
| `cuctl repo create`  | P2 | 创建新仓库 | CreateRepository |

- 仓库标识：支持 `org/repo` 或仅 `repo`（用默认 org，来自配置或 `set-default`）。

### 5.4 pr（合并请求）

| 子命令           | 优先级 | 说明 | 云效 API |
|------------------|:---:|------|----------|
| `cuctl pr list`      | P0 | 列出当前仓库或指定仓库的 MR | ListChangeRequests |
| `cuctl pr create`    | P0 | 基于当前分支创建 MR（需目标分支、标题等） | CreateChangeRequest |
| `cuctl pr view [n]`  | P0 | 查看指定 MR 详情 | GetChangeRequest |
| `cuctl pr merge [n]` | P0 | 合并指定 MR | MergeMergeRequest |
| `cuctl pr checkout [n]` | P0 | 切到 MR 对应的源分支（取源分支名 + git fetch/checkout） | GetChangeRequest |
| `cuctl pr diff [n]`  | P0 | 查看 MR 的代码变更 | GetCompare / MR diff 数据 |
| `cuctl pr close [n]` | P0 | 关闭 MR | CloseChangeRequest |
| `cuctl pr status`    | P0 | 显示当前分支关联的 MR 状态 | ListChangeRequests（按分支过滤） |
| `cuctl pr review [n]` | P1 | 提交评审（通过 PASS / 拒绝 NOT_PASS） | ReviewChangeRequest |
| `cuctl pr comment [n]` | P1 | 添加评论（全局或行内） | CreateChangeRequestComment |
| `cuctl pr edit [n]`  | P2 | 编辑 MR 标题/描述/评审人 | 待确认 |
| `cuctl pr ready [n]` | P2 | 标记草稿 MR 为就绪 | 待确认 |
| `cuctl pr reopen [n]` | P2 | 重新打开已关闭的 MR | 待确认 |

### 5.5 branch（分支）

gh 没有独立的 `branch` 命令组，但云效有完整的分支 API，且分支操作在日常开发中高频使用，因此 Codeup Control 新增此命令组。

| 子命令           | 优先级 | 说明 | 云效 API |
|------------------|:---:|------|----------|
| `cuctl branch list`   | P1 | 列出仓库分支 | ListBranches |
| `cuctl branch create <name> [--from ref]` | P1 | 创建分支（可指定来源分支/tag/commit） | CreateBranch |
| `cuctl branch delete <name>` | P1 | 删除分支 | DeleteBranch |
| `cuctl branch view <name>` | P2 | 查看分支详情（最新 commit、是否受保护等） | GetBranchInfo |
| `cuctl branch protect <name>` | P2 | 查看/设置保护规则 | CreateProtectedBranch / DeleteProtectedBranch |

---

## 6. 云效 API 对接要点

- **Base URL / 域名**：以云效文档为准（如 `codeup.aliyun.com` 或专属域名），可配置。
- **认证 Header**：`x-yunxiao-token: <PAT>`（具体 header 名以最新文档为准）。
- **组织与仓库**：多数接口需要 `organizationId`、`repositoryId`，可从列表接口或配置中解析。
- **错误处理**：HTTP 4xx/5xx 统一解析为可读错误，避免把 token 或内部信息打到 stderr。

### 6.1 API 端点速查

| 功能域 | 操作 | 方法 | 端点（省略前缀） |
|--------|------|------|-------------------|
| 仓库 | 列表 | GET | `/organizations/{orgId}/repositories` |
| 仓库 | 详情 | GET | `/organizations/{orgId}/repositories/{repoId}` |
| 仓库 | 创建 | POST | `/organizations/{orgId}/repositories` |
| 分支 | 列表 | GET | `/organizations/repositories/{repoId}/branches` |
| 分支 | 详情 | GET | `/organizations/repositories/{repoId}/branches/{name}` |
| 分支 | 创建 | POST | `/organizations/{orgId}/repositories/{repoId}/branches` |
| 分支 | 删除 | DELETE | `/organizations/{orgId}/repositories/{repoId}/branches/{name}` |
| 合并请求 | 列表 | GET | `/organizations/{orgId}/changeRequests` |
| 合并请求 | 创建 | POST | `/organizations/{orgId}/repositories/{repoId}/changeRequests` |
| 合并请求 | 详情 | GET | `/organizations/{orgId}/repositories/{repoId}/changeRequests/{id}` |
| 合并请求 | 合并 | PUT/POST | MergeMergeRequest |
| 合并请求 | 关闭 | PUT/POST | CloseChangeRequest |
| 合并请求 | 评审 | POST | `/…/changeRequests/{id}/review` |
| 合并请求 | 评论 | POST | `/…/changeRequests/{id}/comments` |

> 端点以 `https://{domain}/oapi/v1/codeup` 为前缀，具体以阿里云最新文档为准。

---

## 7. 依赖与构建

- 直接依赖：`github.com/spf13/cobra`、`gopkg.in/yaml.v3`、`golang.org/x/term`；后续可按需引入 Viper、表格式输出库等。
- 构建：`go build -o cuctl .`；可加 `-ldflags "-s -w"` 减小体积。
- 测试：关键路径为 `internal/client`、`internal/config` 及 `cmd` 的单元测试；API 部分可 mock HTTP。

---

## 8. 版本与迭代

- **v0.1**：认证（`auth login/logout/status`）+ `repo list`。
- **v0.2**：`repo clone`、`repo view`、`repo set-default`。
- **v0.3**：`pr list`、`pr create`、`pr view`、`pr merge`。
- **v0.4**：`pr checkout`、`pr diff`、`pr close`、`pr status`。
- **v0.5**：`pr review`、`pr comment`、`branch list/create/delete`。
- **v0.6**：`repo create`、`branch view/protect`、`pr edit/ready/reopen`。
- 后续：OAuth / 飞连（若云效支持）、与云效 Flow（CI/CD）集成等。

---

## 9. 参考

- 云效 Codeup OpenAPI：阿里云 OpenAPI 门户 codeup 2020-04-14。
- 个人访问令牌：云效帮助中心「个人访问令牌」「如何使用个人访问令牌调用 API」。
- GitHub CLI (gh)：命令与交互风格参考。
