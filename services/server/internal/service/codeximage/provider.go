// Package codeximage adapts native Codex image events to workbench media assets.
package codeximage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/mediago-dev/mediago-drama/packages/core/pkg/generation"
	"github.com/mediago-dev/mediago-drama/services/server/internal/platform/codexapp"
)

// Provider starts a bounded, isolated native Codex turn using the managed account.
type Provider struct {
	StartSession func(context.Context) (codexapp.Client, error)
}

// Name identifies the subscription route, not a third-party credential.
func (*Provider) Name() string { return core.ProviderCodexImage }

// Get is unused because the normal generation workflow persists synchronous results.
func (*Provider) Get(context.Context, string) (core.Response, error) {
	return core.Response{}, errors.New("Codex 生图不支持远程任务轮询")
}

type imageItem struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Result string `json:"result"`
}

const codexImageReferenceMaxBytes = 32 * 1024 * 1024

func codexImageTurnInput(cwd string, request core.Request) ([]any, error) {
	inputs := make([]any, 0, len(request.ReferenceURLs)+1)
	for index, reference := range request.ReferenceURLs {
		reference = strings.TrimSpace(reference)
		if reference == "" {
			continue
		}
		input, err := codexImageReferenceInput(cwd, index, reference)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	instruction := "Generate exactly one image using image_generation."
	if len(inputs) > 0 {
		instruction += " The supplied reference images are ordered from highest to lowest consistency priority. Preserve identity and continuity from earlier references when later references conflict."
	}
	inputs = append(inputs, map[string]any{
		"type":          "text",
		"text":          instruction + "\n\n" + request.Prompt,
		"text_elements": []any{},
	})
	return inputs, nil
}

func codexImageReferenceInput(cwd string, index int, reference string) (map[string]any, error) {
	if strings.HasPrefix(reference, "data:") {
		header, payload, ok := strings.Cut(reference, ",")
		if !ok || !strings.Contains(strings.ToLower(header), ";base64") {
			return nil, errors.New("Codex 参考图 data URI 必须使用 base64")
		}
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil || len(decoded) == 0 {
			return nil, errors.New("Codex 参考图 data URI 无法解码")
		}
		if len(decoded) > codexImageReferenceMaxBytes {
			return nil, errors.New("Codex 单张参考图超过 32MB 限制")
		}
		mimeType := http.DetectContentType(decoded)
		extension := ""
		switch mimeType {
		case "image/png":
			extension = ".png"
		case "image/jpeg":
			extension = ".jpg"
		case "image/webp":
			extension = ".webp"
		default:
			return nil, fmt.Errorf("Codex 参考图格式不支持: %s", mimeType)
		}
		path := filepath.Join(cwd, fmt.Sprintf("reference-%02d%s", index+1, extension))
		if err := os.WriteFile(path, decoded, 0o600); err != nil {
			return nil, fmt.Errorf("写入 Codex 参考图: %w", err)
		}
		return map[string]any{"type": "localImage", "path": path, "detail": "high"}, nil
	}
	if strings.HasPrefix(reference, "https://") || strings.HasPrefix(reference, "http://") {
		return map[string]any{"type": "image", "url": reference, "detail": "high"}, nil
	}
	return nil, errors.New("Codex 参考图只支持 data URI 或 http(s) URL")
}

// Generate returns real image bytes only; a textual claim of success is an error.
func (provider *Provider) Generate(ctx context.Context, request core.Request) (core.Response, error) {
	if strings.TrimSpace(request.Prompt) == "" {
		return core.Response{}, errors.New("缺少生图提示词")
	}
	if err := core.ValidateRequestForRoute(request, core.CodexImageRoute()); err != nil {
		return core.Response{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	session, err := provider.StartSession(ctx)
	if err != nil {
		return core.Response{}, err
	}
	defer session.Close()
	available, err := codexapp.ImageAvailable(ctx, session)
	if err != nil {
		return core.Response{}, err
	}
	if !available {
		return core.Response{}, errors.New("请在 Codex 接入登录支持生图的 ChatGPT 账户")
	}
	// No project files are placed in this image-only session.
	cwd, err := os.MkdirTemp("", "jw-codex-image-")
	if err != nil {
		return core.Response{}, err
	}
	defer os.RemoveAll(cwd)
	var configuration struct {
		Config struct {
			MCPServers map[string]json.RawMessage `json:"mcp_servers"`
		} `json:"config"`
	}
	if err := session.Call(ctx, "config/read", map[string]any{"includeLayers": false, "cwd": cwd}, &configuration); err != nil {
		return core.Response{}, fmt.Errorf("检查图片会话工具隔离: %w", err)
	}
	overrides := map[string]any{"web_search": "disabled", "features.shell_tool": false, "features.apps": false, "features.plugins": false, "features.image_generation": true}
	for name := range configuration.Config.MCPServers {
		overrides["mcp_servers."+name+".enabled"] = false
	}
	// Resolve an official account model instead of inheriting a third-party text model.
	var models struct {
		Data []struct {
			Model     string `json:"model"`
			IsDefault bool   `json:"isDefault"`
			Hidden    bool   `json:"hidden"`
		} `json:"data"`
	}
	if err := session.Call(ctx, "model/list", struct{}{}, &models); err != nil {
		return core.Response{}, fmt.Errorf("列出 Codex 官方模型: %w", err)
	}
	model := ""
	for _, candidate := range models.Data {
		if !candidate.Hidden && candidate.Model != "" {
			if model == "" || candidate.IsDefault {
				model = candidate.Model
			}
			if candidate.IsDefault {
				break
			}
		}
	}
	if model == "" {
		return core.Response{}, errors.New("Codex 账户没有可用的官方模型")
	}
	var started struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := session.Call(ctx, "thread/start", map[string]any{
		"modelProvider": "openai", "model": model, "cwd": cwd, "approvalPolicy": "never", "sandbox": "read-only", "ephemeral": true,
		"config": overrides, "baseInstructions": "You are an image-generation worker. Use only the image_generation tool to create exactly one image from the user's prompt. If reference images are supplied, use them for identity, appearance, scene, and short-term continuity in the exact order provided; earlier references have higher consistency priority. Do not use shell, files, web search, MCP, or other tools. Do not provide a textual substitute for an image.",
	}, &started); err != nil {
		return core.Response{}, fmt.Errorf("启动 Codex 图片线程: %w", err)
	}
	if started.Thread.ID == "" {
		return core.Response{}, errors.New("Codex 未返回图片会话 ID")
	}
	var turn struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	turnInput, err := codexImageTurnInput(cwd, request)
	if err != nil {
		return core.Response{}, err
	}
	if err := session.Call(ctx, "turn/start", map[string]any{"threadId": started.Thread.ID, "input": turnInput}, &turn); err != nil {
		return core.Response{}, fmt.Errorf("提交 Codex 图片任务: %w", err)
	}
	if turn.Turn.ID == "" {
		return core.Response{}, errors.New("Codex 未返回图片任务 ID")
	}
	assets := []core.Asset{}
	seen := map[string]bool{}
	add := func(item imageItem) error {
		if item.Type != "imageGeneration" || seen[item.ID] {
			return nil
		}
		if item.Status != "completed" {
			return errors.New("Codex 图片工具执行未完成")
		}
		raw := item.Result
		if strings.HasPrefix(raw, "data:") {
			_, raw, _ = strings.Cut(raw, ",")
		}
		if len(raw) > 64*1024*1024 {
			return errors.New("Codex 图片结果超过大小限制")
		}
		bytes, err := base64.StdEncoding.DecodeString(raw)
		if err != nil || len(bytes) == 0 {
			return errors.New("Codex 未返回可保存的图片数据")
		}
		mime := http.DetectContentType(bytes)
		if mime != "image/png" && mime != "image/jpeg" && mime != "image/webp" {
			return errors.New("Codex 返回了不支持的图片格式")
		}
		assets = append(assets, core.Asset{Kind: core.KindImage, Base64: raw, MIMEType: mime})
		seen[item.ID] = true
		return nil
	}
	for {
		message, err := session.Next(ctx)
		if err != nil {
			return core.Response{}, fmt.Errorf("读取 Codex 图片结果: %w", err)
		}
		if message.Method != "" && len(message.ID) > 0 {
			return core.Response{}, errors.New("图片会话请求了额外工具或权限，已停止")
		}
		switch message.Method {
		case "item/completed":
			var event struct {
				ThreadID string    `json:"threadId"`
				TurnID   string    `json:"turnId"`
				Item     imageItem `json:"item"`
			}
			if json.Unmarshal(message.Params, &event) != nil || event.ThreadID != started.Thread.ID || event.TurnID != turn.Turn.ID {
				continue
			}
			if err := add(event.Item); err != nil {
				return core.Response{}, err
			}
		case "turn/completed":
			var event struct {
				ThreadID string `json:"threadId"`
				Turn     struct {
					ID     string      `json:"id"`
					Status string      `json:"status"`
					Items  []imageItem `json:"items"`
				} `json:"turn"`
			}
			if json.Unmarshal(message.Params, &event) != nil || event.ThreadID != started.Thread.ID || event.Turn.ID != turn.Turn.ID {
				continue
			}
			if event.Turn.Status != "completed" {
				return core.Response{}, errors.New("Codex 生图未完成；请检查订阅额度或稍后重试")
			}
			for _, item := range event.Turn.Items {
				if err := add(item); err != nil {
					return core.Response{}, err
				}
			}
			if len(assets) == 0 {
				return core.Response{}, errors.New("Codex 本轮没有返回图片，文字回复不算生图成功")
			}
			return core.Response{Status: "completed", Model: "image_generation", Assets: assets}, nil
		}
	}
}
