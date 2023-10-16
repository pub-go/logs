package trie

type Tree[T any] struct {
	root *node[T]
}

//	root=5
//
// m      g
// a      i
// i      t
// n      h-u-b - / - a
// 10         20      30
// ""       -> 5
// m        -> 5
// main     -> 10
// github   -> 20
// github/a -> 30
type node[T any] struct {
	data *T // 为了下方需要与 nil 比较 所以取地址
	chd  map[rune]*node[T]
}

func NewTree[T any](rootData T) *Tree[T] {
	return &Tree[T]{root: &node[T]{
		data: &rootData,
		chd:  make(map[rune]*node[T]),
	}}
}

func (t *Tree[T]) Insert(path string, data T) *Tree[T] {
	n := t.root
	for _, r := range path {
		chd, ok := n.chd[r]
		if !ok {
			chd = &node[T]{chd: make(map[rune]*node[T])}
			n.chd[r] = chd
		}
		n = chd
	}
	n.data = &data
	return t
}

func (t *Tree[T]) Search(path string) (result T) {
	currentNode := t.root
	if currentNode.data != nil {
		result = *currentNode.data // 以跟节点结果为兜底
	}
	for _, r := range path { // 向下搜索前缀树
		if chd, ok := currentNode.chd[r]; ok { // 如果有这个节点
			currentNode = chd // 更新当前指向 以便继续向下搜索
			if currentNode.data != nil {
				result = *currentNode.data // 更新当前值
			}
		} else {
			break // 没有了 就用最靠近的前缀结果
		}
	}
	return result
}

func (t *Tree[T]) ToMap() map[string]T {
	result := make(map[string]T)
	dump("", t.root, result)
	return result
}

func dump[T any](path string, n *node[T], result map[string]T) {
	if n.data != nil {
		result[path] = *n.data
	}
	for r, chd := range n.chd {
		dump(path+string(r), chd, result)
	}
}
