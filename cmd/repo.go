package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/hwuu/codeup-control/internal/client"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "管理云效代码库",
	Long:  "列出、克隆、查看、创建、编辑云效 Codeup 代码库。",
}

var (
	repoListPage    int
	repoListPerPage int
	repoListSearch  string
	repoCloneUseSSH bool
)

var (
	repoCreateDesc       string
	repoCreateVisibility string
	repoCreateNamespace  string
	repoCreateReadme     bool
)

var (
	repoEditDesc       string
	repoEditVisibility string
	repoEditBranch     string
)

var (
	repoDeleteYes  bool
	repoRenameRepo string
)

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出有权限的代码库",
	RunE:  runRepoList,
}

var repoCloneCmd = &cobra.Command{
	Use:   "clone [<org>/]<repo>",
	Short: "克隆代码库",
	Long:  "克隆指定代码库。使用 HTTPS 时仍依赖本机已有的 Git 凭证配置；如未配置，建议使用 --ssh。",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoClone,
}

var repoViewCmd = &cobra.Command{
	Use:   "view [<org>/]<repo>",
	Short: "查看代码库概要",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoView,
}

var repoSetDefaultCmd = &cobra.Command{
	Use:   "set-default [<org>/]<repo>",
	Short: "设置当前默认仓库",
	Long:  "设置默认仓库后，后续命令可省略 org/repo 参数。",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoSetDefault,
}

var repoCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "创建代码库",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoCreate,
}

var repoEditCmd = &cobra.Command{
	Use:   "edit [<org>/]<repo>",
	Short: "编辑代码库设置",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoEdit,
}

var repoDeleteCmd = &cobra.Command{
	Use:   "delete [<org>/]<repo>",
	Short: "删除代码库",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoDelete,
}

var repoForkCmd = &cobra.Command{
	Use:   "fork [<org>/]<repo>",
	Short: "Fork 代码库",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoFork,
}

var repoArchiveCmd = &cobra.Command{
	Use:   "archive [<org>/]<repo>",
	Short: "归档代码库",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoArchive,
}

var repoUnarchiveCmd = &cobra.Command{
	Use:   "unarchive [<org>/]<repo>",
	Short: "取消归档代码库",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoUnarchive,
}

var repoRenameCmd = &cobra.Command{
	Use:   "rename <new-name>",
	Short: "重命名代码库",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoRename,
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(
		repoListCmd, repoCloneCmd, repoViewCmd, repoSetDefaultCmd, repoCreateCmd,
		repoEditCmd, repoDeleteCmd, repoForkCmd, repoArchiveCmd, repoUnarchiveCmd,
		repoRenameCmd,
	)

	repoListCmd.Flags().IntVarP(&repoListPage, "page", "p", 1, "页码")
	repoListCmd.Flags().IntVarP(&repoListPerPage, "limit", "l", 20, "每页数量 (1-100)")
	repoListCmd.Flags().StringVarP(&repoListSearch, "search", "s", "", "按路径模糊搜索")

	repoCloneCmd.Flags().BoolVar(&repoCloneUseSSH, "ssh", false, "使用 SSH 地址克隆")

	repoCreateCmd.Flags().StringVarP(&repoCreateDesc, "description", "d", "", "仓库描述")
	repoCreateCmd.Flags().StringVar(&repoCreateVisibility, "visibility", "", "可见性: public 或 private")
	repoCreateCmd.Flags().StringVar(&repoCreateNamespace, "namespace", "", "仓库命名空间/分组路径")
	repoCreateCmd.Flags().BoolVar(&repoCreateReadme, "init-readme", false, "初始化 README 文件")

	repoEditCmd.Flags().StringVarP(&repoEditDesc, "description", "d", "", "仓库描述")
	repoEditCmd.Flags().StringVar(&repoEditVisibility, "visibility", "", "可见性: public 或 private")
	repoEditCmd.Flags().StringVar(&repoEditBranch, "default-branch", "", "默认分支")

	repoDeleteCmd.Flags().BoolVar(&repoDeleteYes, "yes", false, "跳过确认提示")

	repoRenameCmd.Flags().StringVarP(&repoRenameRepo, "repo", "R", "", "指定仓库，格式为 <org>/<repo>")
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

func runRepoCreate(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repo, err := c.CreateRepository(cfg.OrganizationID, client.CreateRepositoryInput{
		Name:          args[0],
		NamespacePath: strings.TrimSpace(repoCreateNamespace),
		Description:   strings.TrimSpace(repoCreateDesc),
		Visibility:    strings.TrimSpace(repoCreateVisibility),
		InitReadme:    repoCreateReadme,
	})
	if err != nil {
		return err
	}

	fmt.Printf("已创建仓库: %s\n", repo.PathWithNamespace)
	fmt.Printf("Web URL:  %s\n", repo.WebURL)
	fmt.Printf("HTTP URL: %s\n", repo.HTTPURLToRepo)
	fmt.Printf("SSH URL:  %s\n", repo.SSHURLToRepo)
	return nil
}

func runRepoEdit(cmd *cobra.Command, args []string) error {
	if repoEditDesc == "" && repoEditVisibility == "" && repoEditBranch == "" {
		return fmt.Errorf("请至少指定 --description、--visibility 或 --default-branch")
	}

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

	input := client.UpdateRepositoryInput{}
	if repoEditDesc != "" {
		input.Description = repoEditDesc
	}
	if repoEditVisibility != "" {
		input.Visibility = repoEditVisibility
	}
	if repoEditBranch != "" {
		input.DefaultBranch = repoEditBranch
	}

	repo, err := c.UpdateRepository(cfg.OrganizationID, repoRef, input)
	if err != nil {
		return err
	}

	fmt.Printf("已更新仓库: %s\n", repo.PathWithNamespace)
	return nil
}

func runRepoDelete(cmd *cobra.Command, args []string) error {
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

	if !repoDeleteYes {
		parts := strings.Split(repoRef, "/")
		expectedName := parts[len(parts)-1]
		fmt.Printf("确定要删除仓库 %q 吗？此操作不可撤销。\n输入仓库名称 %q 确认: ", repoRef, expectedName)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(answer) != expectedName {
			fmt.Println("输入不匹配，已取消。")
			return nil
		}
	}

	if err := c.DeleteRepository(cfg.OrganizationID, repoRef); err != nil {
		return err
	}

	fmt.Printf("已删除仓库: %s\n", repoRef)
	return nil
}

func runRepoFork(cmd *cobra.Command, args []string) error {
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

	repo, err := c.ForkRepository(cfg.OrganizationID, repoRef)
	if err != nil {
		return err
	}

	fmt.Printf("已 fork 仓库: %s\n", repo.PathWithNamespace)
	if repo.WebURL != "" {
		fmt.Printf("Web URL: %s\n", repo.WebURL)
	}
	return nil
}

func runRepoArchive(cmd *cobra.Command, args []string) error {
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

	if err := c.ArchiveRepository(cfg.OrganizationID, repoRef); err != nil {
		return err
	}

	fmt.Printf("已归档仓库: %s\n", repoRef)
	return nil
}

func runRepoUnarchive(cmd *cobra.Command, args []string) error {
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

	if err := c.UnarchiveRepository(cfg.OrganizationID, repoRef); err != nil {
		return err
	}

	fmt.Printf("已取消归档仓库: %s\n", repoRef)
	return nil
}

func runRepoRename(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, repoRenameRepo)
	if err != nil {
		return err
	}

	repo, err := c.UpdateRepository(cfg.OrganizationID, repoRef, client.UpdateRepositoryInput{
		Name: args[0],
	})
	if err != nil {
		return err
	}

	fmt.Printf("已重命名仓库: %s\n", repo.PathWithNamespace)
	return nil
}
