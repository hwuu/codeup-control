package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "管理云效代码库",
	Long:  "列出、克隆、查看、创建云效 Codeup 代码库。",
}

// --- repo list ---

var repoListPage int
var repoListPerPage int
var repoListSearch string
var repoCloneUseSSH bool

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出有权限的代码库",
	RunE:  runRepoList,
}

// --- repo clone ---

var repoCloneCmd = &cobra.Command{
	Use:   "clone [<org>/]<repo>",
	Short: "克隆代码库",
	Long:  "克隆指定代码库。使用 HTTPS 时仍依赖本机已有的 Git 凭证配置；如未配置，建议使用 --ssh。",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoClone,
}

// --- repo view ---

var repoViewCmd = &cobra.Command{
	Use:   "view [<org>/]<repo>",
	Short: "查看代码库概要",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoView,
}

// --- repo set-default ---

var repoSetDefaultCmd = &cobra.Command{
	Use:   "set-default [<org>/]<repo>",
	Short: "设置当前默认仓库",
	Long:  "设置默认仓库后，后续命令可省略 org/repo 参数。",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoSetDefault,
}

// --- repo create ---

var repoCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "创建代码库",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("repo create %s: 待实现\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoListCmd, repoCloneCmd, repoViewCmd, repoSetDefaultCmd, repoCreateCmd)

	repoListCmd.Flags().IntVarP(&repoListPage, "page", "p", 1, "页码")
	repoListCmd.Flags().IntVarP(&repoListPerPage, "limit", "l", 20, "每页数量 (1-100)")
	repoListCmd.Flags().StringVarP(&repoListSearch, "search", "s", "", "按路径模糊搜索")
	repoCloneCmd.Flags().BoolVar(&repoCloneUseSSH, "ssh", false, "使用 SSH 地址克隆")
}

func runRepoList(cmd *cobra.Command, args []string) error {
	if repoListPage < 1 {
		return fmt.Errorf("--page 取值必须 >= 1，当前值: %d", repoListPage)
	}
	if repoListPerPage < 1 || repoListPerPage > 100 {
		return fmt.Errorf("--limit 取值范围为 1-100，当前值: %d", repoListPerPage)
	}

	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repos, err := c.ListRepositories(cfg.OrganizationID, repoListPage, repoListPerPage, repoListSearch)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("没有找到代码库。")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "名称\t路径\t可见性\t描述\n")
	fmt.Fprintf(w, "----\t----\t------\t----\n")
	for _, r := range repos {
		desc := r.Description
		if len([]rune(desc)) > 40 {
			desc = string([]rune(desc)[:37]) + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.PathWithNamespace, r.Visibility, desc)
	}
	w.Flush()

	fmt.Printf("\n显示 %d 个仓库（第 %d 页，每页 %d）\n", len(repos), repoListPage, repoListPerPage)
	return nil
}

func runRepoView(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoArg := ""
	if len(args) > 0 {
		repoArg = args[0]
	}
	repoRef, err := resolveRepoRef(cfg, repoArg)
	if err != nil {
		return err
	}

	repo, err := c.GetRepository(cfg.OrganizationID, repoRef)
	if err != nil {
		return err
	}

	fmt.Printf("名称:       %s\n", repo.Name)
	fmt.Printf("路径:       %s\n", repo.PathWithNamespace)
	fmt.Printf("默认分支:   %s\n", repo.DefaultBranch)
	fmt.Printf("可见性:     %s\n", repo.Visibility)
	fmt.Printf("访问级别:   %d\n", repo.AccessLevel)
	fmt.Printf("允许推送:   %t\n", repo.AllowPush)
	fmt.Printf("归档:       %t\n", repo.Archived)
	fmt.Printf("收藏数:     %d\n", repo.StarCount)
	fmt.Printf("Fork 数:    %d\n", repo.ForkCount)
	if repo.Description != "" {
		fmt.Printf("描述:       %s\n", repo.Description)
	}
	fmt.Printf("Web URL:    %s\n", repo.WebURL)
	fmt.Printf("HTTP URL:   %s\n", repo.HTTPURLToRepo)
	fmt.Printf("SSH URL:    %s\n", repo.SSHURLToRepo)
	fmt.Printf("创建时间:   %s\n", repo.CreatedAt)
	fmt.Printf("更新时间:   %s\n", repo.UpdatedAt)
	fmt.Printf("最近活跃:   %s\n", repo.LastActivityAt)
	return nil
}

func runRepoSetDefault(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repo, err := c.GetRepository(cfg.OrganizationID, args[0])
	if err != nil {
		return err
	}

	cfg.DefaultRepo = repo.PathWithNamespace
	if err := cfg.Save(GlobalCfgFile); err != nil {
		return err
	}

	fmt.Printf("默认仓库已设置为: %s\n", cfg.DefaultRepo)
	return nil
}

func runRepoClone(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repo, err := c.GetRepository(cfg.OrganizationID, args[0])
	if err != nil {
		return err
	}

	cloneURL := repo.HTTPURLToRepo
	if repoCloneUseSSH {
		cloneURL = repo.SSHURLToRepo
	}
	if cloneURL == "" {
		return fmt.Errorf("仓库未返回可用的克隆地址")
	}

	gitCmd := exec.Command("git", "clone", cloneURL)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	gitCmd.Stdin = os.Stdin

	if GlobalDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] git clone %s\n", cloneURL)
	}

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("执行 git clone 失败: %w；如为私有仓库，请确认本机已配置 Git 凭证，或改用 --ssh", err)
	}
	return nil
}
