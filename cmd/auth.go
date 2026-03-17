package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hwuu/codeup-control/internal/client"
	"github.com/hwuu/codeup-control/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "管理认证凭证",
	Long:  "登录、登出以及查看当前认证状态。",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "配置认证凭证（当前为 PAT）",
	RunE:  runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "清除本地存储的凭证",
	RunE:  runAuthLogout,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示当前认证状态",
	RunE:  runAuthStatus,
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "输出当前存储的令牌",
	RunE:  runAuthToken,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd, authTokenCmd)
}

func prompt(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func promptSecret(label string) (string, error) {
	fmt.Printf("%s: ", label)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}
	return strings.TrimSpace(string(bytes)), nil
}

func friendlyTokenSource(source string) string {
	switch source {
	case "env:CODEUP_PERSONAL_ACCESS_TOKEN":
		return "环境变量 CODEUP_PERSONAL_ACCESS_TOKEN"
	case "env:CODEUP_TOKEN":
		return "环境变量 CODEUP_TOKEN"
	case "env:YUNXIAO_TOKEN":
		return "环境变量 YUNXIAO_TOKEN"
	case "credentials":
		return "本地凭证文件"
	default:
		return source
	}
}

func friendlyOrgSource(source string) string {
	switch source {
	case "env:CODEUP_ORGANIZATION_ID":
		return "环境变量 CODEUP_ORGANIZATION_ID"
	case "env:YUNXIAO_ORGANIZATION_ID":
		return "环境变量 YUNXIAO_ORGANIZATION_ID"
	case "config":
		return "配置文件"
	default:
		return source
	}
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(GlobalCfgFile)
	if err != nil {
		return err
	}

	fmt.Println("请在云效工作台 > 个人设置 > 个人访问令牌 中创建 PAT。")
	fmt.Println("权限至少需要：read:repo")
	fmt.Println()

	token, err := promptSecret("个人访问令牌 (PAT)")
	if err != nil {
		return err
	}
	if token == "" {
		return fmt.Errorf("令牌不能为空")
	}

	reader := bufio.NewReader(os.Stdin)
	orgID := prompt(reader, "组织 ID (organizationId)", cfg.OrganizationID)
	if orgID == "" {
		return fmt.Errorf("组织 ID 不能为空")
	}

	domain := prompt(reader, "服务域名", cfg.ResolveDomain())

	c := client.New(domain, token, GlobalDebug)
	user, err := c.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("令牌验证失败: %w", err)
	}
	if _, err := c.ListRepositories(orgID, 1, 1, ""); err != nil {
		return fmt.Errorf("组织 ID 校验失败，请确认当前账号可访问该组织: %w", err)
	}

	if err := config.SaveToken(GlobalCfgFile, token); err != nil {
		return err
	}

	cfg.OrganizationID = orgID
	cfg.Domain = domain
	if err := cfg.Save(GlobalCfgFile); err != nil {
		return err
	}

	fmt.Printf("\n登录成功！用户: %s (%s)\n", user.Name, user.Username)
	fmt.Printf("配置已保存至: %s\n", config.Path(GlobalCfgFile))
	fmt.Printf("凭证已保存至: %s\n", config.CredentialsPath(GlobalCfgFile))
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	stored, err := config.LoadToken(GlobalCfgFile)
	if err != nil {
		return err
	}
	if stored == "" {
		fmt.Println("当前配置中没有存储令牌。")
	} else {
		if err := config.ClearToken(GlobalCfgFile); err != nil {
			return err
		}
		fmt.Println("已清除本地令牌。（云效侧令牌未撤销，如需撤销请前往云效工作台操作）")
	}

	if v := os.Getenv("CODEUP_PERSONAL_ACCESS_TOKEN"); v != "" {
		fmt.Println("注意: 检测到环境变量 CODEUP_PERSONAL_ACCESS_TOKEN 仍然生效，如需完全登出请取消该变量。")
	} else if v := os.Getenv("CODEUP_TOKEN"); v != "" {
		fmt.Println("注意: 检测到环境变量 CODEUP_TOKEN 仍然生效，如需完全登出请取消该变量。")
	} else if v := os.Getenv("YUNXIAO_TOKEN"); v != "" {
		fmt.Println("注意: 检测到环境变量 YUNXIAO_TOKEN 仍然生效，如需完全登出请取消该变量。")
	}
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(GlobalCfgFile)
	if err != nil {
		return err
	}

	token, source, err := config.ResolveToken(GlobalCfgFile)
	if err != nil {
		return err
	}
	if token == "" {
		fmt.Println("未认证。请运行 cuctl auth login 或设置环境变量 CODEUP_PERSONAL_ACCESS_TOKEN。")
		return nil
	}

	orgID, orgSource := cfg.ResolveOrganizationID()
	fmt.Printf("令牌来源: %s\n", friendlyTokenSource(source))
	if orgID != "" {
		fmt.Printf("组织 ID:  %s (%s)\n", orgID, friendlyOrgSource(orgSource))
	} else {
		fmt.Println("组织 ID:  未配置")
	}
	fmt.Printf("服务域名: %s\n", cfg.ResolveDomain())

	c := client.New(cfg.ResolveDomain(), token, GlobalDebug)
	user, err := c.GetCurrentUser()
	if err != nil {
		fmt.Printf("令牌状态: 无效或已过期 (%v)\n", err)
		return nil
	}

	fmt.Printf("登录用户: %s (%s)\n", user.Name, user.Username)
	if user.Email != "" {
		fmt.Printf("邮箱:     %s\n", user.Email)
	}
	return nil
}

func runAuthToken(cmd *cobra.Command, args []string) error {
	token, _, err := config.ResolveToken(GlobalCfgFile)
	if err != nil {
		return err
	}
	if token == "" {
		return fmt.Errorf("未认证，请先运行: cuctl auth login")
	}
	fmt.Println(token)
	return nil
}
