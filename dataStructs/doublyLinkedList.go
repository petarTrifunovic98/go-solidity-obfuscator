package datastructs

import "fmt"

type DLLNode struct {
	prev  *DLLNode
	next  *DLLNode
	value interface{}
}

func (node *DLLNode) GetPrevious() *DLLNode {
	return node.prev
}

func (node *DLLNode) GetNext() *DLLNode {
	return node.next
}

func (node *DLLNode) GetValue() interface{} {
	return node.value
}

func (node *DLLNode) SetValue(v interface{}) {
	node.value = v
}

type DoublyLinkedList struct {
	first *DLLNode
	last  *DLLNode
}

func (dll *DoublyLinkedList) append(node *DLLNode) {
	if dll.first == nil && dll.last == nil {
		dll.first = node
		dll.last = node
	} else {
		dll.last.next = node
		node.prev = dll.last
		dll.last = node
	}
}

func (dll *DoublyLinkedList) insertAfter(node *DLLNode, prevNode *DLLNode) {
	node.prev = prevNode
	node.next = prevNode.next
	prevNode.next = node
	node.next.prev = node

	if dll.last == prevNode {
		dll.last = node
	}
}

func (dll *DoublyLinkedList) insertBefore(node *DLLNode, nextNode *DLLNode) {
	node.next = nextNode
	node.prev = nextNode.prev
	nextNode.prev = node
	node.prev.next = node

	if dll.first == nextNode {
		dll.first = node
	}
}

func (dll *DoublyLinkedList) traversePartAndApply(startNode *DLLNode, nodeFunc func(*DLLNode)) {
	currentNode := startNode
	for currentNode != nil {
		nodeFunc(currentNode)
		currentNode = currentNode.next
	}
}

func (dll *DoublyLinkedList) TraverseList() {
	currentNode := dll.first
	for currentNode != nil {
		fmt.Print("{")
		fmt.Print(currentNode.value)
		fmt.Print("} ")
		currentNode = currentNode.next
	}
}
