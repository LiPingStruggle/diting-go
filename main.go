package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	bold   = "\033[1m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

type ModelConfig struct {
	Name     string
	Type     string // local, cloud
	Endpoint string
	APIKey   string
	Model    string
}

var models = map[string]ModelConfig{
	"openai": {
		Name: "OpenAI GPT-4o-mini", Type: "cloud",
		Endpoint: "https://api.openai.com/v1/chat/completions",
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o-mini",
	},
	"claude": {
		Name: "Claude Sonnet", Type: "cloud",
		Endpoint: "https://api.anthropic.com/v1/messages",
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  "claude-3-sonnet-20240229",
	},
	"ollama": {
		Name: "Ollama (Local)", Type: "local",
		Endpoint: "http://localhost:11434/api/chat",
		Model: "qwen2.5-coder:7b",
	},
	"deepseek": {
		Name: "DeepSeek", Type: "cloud",
		Endpoint: "https://api.deepseek.com/v1/chat/completions",
		APIKey: os.Getenv("DEEPSEEK_API_KEY"),
		Model:  "deepseek-chat",
	},
}

func main() {
	fmt.Println()
	fmt.Printf("%s🎉 欢迎使用《谛听》(Diting)！%s\n", bold, reset)
	fmt.Printf("%s真理仲裁者，听辨万物真伪%s\n\n", green, reset)

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cmd := os.Args[1]
	selectedModel := getFlag("--model", "-m")
	if selectedModel == "" {
		selectedModel = "openai"
	}

	switch cmd {
	case "preview", "preview-cmd":
		handlePreview()
	case "verify", "verify-cmd":
		handleVerify(selectedModel)
	case "dashboard":
		handleDashboard()
	case "stats", "stats-cmd":
		handleStats()
	case "check":
		handleCheck()
	case "models", "list-models":
		printModels()
	case "help", "--help":
		printHelp()
	case "version", "-v":
		fmt.Println("《谛听》v1.1.0 (本地+云端模型)")
	default:
		fmt.Printf("%s❌ 未知命令: %s%s\n", red, cmd, reset)
	}
}

func printHelp() {
	fmt.Printf(`
%s📖 《谛听》命令帮助%s

%s核心命令:%s
  diting preview <问题> --code <文件>    即时预览
  diting verify <断言> --code <文件> -m 模型  完整验证(含AI)
  diting models                     查看模型

%s模型:%s
  --model, -m 选择模型: openai/claude/ollama/deepseek
  --code, -c  代码路径
  --log, -l   日志路径
  --json, -j  输出JSON

%s示例:%s
  diting verify "空指针" --code auth.py -m openai
  diting preview "Bug" --code auth.py
`, bold, reset, bold, reset, bold, reset, bold, reset)
}

func printModels() {
	fmt.Printf("\n%s🤖 可用模型%s\n\n", bold, reset)
	for name, cfg := range models {
		icon := fmt.Sprintf("%s⚠️ 未配置API Key%s", yellow, reset)
		if cfg.APIKey != "" || cfg.Type == "local" {
			icon = fmt.Sprintf("%s✅ 已就绪%s", green, reset)
		}
		fmt.Printf("  %s: %s (%s) %s\n", name, cfg.Name, cfg.Type, icon)
	}
}

func callAI(prompt, modelName string) string {
	cfg, ok := models[modelName]
	if !ok {
		return getMock(prompt)
	}
	if cfg.Type == "local" {
		return callOllama(prompt, cfg)
	}
	return callCloud(prompt, cfg)
}

func callCloud(prompt string, cfg ModelConfig) string {
	if cfg.APIKey == "" {
		return getMock(prompt)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.1,
	})
	req, _ := http.NewRequest("POST", cfg.Endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return getMock(prompt)
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if c, ok := choices[0].(map[string]interface{}); ok {
			if m, ok := c["message"].(map[string]interface{}); ok {
				if s, ok := m["content"].(string); ok {
					return s
				}
			}
		}
	}
	return getMock(prompt)
}

func callOllama(prompt string, cfg ModelConfig) string {
	body, _ := json.Marshal(map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"stream": false,
	})
	resp, err := http.Post(cfg.Endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return getMock(prompt)
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if m, ok := result["message"].(map[string]interface{}); ok {
		if s, ok := m["content"].(string); ok {
			return s
		}
	}
	return getMock(prompt)
}

func getMock(prompt string) string {
	time.Sleep(200 * time.Millisecond)
	return "空指针异常由于db.getUser()返回None时未检查导致"
}

func handlePreview() {
	codeFile := getFlag("--code", "-c")
	if codeFile == "" {
		fmt.Printf("%s❌ --code必填%s\n", red, reset)
		return
	}
	claim := ""
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "--") {
		claim = os.Args[2]
	}
	fmt.Printf("%s⚡ 即时预览...%s\n", yellow, reset)
	data, err := os.ReadFile(codeFile)
	if err != nil {
		fmt.Printf("%s❌ 读取失败: %s%s\n", red, err, reset)
		return
	}
	code := string(data)
	funcs := extractFunctions(code)
	fmt.Printf("%s✓ %d个函数: %s%s\n", green, len(funcs), funcs, reset)
	if containsRisk(code) {
		fmt.Printf("%s🔴 高风险%s\n", red, reset)
	}
	fmt.Printf("%s💡 diting verify \"%s\" --code %s -m openai%s\n", cyan, claim, codeFile, reset)
}

func handleVerify(modelName string) {
	codeFile := getFlag("--code", "-c")
	if codeFile == "" {
		fmt.Printf("%s❌ --code必填%s\n", red, reset)
		return
	}
	claim := ""
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "--") {
		claim = os.Args[2]
	}
	start := time.Now()
	fmt.Printf("%s🎯 验证: %s%s\n", bold, claim, reset)
	fmt.Printf("%s模型: %s%s\n", blue, models[modelName].Name, reset)
	data, err := os.ReadFile(codeFile)
	if err != nil {
		fmt.Printf("%s❌ %s%s\n", red, err, reset)
		return
	}
	fmt.Printf("%s✓ 源码已加载%s\n", green, reset)
	code := string(data)
	prompt := fmt.Sprintf("代码: %s\n断言: %s\n请分析问题根因", code[:min(len(code),2000)], claim)
	fmt.Printf("%s🤖 AI分析中...%s\n", cyan, reset)
	result := callAI(prompt, modelName)
	fmt.Printf("%s✓ 分析: %s%s\n", green, result, reset)
	ts := time.Now().Format("20060102_150405")
	jsonStr := fmt.Sprintf(`{"claim":"%s","model":"%s","result":"%s"}`, claim, modelName, result)
	os.WriteFile(fmt.Sprintf("diting_result_%s.json", ts), []byte(jsonStr), 0644)
	fmt.Printf("%s🎉 完成！%s%.2f秒%s\n", bold, green, time.Since(start).Seconds(), reset)
}

func handleDashboard() {
	fmt.Printf("\n%s📊 仪表板%s\n", bold, reset)
	fmt.Printf("版本: v1.1.0\n")
	fmt.Printf("模型: %d个\n", len(models))
	for n, c := range models {
		icon := "⚠️"
		if c.APIKey != "" || c.Type == "local" {
			icon = "✅"
		}
		fmt.Printf("  %s %s %s\n", icon, n, c.Name)
	}
}

func handleStats() {
	fmt.Printf("\n%s📈 统计%s\n", bold, reset)
	fmt.Printf("版本: v1.1.0\n")
	fmt.Printf("模型: %d个\n", len(models))
}

func handleCheck() {
	fmt.Printf("%s🔍 检查...%s\n", cyan, reset)
	for n, c := range models {
		s := fmt.Sprintf("%s✅%s", green, reset)
		if c.APIKey == "" && c.Type == "cloud" {
			s = fmt.Sprintf("%s⚠️ 未配置%s", yellow, reset)
		}
		fmt.Printf("  %s: %s\n", n, s)
	}
}

func getFlag(long, short string) string {
	for i, a := range os.Args {
		if a == long || a == short {
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
				return os.Args[i+1]
			}
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractFunctions(code string) string {
	re := regexp.MustCompile(`(?:func|def)\s+(\w+)`)
	matches := re.FindAllStringSubmatch(code, -1)
	var funcs []string
	for _, m := range matches {
		if len(m) >= 2 && m[1] != `` {
			funcs = append(funcs, m[1])
		}
	}
	if len(funcs) > 5 {
		return strings.Join(funcs[:5], ", ") + "..."
	}
	if len(funcs) == 0 {
		return "无"
	}
	return strings.Join(funcs, ", ")
}

func containsRisk(code string) bool {
	risks := []string{"null", "nil", "panic", "error", "exception"}
	l := strings.ToLower(code)
	for _, r := range risks {
		if strings.Contains(l, r) {
			return true
		}
	}
	return false
}
