package processinformation

import (
	"fmt"
	datastructs "solidity-obfuscator/dataStructs"
	"sync"
)

type kp interface {
	GetCurrentLine() int
	GetReducedLine() int
	SetCurrentLine(int)
	SetReducedLine(int)
}

type keyPair struct {
	currentLine int
	reducedLine int
}

func (k keyPair) GetCurrentLine() int {
	return k.currentLine
}

func (k keyPair) GetReducedLine() int {
	return k.reducedLine
}

func (k keyPair) SetCurrentLine(cl int) {
	k.currentLine = cl
}

func (k keyPair) SetReducedLine(rl int) {
	k.reducedLine = rl
	fmt.Println(k.reducedLine, rl)
}

type spreadPair struct {
	realSpread      int
	increasedSpread int
}

func less(existingKey, newKey keyPair) bool {
	if newKey.reducedLine == 0 {
		return existingKey.currentLine < newKey.currentLine
	} else {
		return existingKey.reducedLine < newKey.reducedLine
	}
}

func traversal[T kp](node *datastructs.DLLNode) {
	prev := node.GetPrevious()
	if prev != nil {
		prevNodeValue := prev.GetValue().(datastructs.DLListValue[T])
		nodeValue := node.GetValue().(datastructs.DLListValue[T])
		newValue := spreadPair{
			realSpread:      nodeValue.Value.(spreadPair).realSpread,
			increasedSpread: nodeValue.Value.(spreadPair).realSpread + prevNodeValue.Value.(spreadPair).increasedSpread,
		}
		node.SetValue(datastructs.DLListValue[T]{
			Value:        newValue,
			MyRBTreeNode: nodeValue.MyRBTreeNode,
		})

		nodeRBTreeKey := any(nodeValue.MyRBTreeNode.Key).(keyPair)
		newKey := keyPair{}
		if nodeRBTreeKey.GetReducedLine() == 0 {
			if newValue.realSpread == -2 {
				fmt.Println(prevNodeValue.Value.(spreadPair))
			}
			newKey.currentLine = nodeRBTreeKey.GetCurrentLine()
			newKey.reducedLine = newKey.currentLine - prevNodeValue.Value.(spreadPair).increasedSpread
		} else {
			newKey.reducedLine = nodeRBTreeKey.GetReducedLine()
			newKey.currentLine = nodeRBTreeKey.GetReducedLine() + prevNodeValue.Value.(spreadPair).increasedSpread
		}
		nodeValue.MyRBTreeNode.SetKey(any(newKey).(T))
	} else {
		nodeValue := node.GetValue().(datastructs.DLListValue[T])
		newValue := spreadPair{
			realSpread:      nodeValue.Value.(spreadPair).realSpread,
			increasedSpread: nodeValue.Value.(spreadPair).realSpread,
		}
		node.SetValue(datastructs.DLListValue[T]{
			Value:        newValue,
			MyRBTreeNode: nodeValue.MyRBTreeNode,
		})
		nodeRBTreeKey := nodeValue.MyRBTreeNode.Key
		newKey := keyPair{
			currentLine: nodeRBTreeKey.GetCurrentLine(),
			reducedLine: nodeRBTreeKey.GetCurrentLine(),
		}
		nodeValue.MyRBTreeNode.SetKey(any(newKey).(T))

	}
}

type sourceCodeChangeInformation struct {
	sourceCodeChangeTracker datastructs.RBTreeWithList[keyPair]
}

var sourceCodeChangeOnce sync.Once
var sourceCodeChangeInstance *sourceCodeChangeInformation

func SourceCodeChangeInformation() *sourceCodeChangeInformation {
	sourceCodeChangeOnce.Do(func() {
		rbTree := datastructs.RBTree[keyPair, datastructs.RBTreeData]{
			Less: less,
		}

		dlList := datastructs.DoublyLinkedList{}
		sourceCodeChangeInstance = &sourceCodeChangeInformation{
			sourceCodeChangeTracker: datastructs.RBTreeWithList[keyPair]{
				RbTree: &rbTree,
				DlList: &dlList,
			},
		}
	})

	return sourceCodeChangeInstance
}

func (scci *sourceCodeChangeInformation) ReportSourceCodeChange(startingIndex int, numCharsAdded int) {
	scci.sourceCodeChangeTracker.Insert(keyPair{startingIndex, 0}, spreadPair{numCharsAdded, 0}, traversal[keyPair])
}

func (scci *sourceCodeChangeInformation) NumToAddToSearch(oldSourceCodeIndex int) int {
	dllValue := scci.sourceCodeChangeTracker.FindBiggestSmallerOrEqual(keyPair{0, oldSourceCodeIndex}).(spreadPair)
	return dllValue.increasedSpread

}
