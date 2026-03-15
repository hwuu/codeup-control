package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "管理代码库分支",
	Long:  "列出、创建、删除云效 Codeup 代码库分支。",
}

var branchListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出分支",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("branch list: 待实现")
		return nil
	},
}

var branchCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "创建分支",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("branch create %s: 待实现\n", args[0])
		return nil
	},
}

var branchDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "删除分支",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("branch delete %s: 待实现\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
	branchCmd.AddCommand(branchListCmd, branchCreateCmd, branchDeleteCmd)

	branchCreateCmd.Flags().String("from", "", "来源分支、tag 或 commit SHA")
}
