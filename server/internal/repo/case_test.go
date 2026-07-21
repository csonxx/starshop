package repo

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildCaseMatchUsesAndAcrossDimensions(t *testing.T) {
	match := buildCaseMatch(CaseFilter{
		Style:      "modern",
		Space:      []string{"客厅", "餐厅"},
		Color:      []string{"原木", "奶油白"},
		Size:       []string{"2.4m满墙电视柜"},
		Price:      []string{"1-3万", "3-5万"},
		OnlyActive: true,
	})

	if match["style"] != "modern" || match["enabled"] != true {
		t.Fatalf("基础过滤不正确: %#v", match)
	}
	if !reflect.DeepEqual(match["space"], bson.M{"$in": []string{"客厅", "餐厅"}}) {
		t.Fatalf("空间多选不正确: %#v", match["space"])
	}
	if !reflect.DeepEqual(match["colors"], bson.M{"$in": []string{"原木", "奶油白"}}) {
		t.Fatalf("颜色多选不正确: %#v", match["colors"])
	}
	if match["size"] != "2.4m满墙电视柜" {
		t.Fatalf("尺寸单选不正确: %#v", match["size"])
	}
	if !reflect.DeepEqual(match["priceLabel"], bson.M{"$in": []string{"1-3万", "3-5万"}}) {
		t.Fatalf("价格多选不正确: %#v", match["priceLabel"])
	}
	if _, exists := match["$and"]; exists {
		t.Fatalf("不同字段天然为 AND，不应生成冗余 $and: %#v", match)
	}
}

func TestBuildCaseMatchEscapesSearch(t *testing.T) {
	match := buildCaseMatch(CaseFilter{Q: ".*"})
	or, ok := match["$or"].([]bson.M)
	if !ok || len(or) != 3 {
		t.Fatalf("搜索条件不正确: %#v", match)
	}
	regex := or[0]["title"].(bson.M)["$regex"]
	if regex != `\.\*` {
		t.Fatalf("正则未转义: %q", regex)
	}
}

func TestSplitListsSupportsRepeatedAndCommaSeparatedValues(t *testing.T) {
	got := SplitLists("客厅,餐厅", "餐厅", " 主卧 ", "")
	want := []string{"客厅", "餐厅", "主卧"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SplitLists() = %#v, want %#v", got, want)
	}
}
