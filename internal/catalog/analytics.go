package catalog

import (
	"fmt"
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

type MaterialStats struct {
	Total         int            `json:"total"`
	Ready         int            `json:"ready"`
	Blocked       int            `json:"blocked"`
	WithNotes     int            `json:"with_notes"`
	ByChannel     map[string]int `json:"by_channel"`
	ByKind        map[string]int `json:"by_kind"`
	NextSequence  int            `json:"next_sequence"`
	MissingFields []string       `json:"missing_fields"`
}

type Page struct {
	Items      []model.MaterialRecord `json:"items"`
	Offset     int                    `json:"offset"`
	Limit      int                    `json:"limit"`
	Total      int                    `json:"total"`
	HasMore    bool                   `json:"has_more"`
	NextOffset int                    `json:"next_offset"`
}

func BuildStats(project model.Project) MaterialStats {
	stats := MaterialStats{ByChannel: map[string]int{}, ByKind: map[string]int{}, NextSequence: 1}
	for _, material := range project.Materials {
		stats.Total++
		if material.Status == model.StatusReady || material.Approved {
			stats.Ready++
		}
		if material.Status == model.StatusBlocked {
			stats.Blocked++
		}
		if strings.TrimSpace(material.Note) != "" {
			stats.WithNotes++
		}
		stats.ByChannel[material.Channel]++
		stats.ByKind[material.Kind]++
		if material.Sequence >= stats.NextSequence {
			stats.NextSequence = material.Sequence + 1
		}
	}
	stats.MissingFields = MissingMaterialFields(project.Materials)
	return stats
}

func MissingMaterialFields(items []model.MaterialRecord) []string {
	missing := []string{}
	for _, item := range items {
		fields := []string{}
		if strings.TrimSpace(item.Title) == "" {
			fields = append(fields, "title")
		}
		if strings.TrimSpace(item.Source) == "" {
			fields = append(fields, "source")
		}
		if strings.TrimSpace(item.Channel) == "" {
			fields = append(fields, "channel")
		}
		if len(fields) > 0 {
			missing = append(missing, item.ID+":"+strings.Join(fields, "+"))
		}
	}
	return missing
}

func GroupByChannel(items []model.MaterialRecord) map[string][]model.MaterialRecord {
	groups := map[string][]model.MaterialRecord{}
	for _, item := range items {
		groups[item.Channel] = append(groups[item.Channel], item)
	}
	for channel := range groups {
		groups[channel] = sortMaterials(groups[channel])
	}
	return groups
}

func GroupByKind(items []model.MaterialRecord) map[string][]model.MaterialRecord {
	groups := map[string][]model.MaterialRecord{}
	for _, item := range items {
		groups[item.Kind] = append(groups[item.Kind], item)
	}
	for kind := range groups {
		groups[kind] = sortMaterials(groups[kind])
	}
	return groups
}

func Pending(items []model.MaterialRecord) []model.MaterialRecord {
	pending := make([]model.MaterialRecord, 0, len(items))
	for _, item := range items {
		if !item.Approved && item.Status != model.StatusReady {
			pending = append(pending, item)
		}
	}
	return sortMaterials(pending)
}

func Approved(items []model.MaterialRecord) []model.MaterialRecord {
	approved := make([]model.MaterialRecord, 0, len(items))
	for _, item := range items {
		if item.Approved || item.Status == model.StatusReady {
			approved = append(approved, item)
		}
	}
	return sortMaterials(approved)
}

func FindBySequence(items []model.MaterialRecord, sequence int) (model.MaterialRecord, error) {
	if sequence <= 0 {
		return model.MaterialRecord{}, fmt.Errorf("sequence must be positive")
	}
	for _, item := range items {
		if item.Sequence == sequence {
			return item, nil
		}
	}
	return model.MaterialRecord{}, fmt.Errorf("material sequence not found: %d", sequence)
}

func PageMaterials(items []model.MaterialRecord, offset, limit int) (Page, error) {
	if offset < 0 {
		return Page{}, fmt.Errorf("offset cannot be negative")
	}
	if limit <= 0 {
		return Page{}, fmt.Errorf("limit must be positive")
	}
	ordered := sortMaterials(append([]model.MaterialRecord(nil), items...))
	if offset > len(ordered) {
		offset = len(ordered)
	}
	end := offset + limit
	if end > len(ordered) {
		end = len(ordered)
	}
	page := Page{Items: ordered[offset:end], Offset: offset, Limit: limit, Total: len(ordered), NextOffset: end}
	page.HasMore = end < len(ordered)
	return page, nil
}

func NormalizeSequences(items []model.MaterialRecord) []model.MaterialRecord {
	ordered := sortMaterials(append([]model.MaterialRecord(nil), items...))
	for index := range ordered {
		ordered[index].Sequence = index + 1
	}
	return ordered
}

func ChannelNames(items []model.MaterialRecord) []string {
	seen := map[string]bool{}
	for _, item := range items {
		if strings.TrimSpace(item.Channel) != "" {
			seen[item.Channel] = true
		}
	}
	channels := make([]string, 0, len(seen))
	for channel := range seen {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	return channels
}

func KindNames(items []model.MaterialRecord) []string {
	seen := map[string]bool{}
	for _, item := range items {
		if strings.TrimSpace(item.Kind) != "" {
			seen[item.Kind] = true
		}
	}
	kinds := make([]string, 0, len(seen))
	for kind := range seen {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}
