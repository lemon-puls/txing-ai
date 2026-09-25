package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	mytool "txing-ai/internal/tool"
	"txing-ai/internal/utils"
)

// websiteFetchRequest website_fetch_tool 请求参数
type websiteFetchRequest struct {
	URL string `json:"url" binding:"required"`
}

// websiteFetchResult website_fetch_tool 返回结果
// 仅包含抓取到的原始事实，不含任何 LLM 生成的描述性内容
type websiteFetchResult struct {
	// 信息来源："github" | "web"
	Source string `json:"source"`
	// GitHub 仓库名
	Name string `json:"name,omitempty"`
	// 网页标题（web 来源）
	Title string `json:"title,omitempty"`
	// 描述（GitHub repo description / 网页 meta description）
	Description string `json:"description,omitempty"`
	// 仓库主页地址（GitHub 项目配置的 homepage，通常是官网）
	HomepageURL string `json:"homepageUrl,omitempty"`
	// 仓库地址（GitHub 场景）
	RepoURL string `json:"repoUrl,omitempty"`
	// GitHub topics（英文短词，供 LLM 参考生成中文标签）
	Topics []string `json:"topics,omitempty"`
	// 头像候选：已上传 COS 的预签名 URL
	AvatarCandidate string `json:"avatarCandidate,omitempty"`
	// 网页正文摘录（web 来源，截断）
	ContentExcerpt string `json:"contentExcerpt,omitempty"`
	// 抓取过程中的非致命错误提示（如 favicon 获取失败）
	Error string `json:"error,omitempty"`
}

// githubRepo GitHub API /repos/{owner}/{repo} 响应中用到的字段
type githubRepo struct {
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Homepage    string   `json:"homepage"`
	Topics      []string `json:"topics"`
	Owner       struct {
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
	Message string `json:"message"` // API 错误时非空，如 "Not Found"、"API rate limit exceeded"
}

// fetchWebsite 抓取网站/GitHub 仓库的原始信息（只读，不写 websites 表）
func (d OpsToolDeps) fetchWebsite(ctx context.Context, req *websiteFetchRequest) (websiteFetchResult, error) {
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		return websiteFetchResult{}, fmt.Errorf("url 不能为空")
	}
	// 容错：缺 scheme 时补 https://
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return websiteFetchResult{}, fmt.Errorf("无效的 URL: %s", req.URL)
	}

	result := websiteFetchResult{}

	// GitHub 仓库：优先走 GitHub API
	if owner, repo, ok := parseGitHubRepoPath(parsed); ok {
		result.Source = "github"
		repoInfo, apiErr := d.fetchGitHubRepo(ctx, owner, repo)
		if apiErr != nil {
			// API 失败（限额/网络等）降级为网页抓取
			log.Warn("GitHub API 调用失败，降级为网页抓取",
				zap.String("url", rawURL), zap.Error(apiErr))
			result.Error = fmt.Sprintf("GitHub API 调用失败(%v)，已降级为网页抓取", apiErr)
		} else {
			result.Name = repoInfo.Name
			result.Description = repoInfo.Description
			result.RepoURL = repoInfo.HTMLURL
			result.HomepageURL = strings.TrimSpace(repoInfo.Homepage)
			result.Topics = repoInfo.Topics
			if repoInfo.Owner.AvatarURL != "" {
				result.AvatarCandidate = d.uploadToCOS(ctx, repoInfo.Owner.AvatarURL, "avatar.png")
			}
			return result, nil
		}
	} else {
		result.Source = "web"
	}

	// 网页抓取兜底/默认路径
	scraped, scrapeErr := mytool.ScrapeWebPage(ctx, &mytool.WebScrapingRequest{
		URL:           rawURL,
		MaxTextLength: 4000,
	})
	if scrapeErr != nil {
		result.Error = strings.TrimSpace(result.Error + "；网页抓取失败: " + scrapeErr.Error())
		return result, nil
	}
	result.Title = scraped.Title
	result.Description = scraped.Description
	result.ContentExcerpt = scraped.Content
	if scraped.Error != "" {
		result.Error = strings.TrimSpace(result.Error + "；" + scraped.Error)
	}

	// 头像：提取 favicon 并上传 COS（失败不阻塞，头像可为空由人工补）
	if faviconURL, faviconErr := utils.ExtractFaviconFromHTML(rawURL); faviconErr == nil && faviconURL != "" {
		ext := strings.TrimPrefix(strings.ToLower(path.Ext(faviconURL)), ".")
		if ext == "" || len(ext) > 5 {
			ext = "ico"
		}
		result.AvatarCandidate = d.uploadToCOS(ctx, faviconURL, "favicon."+ext)
	}

	return result, nil
}

// parseGitHubRepoPath 解析 github.com/{owner}/{repo} 形式的 URL
func parseGitHubRepoPath(u *url.URL) (owner, repo string, ok bool) {
	if u.Host != "github.com" && u.Host != "www.github.com" {
		return "", "", false
	}
	segs := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segs) < 2 || segs[0] == "" || segs[1] == "" {
		return "", "", false
	}
	owner = segs[0]
	repo = strings.TrimSuffix(segs[1], ".git")
	return owner, repo, true
}

// fetchGitHubRepo 调用 GitHub API 获取仓库信息
func (d OpsToolDeps) fetchGitHubRepo(ctx context.Context, owner, repo string) (*githubRepo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "txing-ai-ops-agent")
	// ops_agent 段可能缺省或热更新后被置空，需防御性判空
	if opsCfg := global.LoadConfig().OpsAgentConfig; opsCfg != nil && opsCfg.GithubToken != "" {
		req.Header.Set("Authorization", "Bearer "+opsCfg.GithubToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repoInfo githubRepo
	if err := json.Unmarshal(body, &repoInfo); err != nil {
		return nil, fmt.Errorf("解析 GitHub 响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d: %s", resp.StatusCode, repoInfo.Message)
	}
	return &repoInfo, nil
}

// uploadToCOS 将图片上传 COS 并返回预签名 URL，失败返回空串（头像可缺省）
func (d OpsToolDeps) uploadToCOS(ctx context.Context, imageURL, fileName string) string {
	if d.COS == nil || imageURL == "" {
		return ""
	}
	currentDate := time.Now().Format("2006-01-02")
	keyPath := path.Join(strconv.FormatInt(d.UserID, 10), currentDate,
		fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName))
	key, err := d.COS.PutFromURL(ctx, keyPath, imageURL)
	if err != nil {
		log.Warn("运营助手头像上传 COS 失败", zap.String("url", imageURL), zap.Error(err))
		return ""
	}
	presigned, err := d.COS.GenerateDownloadPresignedURL(key)
	if err != nil {
		log.Warn("运营助手头像生成预签名 URL 失败", zap.String("key", key), zap.Error(err))
		return ""
	}
	return presigned
}
