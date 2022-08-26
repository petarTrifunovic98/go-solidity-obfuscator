package datastructs

import (
	"errors"
	"fmt"
)

type RBNode[T any, D any] struct {
	Data       D
	Key        T
	leftChild  *RBNode[T, D]
	rightChild *RBNode[T, D]
	Parent     *RBNode[T, D]
	isBlack    bool
}

func (node *RBNode[T, D]) swapColor(withNode *RBNode[T, D]) error {
	if withNode == nil {
		return errors.New("Cannot swap colors with a nil node!")
	}

	tempColor := node.isBlack
	node.isBlack = withNode.isBlack
	withNode.isBlack = tempColor

	return nil
}

func (node *RBNode[T, D]) GetParent() *RBNode[T, D] {
	return node.Parent
}

func (node *RBNode[T, D]) GetLeftChild() *RBNode[T, D] {
	return node.leftChild
}

func (node *RBNode[T, D]) GetRightChild() *RBNode[T, D] {
	return node.rightChild
}

func (node *RBNode[T, D]) GetKey() T {
	return node.Key
}

func (node *RBNode[T, D]) GetData() D {
	return node.Data
}

func (node *RBNode[T, D]) SetKey(newKey T) {
	node.Key = newKey
}

type RBTree[T any, D any] struct {
	Root *RBNode[T, D]
	Less func(T, T) bool
}

func (tree *RBTree[T, D]) rightRotate(atNode *RBNode[T, D]) error {
	if atNode == nil || atNode.leftChild == nil {
		return errors.New("Illegal rotation! Either the node or its left child is nil!")
	}

	parentNode := atNode.Parent
	atNodeLeftChild := atNode.leftChild

	atNode.leftChild = atNodeLeftChild.rightChild
	if atNodeLeftChild.rightChild != nil {
		atNodeLeftChild.rightChild.Parent = atNode
	}

	atNodeLeftChild.rightChild = atNode
	atNode.Parent = atNodeLeftChild

	tree.replaceChild(parentNode, atNodeLeftChild, atNode)

	return nil
}

func (tree *RBTree[T, D]) leftRotate(atNode *RBNode[T, D]) error {
	if atNode == nil || atNode.rightChild == nil {
		return errors.New("Illegal rotation! Either the node or its right child is nil!")
	}

	parentNode := atNode.Parent
	atNodeRightChild := atNode.rightChild

	atNode.rightChild = atNodeRightChild.leftChild
	if atNodeRightChild.leftChild != nil {
		atNodeRightChild.leftChild.Parent = atNode
	}

	atNodeRightChild.leftChild = atNode
	atNode.Parent = atNodeRightChild

	tree.replaceChild(parentNode, atNodeRightChild, atNode)

	return nil
}

func (tree *RBTree[T, D]) replaceChild(parent *RBNode[T, D], newChild *RBNode[T, D], oldChild *RBNode[T, D]) {
	if parent == nil {
		tree.Root = newChild
	} else if parent.leftChild == oldChild {
		parent.leftChild = newChild
	} else {
		parent.rightChild = newChild
	}

	if newChild != nil {
		newChild.Parent = parent
	}
}

func (tree *RBTree[T, D]) adaptTreeToRBConditions(mainNode *RBNode[T, D]) {
	tree.Root.isBlack = true
	if mainNode == tree.Root || mainNode.isBlack == true || mainNode.Parent.isBlack == true {
		return
	}

	parentNode := mainNode.Parent
	grandParentNode := parentNode.Parent

	if !parentNode.isBlack {
		var uncleNode *RBNode[T, D]
		if grandParentNode.leftChild == parentNode {
			uncleNode = grandParentNode.rightChild
		} else {
			uncleNode = grandParentNode.leftChild
		}

		if uncleNode != nil && !uncleNode.isBlack {
			parentNode.isBlack = !parentNode.isBlack
			uncleNode.isBlack = !uncleNode.isBlack
			grandParentNode.isBlack = !grandParentNode.isBlack
			mainNode = grandParentNode
			tree.adaptTreeToRBConditions(mainNode)
		} else if uncleNode == grandParentNode.leftChild { //parent is grandparents right node
			if mainNode == parentNode.leftChild {
				tree.rightRotate(parentNode)
				tree.leftRotate(grandParentNode)
				mainNode.isBlack = true
				grandParentNode.isBlack = false
			} else {
				tree.leftRotate(grandParentNode)
				parentNode.isBlack = true
				grandParentNode.isBlack = false
			}
		} else { //parent is grandparents left node
			if mainNode == parentNode.rightChild {
				tree.leftRotate(parentNode)
				tree.rightRotate(grandParentNode)
				mainNode.isBlack = true
				grandParentNode.isBlack = false
			} else {
				tree.rightRotate(grandParentNode)
				parentNode.isBlack = true
				grandParentNode.isBlack = false
			}
		}
	}

}

func (tree *RBTree[T, D]) InsertNewNode(key T, data D) *RBNode[T, D] {
	newNode := RBNode[T, D]{
		Key:     key,
		Data:    data,
		isBlack: false,
	}

	tree.Insert(&newNode)
	return &newNode
}

func (tree *RBTree[T, D]) Insert(node *RBNode[T, D]) *RBNode[T, D] {
	currentNode := tree.Root
	var parentNode *RBNode[T, D] = nil

	if currentNode == nil {
		node.isBlack = true
		tree.Root = node
		return nil
	}

	node.isBlack = false
	for true {
		parentNode = currentNode
		if tree.Less(currentNode.Key, node.Key) {
			currentNode = currentNode.rightChild
			if currentNode == nil {
				parentNode.rightChild = node
				break
			}
		} else {
			currentNode = currentNode.leftChild
			if currentNode == nil {
				parentNode.leftChild = node
				break
			}
		}
	}

	node.Parent = parentNode
	tree.adaptTreeToRBConditions(node)

	return node.Parent
}

func InOrderTraversal[T any, D any](root *RBNode[T, D]) {
	if root == nil {
		return
	}

	InOrderTraversal(root.leftChild)
	fmt.Print("{")
	fmt.Print(root.Key)
	fmt.Print(" ")
	if root.isBlack {
		fmt.Print("B")
	} else {
		fmt.Print("R")
	}
	fmt.Print("} ")
	InOrderTraversal(root.rightChild)
}

func PreOrderTraversal[T any, D any](root *RBNode[T, D]) {
	if root == nil {
		return
	}

	fmt.Print("{")
	fmt.Print(root.Key)
	fmt.Print(" ")
	if root.isBlack {
		fmt.Print("B")
	} else {
		fmt.Print("R")
	}
	fmt.Print("} ")
	PreOrderTraversal(root.leftChild)
	PreOrderTraversal(root.rightChild)
}
