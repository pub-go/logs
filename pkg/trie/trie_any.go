//go:build !go1.18
// +build !go1.18

package trie

type any = interface{}

type Tree struct {
	root *node
}

type node struct {
	data any
	chd  map[rune]*node
}

func NewTree(rootData any) *Tree {
	return &Tree{
		root: &node{
			data: rootData,
			chd:  make(map[rune]*node),
		},
	}
}

func (t *Tree) Insert(path string, data any) *Tree {
	n := t.root
	for _, r := range path {
		chd, ok := n.chd[r]
		if !ok {
			chd = &node{chd: make(map[rune]*node)}
			n.chd[r] = chd
		}
		n = chd
	}
	n.data = data
	return t
}

func (t *Tree) Search(path string) (result any) {
	currentNode := t.root
	if currentNode.data != nil {
		result = currentNode.data // 以跟节点结果为兜底
	}
	for _, r := range path { // 向下搜索前缀树
		if chd, ok := currentNode.chd[r]; ok { // 如果有这个节点
			currentNode = chd // 更新当前指向 以便继续向下搜索
			if currentNode.data != nil {
				result = currentNode.data // 更新当前值
			}
		} else {
			break // 没有了 就用最靠近的前缀结果
		}
	}
	return result
}

func (t *Tree) ToMap() map[string]any {
	result := make(map[string]any)
	dump("", t.root, result)
	return result
}

func dump(path string, n *node, result map[string]any) {
	if n.data != nil {
		result[path] = n.data
	}
	for r, chd := range n.chd {
		dump(path+string(r), chd, result)
	}
}
