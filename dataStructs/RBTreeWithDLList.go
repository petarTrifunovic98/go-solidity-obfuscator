package datastructs

import (
	"fmt"
)

type RBTreeData struct {
	data      interface{}
	myDLLNode *DLLNode
}

type DLListValue[T any] struct {
	Value        interface{}
	MyRBTreeNode *RBNode[T, RBTreeData]
}

type RBTreeWithList[T any] struct {
	RbTree *RBTree[T, RBTreeData]
	DlList *DoublyLinkedList
}

func (rbtwl *RBTreeWithList[T]) Insert(key T, data interface{}, listTraversalFunc func(*DLLNode)) {
	newDLLNode := new(DLLNode)

	newRBTreeData := RBTreeData{
		data:      data,
		myDLLNode: newDLLNode,
	}

	newRBTreeNode := rbtwl.RbTree.InsertNewNode(key, newRBTreeData)

	newDLListValue := DLListValue[T]{
		Value:        data,
		MyRBTreeNode: newRBTreeNode,
	}

	newDLLNode.value = newDLListValue

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

		//fmt.Println(newDLLNode, parentDllNode)
		if rbtwl.RbTree.Less(key, nodeParent.GetKey()) {
			rbtwl.DlList.insertBefore(newDLLNode, parentDllNode)
		} else {
			rbtwl.DlList.insertAfter(newDLLNode, parentDllNode)
		}
	}

	rbtwl.DlList.traversePartAndApply(newDLLNode, listTraversalFunc)
}

func (rbtwl *RBTreeWithList[T]) PrintCurrentState() {
	fmt.Print("Tree data inorder:  ")
	InOrderTraversal(rbtwl.RbTree.Root)
	fmt.Println()
	fmt.Print("Tree data preorder:  ")
	PreOrderTraversal(rbtwl.RbTree.Root)
	fmt.Println()
	fmt.Print("Doubly linked list data:  ")
	rbtwl.DlList.TraverseList()
	fmt.Println()
}
