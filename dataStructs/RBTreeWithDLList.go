package datastructs

import (
	"golang.org/x/exp/constraints"
)

type RBTreeData struct {
	data      interface{}
	myDLLNode *DLLNode
}

type RBTreeDataCompatible interface {
	RBTreeData | ~int
}

type Petar struct {
	myDLLNode *DLLNode
	data      interface{}
}

type RBTreeWithList[T constraints.Ordered] struct {
	rbTree *RBTree[T, RBTreeData]
	dlList *DoublyLinkedList
}

func (rbtwl *RBTreeWithList[T]) insertNewNode(key T) {

}

func (rbtwl *RBTreeWithList[T]) insert(key T, data interface{}) {
	newDLLNode := new(DLLNode)
	newDLLNode.value = data

	newRBTreeData := RBTreeData{
		data:      data,
		myDLLNode: newDLLNode,
	}

	newRBTreeNode := rbtwl.rbTree.InsertNewNode(key, newRBTreeData)
	nodeParent := newRBTreeNode.GetParent()
	if nodeParent == nil {
		rbtwl.dlList.append(newDLLNode)
	} else {
		parentDllNode := nodeParent.GetData().myDLLNode

		if key < nodeParent.GetKey() {
			rbtwl.dlList.insertBefore(newDLLNode, parentDllNode)
		} else {
			rbtwl.dlList.insertAfter(newDLLNode, parentDllNode)
		}
	}
}
