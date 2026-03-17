<div align="center">

# codeup-control

云效 Codeup 命令行工具  
用一个 `cuctl` 管理代码库、合并请求和分支

[你可以拿它做什么](#你可以拿它做什么) • [快速开始](#快速开始) • [命令概览](#命令概览) • [设计文档](./docs/design.md)

</div>

---

## 你可以拿它做什么

在终端里直接操作云效 Codeup：

- 登录认证，查看认证状态
- 列出、查看、克隆、创建代码库
- 创建、查看、合并、关闭合并请求
- 查看 PR diff、提交评审、添加评论
- 切换到 PR 对应的分支
- 列出、创建、删除分支

适用场景：

- 不想频繁切到浏览器操作 Codeup
- 在脚本或 CI 中调用 Codeup API
- 快速查看和管理合并请求

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
| `auth`   | `login`, `logout`, `status` |
| `repo`   | `list`, `clone`, `view`, `set-default`, `create` |
| `pr`     | `list`, `create`, `view`, `merge`, `checkout`, `diff`, `close`, `status`, `review`, `comment` |
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

# 设置默认仓库后，pr 命令可省略 --repo
./cuctl repo set-default my-group/my-repo

# 查看打开中的合并请求
./cuctl pr list --repo my-group/my-repo

# 基于当前 git 分支创建合并请求
./cuctl pr create --repo my-group/my-repo --title "feat: add demo"

# 查看当前分支关联的 PR 状态
./cuctl pr status --repo my-group/my-repo
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

已实现的命令：

- `cuctl auth login` / `logout` / `status`
- `cuctl repo list` / `clone` / `view` / `set-default`
- `cuctl pr list` / `create` / `view` / `close` / `status`

已注册但暂未实现：

- `cuctl repo create`
- `cuctl pr merge` / `checkout` / `diff` / `review` / `comment`
- `cuctl branch list` / `create` / `delete`

## 已知说明

- `auth login` 会同时校验令牌有效性和 `organizationId` 可访问性
- 仓库和合并请求接口依赖正确的 `organizationId`
- `repo clone` 只负责解析克隆地址；HTTPS 方式依赖本机 Git 凭证配置，未配置时建议使用 `--ssh`

## 设计

详见 [docs/design.md](docs/design.md)。
