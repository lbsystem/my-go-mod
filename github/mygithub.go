package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GitHubFile struct {
	Token     string
	Username  string
	RepoName  string
	ApiVersion string
}

func NewGitHubFile(token, username, repoName string) *GitHubFile {
	return &GitHubFile{
		Token:     token,
		Username:  username,
		RepoName:  repoName,
		ApiVersion: "2022-11-28",
	}
}

func (g *GitHubFile) Read(fileName string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", g.Username, g.RepoName, fileName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.VERSION.raw") // 这里添加.raw自定义媒体类型
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("X-GitHub-Api-Version", g.ApiVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}


func (g *GitHubFile) GetSha(fileName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", g.Username, g.RepoName, fileName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var content struct {
		Sha string `json:"sha"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return "", err
	}

	return content.Sha, nil
}

func (g *GitHubFile) CreateOrUpdate(fileName string, content []byte, message, mode string) error {
	sha := ""
	if mode != "create" {
		var err error
		sha, err = g.GetSha(fileName)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", g.Username, g.RepoName, fileName)
	encodedContent := base64.StdEncoding.EncodeToString(content)

	data := map[string]interface{}{
		"message": message,
		"content": encodedContent,
	}

	if sha != "" {
		data["sha"] = sha
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("创建或更新文件失败，状态码: %d", resp.StatusCode)
	}

	return nil
}


func (g *GitHubFile) Delete(fileName, message string) error {
	sha, err := g.GetSha(fileName)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", g.Username, g.RepoName, fileName)
	data := map[string]interface{}{
		"message": message,
		"sha":     sha,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("删除文件失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
