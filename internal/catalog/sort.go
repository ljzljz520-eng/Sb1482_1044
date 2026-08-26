package catalog

import (
	"sort"

	"example.com/materialconsole/internal/model"
)

func sortMaterials(items []model.MaterialRecord) []model.MaterialRecord {
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].Sequence == items[right].Sequence {
			return items[left].ID < items[right].ID
		}
		return items[left].Sequence < items[right].Sequence
	})
	return items
}
