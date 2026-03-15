package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/hwuu/codeup-control/internal/client"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "管理合并请求",
	Long:  "创建、查看、评审、合并云效 Codeup 合并请求。",
}

var (
	prRepo        string
	prListPage    int
	prListPerPage int
	prListState   string
	prListSearch  string
	prCreateTitle string
	prCreateBody  string
	prCreateBase  string
	prCreateHead  string
)

// P0

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出合并请求",
	RunE:  runPRList,
}

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建合并请求",
	RunE:  runPRCreate,
}

var prViewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "查看合并请求详情",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRView,
}

var prMergeCmd = &cobra.Command{
	Use:   "merge [<number>]",
	Short: "执行合并",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pr merge: 待实现")
		return nil
	},
}

var prCheckoutCmd = &cobra.Command{
	Use:   "checkout <number>",
	Short: "切换到合并请求对应的分支",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pr checkout: 待实现")
		return nil
	},
}

var prDiffCmd = &cobra.Command{
	Use:   "diff [<number>]",
	Short: "查看合并请求的代码变更",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pr diff: 待实现")
		return nil
	},
}

var prCloseCmd = &cobra.Command{
	Use:   "close <number>",
	Short: "关闭合并请求",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRClose,
}

var prStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示当前分支关联的合并请求状态",
	RunE:  runPRStatus,
}

// P1

var prReviewCmd = &cobra.Command{
	Use:   "review <number>",
	Short: "提交评审（通过或拒绝）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pr review: 待实现")
		return nil
	},
}

var prCommentCmd = &cobra.Command{
	Use:   "comment <number>",
	Short: "添加评论",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pr comment: 待实现")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.AddCommand(
		prListCmd, prCreateCmd, prViewCmd, prMergeCmd,
		prCheckoutCmd, prDiffCmd, prCloseCmd, prStatusCmd,
		prReviewCmd, prCommentCmd,
	)

	prCmd.PersistentFlags().StringVar(&prRepo, "repo", "", "指定仓库，格式为 <org>/<repo>")

	prListCmd.Flags().IntVarP(&prListPage, "page", "p", 1, "页码")
	prListCmd.Flags().IntVarP(&prListPerPage, "limit", "l", 20, "每页数量 (1-100)")
	prListCmd.Flags().StringVar(&prListState, "state", "opened", "状态筛选: opened, merged, closed")
	prListCmd.Flags().StringVarP(&prListSearch, "search", "s", "", "按标题搜索")

	prCreateCmd.Flags().StringVarP(&prCreateTitle, "title", "t", "", "合并请求标题")
	prCreateCmd.Flags().StringVarP(&prCreateBody, "body", "b", "", "合并请求描述")
	prCreateCmd.Flags().StringVar(&prCreateBase, "base", "", "目标分支，默认仓库默认分支")
	prCreateCmd.Flags().StringVar(&prCreateHead, "head", "", "源分支，默认当前 git 分支")
}

func runPRList(cmd *cobra.Command, args []string) error {
	if prListPage < 1 {
		return fmt.Errorf("--page 取值必须 >= 1，当前值: %d", prListPage)
	}
	if prListPerPage < 1 || prListPerPage > 100 {
		return fmt.Errorf("--limit 取值范围为 1-100，当前值: %d", prListPerPage)
	}

	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	projectID := ""
	if strings.TrimSpace(prRepo) != "" || strings.TrimSpace(cfg.DefaultRepo) != "" {
		_, projectID, err = resolveRepoProjectID(c, cfg, prRepo)
		if err != nil {
			return err
		}
	}

	prs, err := c.ListChangeRequests(cfg.OrganizationID, client.ListChangeRequestsOptions{
		Page:      prListPage,
		PerPage:   prListPerPage,
		ProjectID: projectID,
		State:     prListState,
		Search:    prListSearch,
	})
	if err != nil {
		return err
	}

	if len(prs) == 0 {
		fmt.Println("没有找到合并请求。")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "编号\t状态\t标题\t分支\t作者\n")
	fmt.Fprintf(w, "----\t----\t----\t----\t----\n")
	for _, pr := range prs {
		state := pr.State
		if state == "" {
			state = pr.Status
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s -> %s\t%s\n",
			pr.LocalID, state, pr.Title, pr.SourceBranch, pr.TargetBranch, pr.Author.Username)
	}
	w.Flush()
	fmt.Printf("\n显示 %d 个合并请求（第 %d 页，每页 %d）\n", len(prs), prListPage, prListPerPage)
	return nil
}

func runPRView(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, prRepo)
	if err != nil {
		return err
	}

	localID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("合并请求编号必须是整数: %s", args[0])
	}

	pr, err := c.GetChangeRequest(cfg.OrganizationID, repoRef, localID)
	if err != nil {
		return err
	}

	state := pr.State
	if state == "" {
		state = pr.Status
	}

	fmt.Printf("编号:       %d\n", pr.LocalID)
	fmt.Printf("标题:       %s\n", pr.Title)
	fmt.Printf("状态:       %s\n", state)
	fmt.Printf("作者:       %s (%s)\n", pr.Author.Name, pr.Author.Username)
	fmt.Printf("分支:       %s -> %s\n", pr.SourceBranch, pr.TargetBranch)
	fmt.Printf("评论:       %d（未解决 %d）\n", pr.TotalCommentCount, pr.UnResolvedCommentCount)
	fmt.Printf("有冲突:     %t\n", pr.HasConflict)
	fmt.Printf("WIP:        %t\n", pr.WorkInProgress)
	if pr.Description != "" {
		fmt.Printf("描述:       %s\n", pr.Description)
	}
	if pr.WebURL != "" {
		fmt.Printf("Web URL:    %s\n", pr.WebURL)
	} else if pr.DetailURL != "" {
		fmt.Printf("详情 URL:   %s\n", pr.DetailURL)
	}
	if pr.CreatedAt != "" || pr.CreateTime != "" {
		fmt.Printf("创建时间:   %s\n", firstNonEmpty(pr.CreatedAt, pr.CreateTime))
	}
	if pr.UpdatedAt != "" || pr.UpdateTime != "" {
		fmt.Printf("更新时间:   %s\n", firstNonEmpty(pr.UpdatedAt, pr.UpdateTime))
	}
	if len(pr.Reviewers) > 0 {
		names := make([]string, 0, len(pr.Reviewers))
		for _, r := range pr.Reviewers {
			names = append(names, r.Username)
		}
		fmt.Printf("评审人:     %s\n", strings.Join(names, ", "))
	}
	return nil
}

func runPRCreate(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(prCreateTitle) == "" {
		return fmt.Errorf("请通过 --title 指定合并请求标题")
	}

	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, prRepo)
	if err != nil {
		return err
	}

	repo, err := c.GetRepository(cfg.OrganizationID, repoRef)
	if err != nil {
		return err
	}

	head := strings.TrimSpace(prCreateHead)
	if head == "" {
		head, err = currentGitBranch()
		if err != nil {
			return fmt.Errorf("无法自动识别当前分支，请通过 --head 指定: %w", err)
		}
	}

	base := strings.TrimSpace(prCreateBase)
	if base == "" {
		base = strings.TrimSpace(repo.DefaultBranch)
	}
	if base == "" {
		return fmt.Errorf("无法确定目标分支，请通过 --base 指定")
	}

	created, err := c.CreateChangeRequest(cfg.OrganizationID, repoRef, client.CreateChangeRequestInput{
		Title:           strings.TrimSpace(prCreateTitle),
		Description:     strings.TrimSpace(prCreateBody),
		SourceBranch:    head,
		SourceProjectID: repo.ID,
		TargetBranch:    base,
		TargetProjectID: repo.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("已创建合并请求 #%d: %s\n", created.LocalID, created.Title)
	if created.WebURL != "" {
		fmt.Printf("Web URL: %s\n", created.WebURL)
	} else if created.DetailURL != "" {
		fmt.Printf("详情 URL: %s\n", created.DetailURL)
	}
	return nil
}

func runPRClose(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	repoRef, err := resolveRepoRef(cfg, prRepo)
	if err != nil {
		return err
	}

	localID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("合并请求编号必须是整数: %s", args[0])
	}

	if err := c.CloseChangeRequest(cfg.OrganizationID, repoRef, localID); err != nil {
		return err
	}

	fmt.Printf("已关闭合并请求 #%d\n", localID)
	return nil
}

func runPRStatus(cmd *cobra.Command, args []string) error {
	c, cfg, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	_, projectID, err := resolveRepoProjectID(c, cfg, prRepo)
	if err != nil {
		return err
	}

	branch, err := currentGitBranch()
	if err != nil {
		return fmt.Errorf("无法识别当前 git 分支: %w", err)
	}

	prs, err := c.ListChangeRequests(cfg.OrganizationID, client.ListChangeRequestsOptions{
		Page:      1,
		PerPage:   100,
		ProjectID: projectID,
		State:     "opened",
	})
	if err != nil {
		return err
	}

	matched := make([]client.ChangeRequest, 0)
	for _, pr := range prs {
		if pr.SourceBranch == branch {
			matched = append(matched, pr)
		}
	}

	if len(matched) == 0 {
		fmt.Printf("当前分支 %q 没有关联的打开中合并请求。\n", branch)
		return nil
	}

	for _, pr := range matched {
		state := pr.State
		if state == "" {
			state = pr.Status
		}
		fmt.Printf("#%d  %s\n", pr.LocalID, pr.Title)
		fmt.Printf("状态: %s\n", state)
		fmt.Printf("分支: %s -> %s\n", pr.SourceBranch, pr.TargetBranch)
		if pr.WebURL != "" {
			fmt.Printf("URL:  %s\n", pr.WebURL)
		} else if pr.DetailURL != "" {
			fmt.Printf("URL:  %s\n", pr.DetailURL)
		}
		fmt.Println()
	}
	return nil
}

func currentGitBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
