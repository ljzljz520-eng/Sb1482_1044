package timeline

import (
	"fmt"
	"strings"

	"example.com/materialconsole/internal/model"
)

func Render(items []model.Timeline) string {
	var builder strings.Builder
	for _, item := range items {
		builder.WriteString(fmt.Sprintf("%02d %s %ds %s\n", item.Position, sceneLabel(item), item.Seconds, item.AssetID))
	}
	return builder.String()
}
