package datastructs

import (
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
)

type RBNode[T constraints.Ordered] struct {
	Data       interface{}
	Key        T
	leftChild  *RBNode[T]
	rightChild *RBNode[T]
	parent     *RBNode[T]
	isBlack    bool
}

func (node *RBNode[T]) swapColor(withNode *RBNode[T]) error {
	if withNode == nil {
		return errors.New("Cannot swap colors with a nil node!")
	}

	tempColor := node.isBlack
	node.isBlack = withNode.isBlack
	withNode.isBlack = tempColor

	return nil
}

type RBTree[T constraints.Ordered] struct {
	Root *RBNode[T]
}

func (tree *RBTree[T]) rightRotate(atNode *RBNode[T]) error {
	if atNode == nil || atNode.leftChild == nil {
		return errors.New("Illegal rotation! Either the node or its left child is nil!")
	}

	parentNode := atNode.parent
	atNodeLeftChild := atNode.leftChild

	atNode.leftChild = atNodeLeftChild.rightChild
	if atNodeLeftChild.rightChild != nil {
		atNodeLeftChild.rightChild.parent = atNode
	}

	atNodeLeftChild.rightChild = atNode
	atNode.parent = atNodeLeftChild

	tree.replaceChild(parentNode, atNodeLeftChild, atNode)

	return nil
}

func (tree *RBTree[T]) leftRotate(atNode *RBNode[T]) error {
	if atNode == nil || atNode.rightChild == nil {
		return errors.New("Illegal rotation! Either the node or its right child is nil!")
	}

	parentNode := atNode.parent
	atNodeRightChild := atNode.rightChild

	atNode.rightChild = atNodeRightChild.leftChild
	if atNodeRightChild.leftChild != nil {
		atNodeRightChild.leftChild.parent = atNode
	}

	atNodeRightChild.leftChild = atNode
	atNode.parent = atNodeRightChild

	tree.replaceChild(parentNode, atNodeRightChild, atNode)

	return nil
}

func (tree *RBTree[T]) replaceChild(parent *RBNode[T], newChild *RBNode[T], oldChild *RBNode[T]) {
	if parent == nil {
		tree.Root = newChild
	} else if parent.leftChild == oldChild {
		parent.leftChild = newChild
	} else {
		parent.rightChild = newChild
	}

	if newChild != nil {
		newChild.parent = parent
	}
}

func (tree *RBTree[T]) adaptTreeToRBConditions(mainNode *RBNode[T]) {
	tree.Root.isBlack = true
	if mainNode == tree.Root || mainNode.isBlack == true || mainNode.parent.isBlack == true {
		return
	}

	parentNode := mainNode.parent
	grandParentNode := parentNode.parent

	if !parentNode.isBlack {
		var uncleNode *RBNode[T]
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

func (tree *RBTree[T]) InsertAndGetParent(node *RBNode[T]) *RBNode[T] {
	currentNode := tree.Root
	var parentNode *RBNode[T] = nil

	if currentNode == nil {
		node.isBlack = true
		tree.Root = node
		return nil
	}

	node.isBlack = false
	for true {
		parentNode = currentNode
		if node.Key > currentNode.Key {
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

	node.parent = parentNode
	tree.adaptTreeToRBConditions(node)

	return parentNode
}

func InOrderTraversal[T constraints.Ordered](root *RBNode[T]) {
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

func PreOrderTraversal[T constraints.Ordered](root *RBNode[T]) {
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
