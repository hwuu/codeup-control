package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/hwuu/codeup-control/internal/client"
	"github.com/spf13/cobra"
)

var (
	branchRepo        string
	branchListPage    int
	branchListPerPage int
	branchListSearch  string
	branchCreateFrom  string
)

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "管理代码库分支",
	Long:  "列出、创建、删除云效 Codeup 代码库分支。",
}

var branchListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出分支",
	RunE:  runBranchList,
}

var branchCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "创建分支",
	Args:  cobra.ExactArgs(1),
	RunE:  runBranchCreate,
}

var branchDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "删除分支",
	Args:  cobra.ExactArgs(1),
	RunE:  runBranchDelete,
}

func init() {
	rootCmd.AddCommand(branchCmd)
	branchCmd.AddCommand(branchListCmd, branchCreateCmd, branchDeleteCmd)

	branchCmd.PersistentFlags().StringVarP(&branchRepo, "repo", "R", "", "指定仓库，格式为 <org>/<repo>")

	branchListCmd.Flags().IntVarP(&branchListPage, "page", "p", 1, "页码")
	branchListCmd.Flags().IntVarP(&branchListPerPage, "limit", "l", 20, "每页数量 (1-100)")
	branchListCmd.Flags().StringVarP(&branchListSearch, "search", "s", "", "按名称搜索")

	branchCreateCmd.Flags().StringVar(&branchCreateFrom, "from", "", "来源分支、tag 或 commit SHA")
}

func runBranchList(cmd *cobra.Command, args []string) error {
	if branchListPage < 1 {
		return fmt.Errorf("--page 取值必须 >= 1，当前值: %d", branchListPage)
	}
	if branchListPerPage < 1 || branchListPerPage > 100 {
		return fmt.Errorf("--limit 取值范围为 1-100，当前值: %d", branchListPerPage)
	}

	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, branchRepo)
	if err != nil {
		return err
	}

	branches, err := c.ListBranches(cfg.OrganizationID, repoRef, branchListPage, branchListPerPage, branchListSearch)
	if err != nil {
		return err
	}

	if len(branches) == 0 {
		fmt.Println("没有找到分支。")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "分支名\t保护\t最新提交\t提交信息\n")
	fmt.Fprintf(w, "------\t----\t--------\t--------\n")
	for _, b := range branches {
		protected := ""
		if b.Protected {
			protected = "是"
		}
		shortID := b.Commit.ShortID
		if shortID == "" && len(b.Commit.ID) >= 8 {
			shortID = b.Commit.ID[:8]
		}
		title := b.Commit.Title
		if len([]rune(title)) > 50 {
			title = string([]rune(title)[:47]) + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.Name, protected, shortID, title)
	}
	w.Flush()
	fmt.Printf("\n显示 %d 个分支（第 %d 页，每页 %d）\n", len(branches), branchListPage, branchListPerPage)
	return nil
}

func runBranchCreate(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, branchRepo)
	if err != nil {
		return err
	}

	ref := strings.TrimSpace(branchCreateFrom)
	if ref == "" {
		repo, err := c.GetRepository(cfg.OrganizationID, repoRef)
		if err != nil {
			return err
		}
		ref = repo.DefaultBranch
		if ref == "" {
			return fmt.Errorf("无法确定来源分支，请通过 --from 指定")
		}
	}

	branch, err := c.CreateBranch(cfg.OrganizationID, repoRef, client.CreateBranchInput{
		BranchName: args[0],
		Ref:        ref,
	})
	if err != nil {
		return err
	}

	fmt.Printf("已创建分支: %s\n", branch.Name)
	if branch.Commit.ID != "" {
		shortID := branch.Commit.ShortID
		if shortID == "" && len(branch.Commit.ID) >= 8 {
			shortID = branch.Commit.ID[:8]
		}
		fmt.Printf("最新提交: %s %s\n", shortID, branch.Commit.Title)
	}
	return nil
}

func runBranchDelete(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, branchRepo)
	if err != nil {
		return err
	}

	if err := c.DeleteBranch(cfg.OrganizationID, repoRef, args[0]); err != nil {
		return err
	}

	fmt.Printf("已删除分支: %s\n", args[0])
	return nil
}
