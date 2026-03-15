package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hwuu/codeup-control/internal/client"
	"github.com/hwuu/codeup-control/internal/config"
)

func loadClientFromConfig() (*client.Client, *config.Config, error) {
	cfg, err := config.Load(GlobalCfgFile)
	if err != nil {
		return nil, nil, err
	}

	token, _, err := config.ResolveToken(GlobalCfgFile)
	if err != nil {
		return nil, nil, err
	}
	if token == "" {
		return nil, nil, fmt.Errorf("未认证，请先运行: cuctl auth login")
	}

	orgID, _ := cfg.ResolveOrganizationID()
	if orgID == "" {
		return nil, nil, fmt.Errorf("未配置组织 ID，请先运行: cuctl auth login，或设置环境变量 CODEUP_ORGANIZATION_ID")
	}
	cfg.OrganizationID = orgID

	c := client.New(cfg.ResolveDomain(), token, GlobalDebug)
	return c, cfg, nil
}

func resolveRepoRef(cfg *config.Config, arg string) (string, error) {
	if strings.TrimSpace(arg) != "" {
		return strings.TrimSpace(arg), nil
	}
	if strings.TrimSpace(cfg.DefaultRepo) != "" {
		return strings.TrimSpace(cfg.DefaultRepo), nil
	}
	return "", fmt.Errorf("未指定仓库，且未设置默认仓库，请使用: cuctl repo set-default <org/repo>")
}

func resolveRepoProjectID(c *client.Client, cfg *config.Config, arg string) (repoRef string, projectID string, err error) {
	repoRef, err = resolveRepoRef(cfg, arg)
	if err != nil {
		return "", "", err
	}

	repo, err := c.GetRepository(cfg.OrganizationID, repoRef)
	if err != nil {
		return "", "", err
	}

	return repoRef, strconv.Itoa(repo.ID), nil
}
