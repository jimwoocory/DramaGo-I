package productionqa

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

const (
	CodeDurationCompression       = "duration_compression"
	CodeSceneCanonConflict        = "scene_canon_conflict"
	CodeRepeatedGlobalBoilerplate = "repeated_global_boilerplate"
	CodeUnsupportedPropUsage      = "unsupported_prop_usage"
)

// Finding is one deterministic production-quality issue detected from project documents.
type Finding struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Evidence string   `json:"evidence,omitempty"`
}

// Report is a compact quality audit of one screenplay/storyboard package.
type Report struct {
	TargetDurationSeconds     float64   `json:"targetDurationSeconds,omitempty"`
	StoryboardDurationSeconds float64   `json:"storyboardDurationSeconds,omitempty"`
	DurationRatio             float64   `json:"durationRatio,omitempty"`
	StoryboardGroupCount      int       `json:"storyboardGroupCount,omitempty"`
	Findings                  []Finding `json:"findings"`
}

var (
	durationMinutePattern = regexp.MustCompile(`(?:本集)?(?:目标)?时长[^\n。；]{0,8}?(?:约|大约|预计)?\s*([0-9]+(?:\.[0-9]+)?|[零〇一二两三四五六七八九十百]+)\s*分钟`)
	storyboardH2Pattern   = regexp.MustCompile(`(?m)^##\s+.+$`)
	timeRangePattern      = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*[–—-]\s*([0-9]+(?:\.[0-9]+)?)`)
	colorObjectPattern    = regexp.MustCompile(`(银灰色?|灰白|灰色|绿色|蓝色|红色|黄色|黑色|白色|棕色|金色|银色)[^。；，,\n]{0,12}(卷帘门|门帘|铁门|木门|窗框|窗|墙面|墙|餐椅|椅|桌面|桌|柜门|柜)`)
	propHeadingPattern    = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	propUsagePattern      = regexp.MustCompile(`^(?:[-*]\s*)?(?:\*\*)?使用关系(?:\*\*)?[：:]\s*(.+)$`)
)

// AuditLegacyDocuments audits the existing Markdown-oriented production path.
// It is deliberately deterministic and conservative; ambiguous semantic issues
// remain warnings for human/agent review rather than silently rewriting content.
func AuditLegacyDocuments(screenplayMarkdown string, sceneMarkdown string, propMarkdown string, storyboardMarkdown string) Report {
	targetDuration := ParseTargetDurationSeconds(screenplayMarkdown)
	storyboardDuration, groupCount := StoryboardDurationSeconds(storyboardMarkdown)
	report := Report{
		TargetDurationSeconds:     targetDuration,
		StoryboardDurationSeconds: storyboardDuration,
		StoryboardGroupCount:      groupCount,
		Findings:                  []Finding{},
	}
	if targetDuration > 0 && storyboardDuration > 0 {
		report.DurationRatio = storyboardDuration / targetDuration
		if report.DurationRatio < 0.75 {
			report.Findings = append(report.Findings, Finding{
				Code:     CodeDurationCompression,
				Severity: SeverityError,
				Message:  fmt.Sprintf("分镜总时长 %.1f 秒，仅达到目标 %.1f 秒的 %.1f%%", storyboardDuration, targetDuration, report.DurationRatio*100),
				Evidence: fmt.Sprintf("target=%.1fs storyboard=%.1fs", targetDuration, storyboardDuration),
			})
		}
	}

	report.Findings = append(report.Findings, sceneColorConflicts(sceneMarkdown, storyboardMarkdown)...)
	if finding, ok := repeatedStoryboardBoilerplate(storyboardMarkdown, groupCount); ok {
		report.Findings = append(report.Findings, finding)
	}
	report.Findings = append(report.Findings, unsupportedPropUsage(screenplayMarkdown, propMarkdown)...)
	return report
}

// ParseTargetDurationSeconds understands Arabic or common Chinese minute counts,
// including text such as “本集时长约十二分钟”.
func ParseTargetDurationSeconds(markdown string) float64 {
	match := durationMinutePattern.FindStringSubmatch(markdown)
	if len(match) != 2 {
		return 0
	}
	minutes, ok := parseMinuteValue(match[1])
	if !ok || minutes <= 0 {
		return 0
	}
	return minutes * 60
}

// StoryboardDurationSeconds sums the maximum local end time of each H2 group.
// Existing MediaGo storyboard groups reset their local clock to 0 for each group.
func StoryboardDurationSeconds(markdown string) (float64, int) {
	lines := strings.Split(markdown, "\n")
	groupCount := 0
	currentMax := 0.0
	total := 0.0
	flush := func() {
		if groupCount > 0 {
			total += currentMax
			currentMax = 0
		}
	}
	for _, line := range lines {
		if storyboardH2Pattern.MatchString(line) {
			flush()
			groupCount++
			continue
		}
		if groupCount == 0 {
			continue
		}
		for _, match := range timeRangePattern.FindAllStringSubmatch(line, -1) {
			if len(match) != 3 {
				continue
			}
			end, err := strconv.ParseFloat(match[2], 64)
			if err == nil && end > currentMax {
				currentMax = end
			}
		}
	}
	flush()
	return math.Round(total*100) / 100, groupCount
}

func sceneColorConflicts(canon string, storyboard string) []Finding {
	canonFacts := colorFacts(canon)
	storyboardFacts := colorFacts(storyboard)
	findings := []Finding{}
	keys := make([]string, 0, len(storyboardFacts))
	for object := range storyboardFacts {
		keys = append(keys, object)
	}
	sort.Strings(keys)
	for _, object := range keys {
		canonColors := canonFacts[object]
		if len(canonColors) == 0 {
			continue
		}
		for color := range storyboardFacts[object] {
			if canonColors[color] {
				continue
			}
			findings = append(findings, Finding{
				Code:     CodeSceneCanonConflict,
				Severity: SeverityError,
				Message:  fmt.Sprintf("分镜把%s写成%s，与场景 Canon 颜色冲突", object, color),
				Evidence: fmt.Sprintf("canon=%s storyboard=%s", joinSet(canonColors), color),
			})
		}
	}
	return findings
}

func colorFacts(text string) map[string]map[string]bool {
	facts := map[string]map[string]bool{}
	for _, match := range colorObjectPattern.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}
		color := normalizeColor(match[1])
		object := strings.TrimSpace(match[2])
		if color == "" || object == "" {
			continue
		}
		if facts[object] == nil {
			facts[object] = map[string]bool{}
		}
		facts[object][color] = true
	}
	return facts
}

func normalizeColor(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "银灰", "银灰色":
		return "银灰色"
	default:
		return value
	}
}

func repeatedStoryboardBoilerplate(storyboard string, groupCount int) (Finding, bool) {
	if groupCount <= 0 {
		return Finding{}, false
	}
	tokens := []string{"无字幕", "无水印", "无 BGM"}
	counts := make([]string, 0, len(tokens))
	repeated := 0
	for _, token := range tokens {
		count := strings.Count(storyboard, token)
		counts = append(counts, fmt.Sprintf("%s=%d", token, count))
		if count >= groupCount*2 {
			repeated++
		}
	}
	if repeated < 2 {
		return Finding{}, false
	}
	return Finding{
		Code:     CodeRepeatedGlobalBoilerplate,
		Severity: SeverityWarning,
		Message:  "分镜逐镜重复全局禁用/质量约束，应迁移到 Style Profile 或生成端固定约束",
		Evidence: strings.Join(counts, " "),
	}, true
}

func unsupportedPropUsage(screenplay string, propMarkdown string) []Finding {
	verbs := []string{"拍", "录", "传出", "分享", "上传", "清洁", "加热", "拆开", "砸坏"}
	findings := []Finding{}
	currentProp := ""
	seen := map[string]bool{}
	for _, rawLine := range strings.Split(propMarkdown, "\n") {
		line := strings.TrimSpace(rawLine)
		if heading := propHeadingPattern.FindStringSubmatch(line); len(heading) == 2 {
			currentProp = strings.TrimSpace(heading[1])
			continue
		}
		usage := propUsagePattern.FindStringSubmatch(line)
		if len(usage) != 2 || currentProp == "" {
			continue
		}
		usageText := usage[1]
		missing := []string{}
		for _, verb := range verbs {
			if strings.Contains(usageText, verb) && !strings.Contains(screenplay, verb) {
				missing = append(missing, verb)
			}
		}
		if len(missing) == 0 {
			continue
		}
		key := currentProp + "\x00" + strings.Join(missing, ",")
		if seen[key] {
			continue
		}
		seen[key] = true
		findings = append(findings, Finding{
			Code:     CodeUnsupportedPropUsage,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("道具“%s”的使用关系包含原剧本未出现的动作：%s", currentProp, strings.Join(missing, "、")),
			Evidence: usageText,
		})
	}
	return findings
}

func parseMinuteValue(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return parsed, true
	}
	integer, ok := parseChineseInteger(value)
	return float64(integer), ok
}

func parseChineseInteger(value string) (int, bool) {
	digits := map[rune]int{'零': 0, '〇': 0, '一': 1, '二': 2, '两': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9}
	if value == "十" {
		return 10, true
	}
	if strings.ContainsRune(value, '百') {
		parts := strings.SplitN(value, "百", 2)
		hundreds := 1
		if parts[0] != "" {
			runes := []rune(parts[0])
			if len(runes) != 1 {
				return 0, false
			}
			var ok bool
			hundreds, ok = digits[runes[0]]
			if !ok || hundreds == 0 {
				return 0, false
			}
		}
		rest := 0
		if parts[1] != "" {
			var ok bool
			rest, ok = parseChineseInteger(strings.TrimPrefix(parts[1], "零"))
			if !ok {
				return 0, false
			}
		}
		return hundreds*100 + rest, true
	}
	if strings.ContainsRune(value, '十') {
		parts := strings.SplitN(value, "十", 2)
		tens := 1
		if parts[0] != "" {
			runes := []rune(parts[0])
			if len(runes) != 1 {
				return 0, false
			}
			var ok bool
			tens, ok = digits[runes[0]]
			if !ok || tens == 0 {
				return 0, false
			}
		}
		ones := 0
		if parts[1] != "" {
			runes := []rune(parts[1])
			if len(runes) != 1 {
				return 0, false
			}
			var ok bool
			ones, ok = digits[runes[0]]
			if !ok {
				return 0, false
			}
		}
		return tens*10 + ones, true
	}
	runes := []rune(value)
	if len(runes) != 1 {
		return 0, false
	}
	result, ok := digits[runes[0]]
	return result, ok
}

func joinSet(values map[string]bool) string {
	items := make([]string, 0, len(values))
	for value := range values {
		items = append(items, value)
	}
	sort.Strings(items)
	return strings.Join(items, "/")
}
