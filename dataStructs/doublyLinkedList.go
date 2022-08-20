package datastructs

type DLLNode struct {
	prev  *DLLNode
	next  *DLLNode
	value interface{}
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
	}
}
