package trie

type Tree[T any] struct {
	root *node[T]
}

type node[T any] struct {
	data *T
	chd  map[rune]*node[T]
}

func NewTree[T any](rootData T) *Tree[T] {
	return &Tree[T]{root: &node[T]{data: &rootData, chd: make(map[rune]*node[T])}}
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
	n := t.root
	if n.data != nil {
		result = *n.data
	}
	for _, r := range path {
		if chd, ok := n.chd[r]; ok {
			n = chd
			if n.data != nil {
				result = *n.data
			}
		} else {
			break
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
