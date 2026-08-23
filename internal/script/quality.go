package script

import (
	"fmt"
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

type Quality struct {
	Kind       string   `json:"kind"`
	WordCount  int      `json:"word_count"`
	Duration   int      `json:"duration"`
	HasCTA     bool     `json:"has_cta"`
	HasProduct bool     `json:"has_product"`
	Warnings   []string `json:"warnings"`
	Ready      bool     `json:"ready"`
}

func CheckQuality(item model.Script, product string) Quality {
	quality := Quality{Kind: item.Kind, Duration: item.Duration, Warnings: []string{}}
	for _, line := range item.Lines {
		quality.WordCount += len([]rune(line))
		if strings.Contains(line, "现在") || strings.Contains(line, "了解") || strings.Contains(line, "前往") {
			quality.HasCTA = true
		}
		if strings.Contains(line, product) {
			quality.HasProduct = true
		}
	}
	if len(item.Lines) == 0 {
		quality.Warnings = append(quality.Warnings, "script has no lines")
	}
	if item.Duration <= 0 {
		quality.Warnings = append(quality.Warnings, "script duration is missing")
	}
	if item.Kind == "recall" && !quality.HasCTA {
		quality.Warnings = append(quality.Warnings, "recall needs a call to action")
	}
	if item.Kind == "host" && !quality.HasProduct {
		quality.Warnings = append(quality.Warnings, "host should name the product")
	}
	quality.Ready = len(quality.Warnings) == 0
	return quality
}

func ValidateScripts(items []model.Script, product string) []Quality {
	result := make([]Quality, 0, len(items))
	for _, item := range items {
		result = append(result, CheckQuality(item, product))
	}
	sort.SliceStable(result, func(left, right int) bool { return result[left].Kind < result[right].Kind })
	return result
}

func RequireKinds(items []model.Script) error {
	required := map[string]bool{"host": false, "product-3d": false, "recall": false}
	for _, item := range items {
		if _, ok := required[item.Kind]; ok {
			required[item.Kind] = true
		}
	}
	missing := []string{}
	for kind, present := range required {
		if !present {
			missing = append(missing, kind)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing script kinds: %s", strings.Join(missing, ","))
	}
	return nil
}

func CharacterCount(items []model.Script) int {
	count := 0
	for _, item := range items {
		for _, line := range item.Lines {
			count += len([]rune(line))
		}
	}
	return count
}
