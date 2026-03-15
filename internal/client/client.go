package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Domain string
	Token  string
	HTTP   *http.Client
	Debug  bool
}

func New(domain, token string, debug bool) *Client {
	return &Client{
		Domain: domain,
		Token:  token,
		HTTP:   &http.Client{Timeout: 30 * time.Second},
		Debug:  debug,
	}
}

func (c *Client) baseURL() string {
	return fmt.Sprintf("https://%s/oapi/v1", c.Domain)
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API 请求失败 (HTTP %d): %s", e.StatusCode, e.Body)
}

func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, error) {
	reqURL := c.baseURL() + path
	if c.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, reqURL)
	}

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-yunxiao-token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: truncate(string(respBody), 500)}
	}
	return respBody, nil
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

// --- User ---

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func (c *Client) GetCurrentUser() (*User, error) {
	data, err := c.doRequest("GET", "/platform/user", nil)
	if err != nil {
		return nil, err
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}
	return &u, nil
}

// --- Repository ---

type AccessLevel int

func (a *AccessLevel) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = 0
		return nil
	}

	var numeric int
	if err := json.Unmarshal(data, &numeric); err == nil {
		*a = AccessLevel(numeric)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return fmt.Errorf("解析 accessLevel 失败: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		*a = 0
		return nil
	}

	numeric, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("解析 accessLevel 失败: %w", err)
	}
	*a = AccessLevel(numeric)
	return nil
}

type Repository struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"pathWithNamespace"`
	Description       string `json:"description"`
	Visibility        string `json:"visibility"`
	WebURL            string `json:"webUrl"`
	Archived          bool   `json:"archived"`
	CreatedAt         string `json:"createdAt"`
	LastActivityAt    string `json:"lastActivityAt"`
	StarCount         int    `json:"starCount"`
	AccessLevel       AccessLevel `json:"accessLevel"`
}

type RepositoryDetail struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"pathWithNamespace"`
	NameWithNamespace string `json:"nameWithNamespace"`
	Description       string `json:"description"`
	DefaultBranch     string `json:"defaultBranch"`
	Visibility        string `json:"visibility"`
	WebURL            string `json:"webUrl"`
	HTTPURLToRepo     string `json:"httpUrlToRepo"`
	SSHURLToRepo      string `json:"sshUrlToRepo"`
	Archived          bool   `json:"archived"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	LastActivityAt    string `json:"lastActivityAt"`
	StarCount         int    `json:"starCount"`
	ForkCount         int    `json:"forkCount"`
	AccessLevel       AccessLevel `json:"accessLevel"`
	AllowPush         bool   `json:"allowPush"`
}

type ChangeRequestAuthor struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   string `json:"userId"`
}

type ChangeRequestReviewer struct {
	Name                string `json:"name"`
	Username            string `json:"username"`
	UserID              string `json:"userId"`
	HasReviewed         bool   `json:"hasReviewed"`
	HasCommented        bool   `json:"hasCommented"`
	ReviewOpinionStatus string `json:"reviewOpinionStatus"`
	ReviewTime          string `json:"reviewTime"`
}

type ChangeRequest struct {
	LocalID                int                     `json:"localId"`
	Title                  string                  `json:"title"`
	Description            string                  `json:"description"`
	State                  string                  `json:"state"`
	Status                 string                  `json:"status"`
	SourceBranch           string                  `json:"sourceBranch"`
	TargetBranch           string                  `json:"targetBranch"`
	SourceProjectID        int                     `json:"sourceProjectId"`
	TargetProjectID        int                     `json:"targetProjectId"`
	ProjectID              int                     `json:"projectId"`
	WebURL                 string                  `json:"webUrl"`
	DetailURL              string                  `json:"detailUrl"`
	CreatedAt              string                  `json:"createdAt"`
	UpdatedAt              string                  `json:"updatedAt"`
	CreateTime             string                  `json:"createTime"`
	UpdateTime             string                  `json:"updateTime"`
	Author                 ChangeRequestAuthor     `json:"author"`
	Reviewers              []ChangeRequestReviewer `json:"reviewers"`
	HasConflict            bool                    `json:"hasConflict"`
	WorkInProgress         bool                    `json:"workInProgress"`
	TotalCommentCount      int                     `json:"totalCommentCount"`
	UnResolvedCommentCount int                     `json:"unResolvedCommentCount"`
}

type ListChangeRequestsOptions struct {
	Page      int
	PerPage   int
	ProjectID string
	State     string
	Search    string
}

type CreateChangeRequestInput struct {
	Title           string   `json:"title"`
	Description     string   `json:"description,omitempty"`
	SourceBranch    string   `json:"sourceBranch"`
	SourceProjectID int      `json:"sourceProjectId"`
	TargetBranch    string   `json:"targetBranch"`
	TargetProjectID int      `json:"targetProjectId"`
	ReviewerUserIDs []string `json:"reviewerUserIds,omitempty"`
}

type boolResult struct {
	Result bool `json:"result"`
}

func (c *Client) ListRepositories(orgID string, page, perPage int, search string) ([]Repository, error) {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories?page=%d&perPage=%d",
		orgID, page, perPage)
	if search != "" {
		path += "&search=" + url.QueryEscape(search)
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var repos []Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("解析仓库列表失败: %w", err)
	}
	return repos, nil
}

func (c *Client) GetRepository(orgID, repoIDOrPath string) (*RepositoryDetail, error) {
	repoIDOrPath = strings.TrimSpace(repoIDOrPath)
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s", orgID, url.PathEscape(repoIDOrPath))

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var repo RepositoryDetail
	if err := json.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("解析仓库详情失败: %w", err)
	}
	return &repo, nil
}

func (c *Client) ListChangeRequests(orgID string, opt ListChangeRequestsOptions) ([]ChangeRequest, error) {
	values := url.Values{}
	if opt.Page > 0 {
		values.Set("page", fmt.Sprintf("%d", opt.Page))
	}
	if opt.PerPage > 0 {
		values.Set("perPage", fmt.Sprintf("%d", opt.PerPage))
	}
	if opt.ProjectID != "" {
		values.Set("projectIds", opt.ProjectID)
	}
	if opt.State != "" {
		values.Set("state", opt.State)
	}
	if opt.Search != "" {
		values.Set("search", opt.Search)
	}

	path := fmt.Sprintf("/codeup/organizations/%s/changeRequests", orgID)
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var prs []ChangeRequest
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, fmt.Errorf("解析合并请求列表失败: %w", err)
	}
	return prs, nil
}

func (c *Client) GetChangeRequest(orgID, repoIDOrPath string, localID int) (*ChangeRequest, error) {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var pr ChangeRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("解析合并请求详情失败: %w", err)
	}
	return &pr, nil
}

func (c *Client) CreateChangeRequest(orgID, repoIDOrPath string, input CreateChangeRequestInput) (*ChangeRequest, error) {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("序列化合并请求失败: %w", err)
	}

	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))

	data, err := c.doRequest("POST", path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	var pr ChangeRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("解析创建结果失败: %w", err)
	}
	return &pr, nil
}

func (c *Client) CloseChangeRequest(orgID, repoIDOrPath string, localID int) error {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d/close",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)

	data, err := c.doRequest("POST", path, nil)
	if err != nil {
		return err
	}

	var result boolResult
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("解析关闭结果失败: %w", err)
	}
	if !result.Result {
		return fmt.Errorf("关闭合并请求失败")
	}
	return nil
}
