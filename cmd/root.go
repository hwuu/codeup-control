package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "cuctl",
	Short:   "云效 Codeup 命令行工具，仿 gh 操作仓库与合并请求",
	Version: version,
	Long: `cuctl 是 Codeup Control 的命令行入口，用于在终端中操作阿里云效 Codeup，
支持仓库列表、克隆与合并请求（PR）等操作。
认证当前使用个人访问令牌（PAT），后续可扩展 OAuth。`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var GlobalCfgFile string
var GlobalDebug bool

func init() {
	rootCmd.PersistentFlags().StringVar(&GlobalCfgFile, "config", "", "配置文件路径（默认 ~/.config/cuctl/config.yaml）")
	rootCmd.PersistentFlags().BoolVar(&GlobalDebug, "debug", false, "输出调试日志")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
