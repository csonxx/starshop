package handler

import (
	"strings"
	"testing"

	"star/server/internal/model"
)

func TestNormalizeCaseBeforeValidation(t *testing.T) {
	in := model.Case{
		Title:     "  案例  ",
		Style:     " modern ",
		Space:     " 客厅 ",
		Colors:    []string{" 原木 ", "原木", ""},
		Cover:     " /img-pool/case_01.jpg ",
		Images:    []string{" /img-pool/case_02.jpg ", ""},
		Materials: []string{" 板材 ", "板材"},
	}
	normalizeCase(&in)
	if in.Title != "案例" || in.Style != "modern" || in.Space != "客厅" {
		t.Fatalf("文本字段未规范化: %#v", in)
	}
	if len(in.Colors) != 1 || in.Colors[0] != "原木" {
		t.Fatalf("颜色未去空去重: %#v", in.Colors)
	}
	if len(in.Images) != 1 || in.Images[0] != "/img-pool/case_02.jpg" {
		t.Fatalf("图片未规范化: %#v", in.Images)
	}
	if err := (&AdminHandler{}).validateCase(in); err != nil {
		t.Fatalf("规范化后的合法案例校验失败: %v", err)
	}
}

func TestValidateCaseRejectsInvalidFields(t *testing.T) {
	base := model.Case{Title: "案例", Style: "modern", Space: "客厅", Cover: "/img-pool/case_01.jpg"}
	tests := []struct {
		name string
		edit func(*model.Case)
		want string
	}{
		{name: "negative price", edit: func(in *model.Case) { in.Price = -1 }, want: "price"},
		{name: "too many colors", edit: func(in *model.Case) { in.Colors = make([]string, caseColorsMax+1) }, want: "colors"},
		{name: "invalid image", edit: func(in *model.Case) { in.Images = []string{"file:///tmp/a.jpg"} }, want: "image"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base
			tt.edit(&in)
			err := (&AdminHandler{}).validateCase(in)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateCase() error = %v, want contains %q", err, tt.want)
			}
		})
	}
}

func TestLabelForPriceUsesServerControlledBuckets(t *testing.T) {
	tests := map[int]string{0: "请询价", 9999: "1万以下", 10000: "1-3万", 30000: "3-5万", 50000: "5-10万", 100000: "10万+"}
	for price, want := range tests {
		if got := labelForPrice(price); got != want {
			t.Fatalf("labelForPrice(%d) = %q, want %q", price, got, want)
		}
	}
}
