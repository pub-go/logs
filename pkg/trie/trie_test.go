package trie_test

import (
	"reflect"
	"testing"

	"code.gopub.tech/logs/pkg/trie"
)

func TestTrie(t *testing.T) {
	tree := trie.NewTree(10)
	tree.Insert("main", 20)
	tree.Insert("gitee.com", 25)
	tree.Insert("gitee.com/pub-go", 30)
	tree.Insert("code.gopub.tech/logs", 35)
	tree.Insert("code.gopub.tech/logs/中文", 40)
	tree.Insert("code.gopub.tech/logs/中文/inner", 45)
	tree.Insert("code.gopub.tech/logs/inner", 50)
	if !reflect.DeepEqual(tree.ToMap(), map[string]int{
		"":                              10,
		"main":                          20,
		"gitee.com":                     25,
		"gitee.com/pub-go":              30,
		"code.gopub.tech/logs":          35,
		"code.gopub.tech/logs/中文":       40,
		"code.gopub.tech/logs/中文/inner": 45,
		"code.gopub.tech/logs/inner":    50,
	}) {
		t.Errorf("ToMap fail")
	}
	for _, tCase := range []struct {
		path string
		want int
	}{
		{"", 10},
		{"m", 10},
		{"main", 20},
		{"main.main", 20},
		{"github.com", 10},
		{"gitee.com", 25},
		{"gitee.com/", 25},
		{"gitee.com/pub-go", 30},
		{"gitee.com/pub-go/", 30},
		{"code.gopub.tech/logs", 35},
		{"code.gopub.tech/logs/", 35},
		{"code.gopub.tech/logs/中", 35},
		{"code.gopub.tech/logs/中文", 40},
		{"code.gopub.tech/logs/中文/", 40},
		{"code.gopub.tech/logs/中文/i", 40},
		{"code.gopub.tech/logs/中文/inner", 45},
		{"code.gopub.tech/logs/i", 35},
		{"code.gopub.tech/logs/inner", 50},
	} {
		if got := tree.Search(tCase.path); got != tCase.want {
			t.Errorf("got=%v want=%v", got, tCase.want)
		}
	}
}
