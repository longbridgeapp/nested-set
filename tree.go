package nestedset

type Tree struct {
	Children []*TreeNode
	data     map[int64]*TreeNode
}

type TreeNode struct {
	*nestedItem
	Children []*TreeNode
}

func initTree(items []*nestedItem) *Tree {
	tree := &Tree{
		data:     make(map[int64]*TreeNode),
		Children: make([]*TreeNode, 0),
	}

	for _, item := range items {
		node := &TreeNode{
			nestedItem: item,
			Children:   make([]*TreeNode, 0),
		}
		tree.data[node.ID] = node
	}

	for _, item := range items {
		node, _ := tree.getNode(item.ID)
		parent, found := tree.getNode(item.ParentID.Int64)
		if !found {
			tree.Children = append(tree.Children, node)
		} else {
			parent.Children = append(parent.Children, node)
		}
	}

	return tree
}

func (tree *Tree) getNode(id int64) (node *TreeNode, found bool) {
	if id == 0 {
		return nil, false
	}
	node, found = tree.data[id]
	return
}

func (tree *Tree) addNestedItem(item *nestedItem) {
	node := &TreeNode{
		nestedItem: item,
		Children:   make([]*TreeNode, 0),
	}
	parent, found := tree.getNode(item.ParentID.Int64)
	if !found {
		tree.Children = append(tree.Children, node)
	} else {
		parent.Children = append(parent.Children, node)
	}
}

func (tree *Tree) rebuild() *Tree {
	count, depth := 0, 0
	for _, node := range tree.Children {
		count = travelNode(node, count, depth)
	}
	return tree
}

func travelNode(node *TreeNode, lft, depth int) int {
	original := &nestedItem{
		ID:            node.ID,
		ParentID:      node.ParentID,
		Depth:         node.Depth,
		Lft:           node.Lft,
		Rgt:           node.Rgt,
		ChildrenCount: node.ChildrenCount,
	}
	lft += 1
	node.Lft = lft
	node.ChildrenCount = len(node.Children)
	node.Depth = depth
	for _, childNode := range node.Children {
		lft = travelNode(childNode, lft, depth+1)
	}
	lft += 1
	node.Rgt = lft
	node.IsChanged = node.nestedItem.IsPositionSame(original)
	return lft
}
