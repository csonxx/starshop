package scripts

import (
	"sort"

	"star/server/internal/model"
)

func BuildTagSeed() []model.Tag {
	tags := make([]model.Tag, 0, len(STYLES)+len(SPACES)+len(COLORS)+len(PRICES)+32)
	for i, style := range STYLES {
		tags = append(tags, model.Tag{Type: model.TagStyle, Name: style.Name, Value: style.Key, Enabled: true, Sort: i + 1})
	}
	for i, space := range SPACES {
		tags = append(tags, model.Tag{Type: model.TagSpace, Name: space, Value: space, Enabled: true, Sort: i + 1})
	}
	for i, color := range COLORS {
		tags = append(tags, model.Tag{Type: model.TagColor, Name: color, Value: color, Enabled: true, Sort: i + 1})
	}
	sizes := make(map[string]struct{})
	for _, values := range SPACE_SIZES_MAP {
		for _, value := range values {
			sizes[value] = struct{}{}
		}
	}
	sortedSizes := make([]string, 0, len(sizes))
	for value := range sizes {
		sortedSizes = append(sortedSizes, value)
	}
	sort.Strings(sortedSizes)
	for i, size := range sortedSizes {
		tags = append(tags, model.Tag{Type: model.TagSize, Name: size, Value: size, Enabled: true, Sort: i + 1})
	}
	for i, price := range PRICES {
		tags = append(tags, model.Tag{Type: model.TagPrice, Name: price, Value: price, Enabled: true, Sort: i + 1})
	}
	return tags
}
