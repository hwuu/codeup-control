# Codeup Control

Codeup Control 是一个云效 Codeup 命令行工具，命令名为 `cuctl`，使用 Go + Cobra 实现。

当前版本已支持认证、仓库查询和部分合并请求操作。

## 构建

```bash
go mod tidy
go build -o cuctl .
```

## 当前状态

当前已实现的命令：

- `cuctl auth login`
- `cuctl auth logout`
- `cuctl auth status`
- `cuctl repo list`
- `cuctl repo clone`
- `cuctl repo view`
- `cuctl repo set-default`
- `cuctl pr list`
- `cuctl pr create`
- `cuctl pr view`
- `cuctl pr close`
- `cuctl pr status`

已注册但暂未实现的命令：

- `cuctl repo create`
- `cuctl pr merge`
- `cuctl pr checkout`
- `cuctl pr diff`
- `cuctl pr review`
- `cuctl pr comment`
- `cuctl branch list`
- `cuctl branch create`
- `cuctl branch delete`

## 快速开始

```bash
# 1. 登录（输入 PAT 和组织 ID）
./cuctl auth login

# 2. 查看认证状态
./cuctl auth status

# 3. 列出代码库
./cuctl repo list
./cuctl repo list --search myproject --limit 10
```

也可通过环境变量传入令牌和组织 ID，跳过 `auth login`：

```bash
export CODEUP_PERSONAL_ACCESS_TOKEN=pt-xxxx
export CODEUP_ORGANIZATION_ID=your-org-id
export CODEUP_DOMAIN=openapi-rdc.aliyuncs.com
./cuctl repo list
```

兼容性说明：当前版本优先使用 `CODEUP_PERSONAL_ACCESS_TOKEN`、`CODEUP_ORGANIZATION_ID`、`CODEUP_DOMAIN`，并兼容旧变量 `CODEUP_TOKEN`、`YUNXIAO_*`。

## 配置与认证

- 默认配置文件：`~/.config/cuctl/config.yaml`
- 默认凭证文件：`~/.config/cuctl/credentials`
- 兼容读取旧凭证目录：`~/.config/codeupcl/`
- PAT 通过请求头 `x-yunxiao-token` 调用云效 OpenAPI

推荐环境变量：

- `CODEUP_PERSONAL_ACCESS_TOKEN`
- `CODEUP_ORGANIZATION_ID`
- `CODEUP_DOMAIN`

兼容环境变量：

- `CODEUP_TOKEN`
- `YUNXIAO_TOKEN`
- `YUNXIAO_ORGANIZATION_ID`
- `YUNXIAO_DOMAIN`

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
# 查看认证状态
./cuctl auth status

# 查看仓库详情
./cuctl repo view my-group/my-repo

# 设置默认仓库后，可在 pr 命令中省略 --repo
./cuctl repo set-default my-group/my-repo

# 查看当前仓库的打开中合并请求
./cuctl pr list --repo my-group/my-repo

# 基于当前 git 分支创建合并请求
./cuctl pr create --repo my-group/my-repo --title "feat: add demo"

# 查看当前分支关联的合并请求状态
./cuctl pr status --repo my-group/my-repo
```

## 已知说明

- `auth login` 当前会同时校验令牌有效性和 `organizationId` 可访问性。
- 仓库和合并请求接口仍依赖正确的 `organizationId`。
- `repo clone` 只负责解析仓库克隆地址；若使用 HTTPS，仍依赖本机已有的 Git 凭证配置，未配置时建议使用 `--ssh`。
- `docs/design.md` 描述的是整体设计与后续规划，命令覆盖范围比当前实现更完整。

## 设计

详见 [docs/design.md](docs/design.md)。
