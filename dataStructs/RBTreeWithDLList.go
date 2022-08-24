package datastructs

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

type RBTreeNode struct {
	key        interface{}
	value      interface{}
	dlListNode *DLLNode
}

type RBTreeWithList struct {
	rbTree *rbt.Tree
	dlList *DoublyLinkedList
}

func (rbtwl *RBTreeWithList) insert(key interface{}, RBValue *RBTreeNode, DLLValue *DLLNode) {
	rbtwl.rbTree.Put(key, RBValue)
	RBValue.dlListNode = DLLValue

}
