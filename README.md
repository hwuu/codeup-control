<div align="center">

# codeup-control

云效 Codeup 命令行工具  
用一个 `cuctl` 管理代码库、合并请求和分支

[你可以拿它做什么](#你可以拿它做什么) • [快速开始](#快速开始) • [命令概览](#命令概览) • [设计文档](./docs/design.md)

</div>

---

## 你可以拿它做什么

在终端里直接操作云效 Codeup：

- 登录认证，查看认证状态，输出令牌
- 列出、查看、克隆、创建、编辑、删除、Fork、归档、重命名代码库
- 创建、查看、编辑、合并、关闭、重新打开合并请求
- 查看 PR diff、提交评审、添加评论、标记就绪
- 切换到 PR 对应的分支
- 列出、创建、删除分支

适用场景：

- 不想频繁切到浏览器操作 Codeup
- 在脚本或 CI 中调用 Codeup API
- 快速查看和管理合并请求

## 设计原则

**与 GitHub CLI (`gh`) 使用方法兼容。** `cuctl` 的命令层级、子命令名称、flag 命名与交互行为均对齐 [`gh`](https://cli.github.com/)，熟悉 `gh` 的用户可以直接上手：

- 命令组对齐：`auth`、`repo`、`pr` 与 `gh` 一一对应
- 子命令同名同义：`list`、`create`、`view`、`edit`、`merge`、`checkout`、`diff`、`close`、`reopen`、`ready` 等
- Flag 对齐：`-R/--repo`、`-t/--title`、`-b/--body`、`-B/--base`、`-d/--delete-branch` 等沿用 `gh` 的命名与缩写
- 行为对齐：不指定 PR 编号时自动关联当前 git 分支，与 `gh` 一致
- Codeup 扩展：`branch` 命令组为云效特有（`gh` 无对应命令），以相同风格新增

## 快速开始

### 构建

使用 Go + Cobra 实现。

```bash
go mod tidy
go build -o cuctl .
```

### 登录并使用

```bash
# 登录（输入 PAT 和组织 ID）
./cuctl auth login

# 查看认证状态
./cuctl auth status

# 列出代码库
./cuctl repo list
./cuctl repo list --search myproject --limit 10
```

也可以通过环境变量跳过交互式登录：

```bash
export CODEUP_PERSONAL_ACCESS_TOKEN=pt-xxxx
export CODEUP_ORGANIZATION_ID=your-org-id
export CODEUP_DOMAIN=openapi-rdc.aliyuncs.com
./cuctl repo list
```

## 命令概览

| 命令组   | 子命令 |
|----------|--------|
| `auth`   | `login`, `logout`, `status`, `token` |
| `repo`   | `list`, `clone`, `view`, `create`, `edit`, `delete`, `fork`, `archive`, `unarchive`, `rename`, `set-default` |
| `pr`     | `list`, `create`, `view`, `edit`, `merge`, `checkout`, `diff`, `close`, `reopen`, `ready`, `status`, `review`, `comment` |
| `branch` | `list`, `create`, `delete` |

使用 `--help` 查看任意命令的详细用法：

```bash
./cuctl pr --help
./cuctl repo list --help
```

## 常用示例

```bash
# 查看仓库详情
./cuctl repo view my-group/my-repo

# 创建代码库
./cuctl repo create my-new-repo --namespace my-group --visibility private --init-readme

# 设置默认仓库后，pr/branch 命令可省略 --repo
./cuctl repo set-default my-group/my-repo

# 查看打开中的合并请求
./cuctl pr list --repo my-group/my-repo

# 基于当前 git 分支创建合并请求
./cuctl pr create --repo my-group/my-repo --title "feat: add demo"

# 合并请求（不指定编号则自动匹配当前分支）
./cuctl pr merge 42 --delete-branch
./cuctl pr merge --type squash

# 切到 PR 对应的分支
./cuctl pr checkout 42

# 查看 PR 代码变更
./cuctl pr diff 42

# 编辑 PR 标题或描述（不指定编号则匹配当前分支）
./cuctl pr edit 42 --title "fix: 修复登录问题" --body "更新了描述"

# 标记草稿 PR 为就绪
./cuctl pr ready 42

# 重新打开已关闭的 PR
./cuctl pr reopen 42

# 评审和评论
./cuctl pr review 42 --approve
./cuctl pr comment 42 --body "LGTM"

# 查看当前分支关联的 PR 状态
./cuctl pr status --repo my-group/my-repo

# 编辑仓库设置
./cuctl repo edit my-group/my-repo --description "新描述" --default-branch develop

# 删除仓库（需确认）
./cuctl repo delete my-group/my-repo
./cuctl repo delete my-group/my-repo --yes

# Fork / 归档 / 重命名仓库
./cuctl repo fork my-group/my-repo
./cuctl repo archive my-group/my-repo
./cuctl repo rename new-name --repo my-group/my-repo

# 输出令牌（用于脚本）
./cuctl auth token

# 列出分支
./cuctl branch list --repo my-group/my-repo

# 创建和删除分支
./cuctl branch create feat/new-feature --from main --repo my-group/my-repo
./cuctl branch delete feat/old-feature --repo my-group/my-repo
```

## 配置与认证

默认配置路径：

- 配置文件：`~/.config/cuctl/config.yaml`
- 凭证文件：`~/.config/cuctl/credentials`
- 兼容旧目录：`~/.config/codeupcl/`

PAT 通过请求头 `x-yunxiao-token` 调用云效 OpenAPI。

推荐环境变量：

- `CODEUP_PERSONAL_ACCESS_TOKEN`
- `CODEUP_ORGANIZATION_ID`
- `CODEUP_DOMAIN`

兼容环境变量：

- `CODEUP_TOKEN`
- `YUNXIAO_TOKEN`
- `YUNXIAO_ORGANIZATION_ID`
- `YUNXIAO_DOMAIN`

## 当前状态

所有设计命令均已实现：

- `cuctl auth login` / `logout` / `status` / `token`
- `cuctl repo list` / `clone` / `view` / `create` / `edit` / `delete` / `fork` / `archive` / `unarchive` / `rename` / `set-default`
- `cuctl pr list` / `create` / `view` / `edit` / `merge` / `checkout` / `diff` / `close` / `reopen` / `ready` / `status` / `review` / `comment`
- `cuctl branch list` / `create` / `delete`

## 已知说明

- `auth login` 会同时校验令牌有效性和 `organizationId` 可访问性
- 仓库和合并请求接口依赖正确的 `organizationId`
- `repo clone` 只负责解析克隆地址；HTTPS 方式依赖本机 Git 凭证配置，未配置时建议使用 `--ssh`

## 设计

详见 [docs/design.md](docs/design.md)。
