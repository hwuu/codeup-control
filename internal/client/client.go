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

// --- Update ---

type UpdateChangeRequestInput struct {
	Title          string `json:"title,omitempty"`
	Description    string `json:"description,omitempty"`
	WorkInProgress *bool  `json:"workInProgress,omitempty"`
}

func (c *Client) UpdateChangeRequest(orgID, repoIDOrPath string, localID int, input UpdateChangeRequestInput) (*ChangeRequest, error) {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)
	data, err := c.doRequest("PUT", path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	var pr ChangeRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("解析更新结果失败: %w", err)
	}
	return &pr, nil
}

// --- Merge ---

type MergeChangeRequestInput struct {
	MergeType          string `json:"mergeType,omitempty"`
	DeleteSourceBranch bool   `json:"deleteSourceBranch,omitempty"`
	Message            string `json:"message,omitempty"`
}

func (c *Client) MergeChangeRequest(orgID, repoIDOrPath string, localID int, input MergeChangeRequestInput) error {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d/merge",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)
	_, err = c.doRequest("POST", path, bytes.NewReader(reqBody))
	return err
}

// --- Review ---

type ReviewChangeRequestInput struct {
	ReviewOpinion string `json:"reviewOpinion"`
}

func (c *Client) ReviewChangeRequest(orgID, repoIDOrPath string, localID int, input ReviewChangeRequestInput) error {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d/review",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)
	_, err = c.doRequest("POST", path, bytes.NewReader(reqBody))
	return err
}

// --- Comment ---

type CommentChangeRequestInput struct {
	Content string `json:"content"`
}

func (c *Client) CommentChangeRequest(orgID, repoIDOrPath string, localID int, input CommentChangeRequestInput) error {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d/comments",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)
	_, err = c.doRequest("POST", path, bytes.NewReader(reqBody))
	return err
}

// --- Reopen ---

func (c *Client) ReopenChangeRequest(orgID, repoIDOrPath string, localID int) error {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/changeRequests/%d/reopen",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), localID)
	_, err := c.doRequest("POST", path, nil)
	return err
}

// --- Branch ---

type BranchCommit struct {
	ID        string `json:"id"`
	ShortID   string `json:"shortId"`
	Title     string `json:"title"`
	Author    string `json:"authorName"`
	CreatedAt string `json:"createdAt"`
}

type Branch struct {
	Name      string       `json:"name"`
	Protected bool         `json:"protected"`
	Commit    BranchCommit `json:"commit"`
}

type CreateBranchInput struct {
	BranchName string `json:"branchName"`
	Ref        string `json:"ref"`
}

func (c *Client) ListBranches(orgID, repoIDOrPath string, page, perPage int, search string) ([]Branch, error) {
	values := url.Values{}
	values.Set("page", fmt.Sprintf("%d", page))
	values.Set("perPage", fmt.Sprintf("%d", perPage))
	if search != "" {
		values.Set("search", search)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/branches?%s",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), values.Encode())

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var branches []Branch
	if err := json.Unmarshal(data, &branches); err != nil {
		return nil, fmt.Errorf("解析分支列表失败: %w", err)
	}
	return branches, nil
}

func (c *Client) CreateBranch(orgID, repoIDOrPath string, input CreateBranchInput) (*Branch, error) {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/branches",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))

	data, err := c.doRequest("POST", path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	var branch Branch
	if err := json.Unmarshal(data, &branch); err != nil {
		return nil, fmt.Errorf("解析分支信息失败: %w", err)
	}
	return &branch, nil
}

func (c *Client) DeleteBranch(orgID, repoIDOrPath, branchName string) error {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/branches/%s",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)), url.PathEscape(branchName))
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// --- Repository Create ---

type CreateRepositoryInput struct {
	Name          string `json:"name"`
	NamespacePath string `json:"namespacePath,omitempty"`
	Description   string `json:"description,omitempty"`
	Visibility    string `json:"visibility,omitempty"`
	InitReadme    bool   `json:"initReadme,omitempty"`
}

func (c *Client) CreateRepository(orgID string, input CreateRepositoryInput) (*RepositoryDetail, error) {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories", orgID)

	data, err := c.doRequest("POST", path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	var repo RepositoryDetail
	if err := json.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("解析创建结果失败: %w", err)
	}
	return &repo, nil
}

// --- Repository Update/Delete/Fork/Archive ---

type UpdateRepositoryInput struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	Visibility    string `json:"visibility,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
}

func (c *Client) UpdateRepository(orgID, repoIDOrPath string, input UpdateRepositoryInput) (*RepositoryDetail, error) {
	reqBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))
	data, err := c.doRequest("PUT", path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	var repo RepositoryDetail
	if err := json.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("解析更新结果失败: %w", err)
	}
	return &repo, nil
}

func (c *Client) DeleteRepository(orgID, repoIDOrPath string) error {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

func (c *Client) ForkRepository(orgID, repoIDOrPath string) (*RepositoryDetail, error) {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/fork",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))
	data, err := c.doRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}
	var repo RepositoryDetail
	if err := json.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("解析 fork 结果失败: %w", err)
	}
	return &repo, nil
}

func (c *Client) ArchiveRepository(orgID, repoIDOrPath string) error {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/archive",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))
	_, err := c.doRequest("POST", path, nil)
	return err
}

func (c *Client) UnarchiveRepository(orgID, repoIDOrPath string) error {
	path := fmt.Sprintf("/codeup/organizations/%s/repositories/%s/unarchive",
		orgID, url.PathEscape(strings.TrimSpace(repoIDOrPath)))
	_, err := c.doRequest("POST", path, nil)
	return err
}
