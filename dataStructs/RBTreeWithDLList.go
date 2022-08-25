package datastructs

import (
	"fmt"
)

type RBTreeData struct {
	data      interface{}
	myDLLNode *DLLNode
}

type RBTreeWithList[T any] struct {
	RbTree *RBTree[T, RBTreeData]
	DlList *DoublyLinkedList
}

func (rbtwl *RBTreeWithList[T]) Insert(key T, data interface{}, listTraversalFunc func(*DLLNode), rbTreeUpdateFunc func(T, RBTreeData) T) {
	newDLLNode := new(DLLNode)
	newDLLNode.value = data

	newRBTreeData := RBTreeData{
		data:      data,
		myDLLNode: newDLLNode,
	}

	newRBTreeNode := rbtwl.RbTree.InsertNewNode(key, newRBTreeData)
	nodeParent := newRBTreeNode.GetParent()
	if nodeParent == nil {
		leftChild := newRBTreeNode.GetLeftChild()
		rightChild := newRBTreeNode.GetRightChild()
		if leftChild != nil {
			rbtwl.DlList.insertAfter(newDLLNode, leftChild.GetData().myDLLNode)
		} else if rightChild != nil {
			rbtwl.DlList.insertBefore(newDLLNode, rightChild.GetData().myDLLNode)
		} else {
			rbtwl.DlList.append(newDLLNode)
		}
	} else {
		parentDllNode := nodeParent.GetData().myDLLNode

		if rbtwl.RbTree.Less(key, nodeParent.GetKey()) {
			rbtwl.DlList.insertBefore(newDLLNode, parentDllNode)
		} else {
			rbtwl.DlList.insertAfter(newDLLNode, parentDllNode)
		}
	}

	fmt.Println("Got to traversal of dll")
	rbtwl.DlList.traversePartAndApply(newDLLNode, listTraversalFunc)
	rbTreeUpdateFunc(newRBTreeNode.Key, newRBTreeNode.Data)
}

func (rbtwl *RBTreeWithList[T]) PrintCurrentState() {
	fmt.Print("Tree data:  ")
	InOrderTraversal(rbtwl.RbTree.Root)
	fmt.Println()
	fmt.Print("Doubly linked list data:  ")
	rbtwl.DlList.TraverseList()
	fmt.Println()
}
