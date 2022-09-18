package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	datastructs "solidity-obfuscator/dataStructs"
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

func main() {

	jsonFile, errJson := os.Open("../contract_examples/contract_example_0813_2.sol_json.ast")
	defer jsonFile.Close()
	if errJson != nil {
		fmt.Println(errJson)
		return
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var jsonStringMap map[string]interface{}
	json.Unmarshal([]byte(byteValue), &jsonStringMap)

	sourceFile, errSource := os.Open("../contract_examples/contract_example_0813_2.sol")
	defer sourceFile.Close()
	if errSource != nil {
		fmt.Println(errSource)
		return
	}

	byteValue, _ = ioutil.ReadAll(sourceFile)
	sourceString := string(byteValue)

	//sourceString = ManipulateDefinedFunctionBodies()
	sourceString = ManipulateCalledFunctionsBodies()
	sourceString = ReplaceVarNames()
	sourceString = ReplaceComments()
	//sourceString = ReplaceLiterals()

	outputFile, errOutput := os.Create("../contract_examples/contract_example_0813_2_obf.sol")
	defer outputFile.Close()
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}

	outputFile.WriteString(sourceString)
	//generateTargetAST(12)

	// asd := datastructs.RBTree[int, int]{
	// 	Less: func(i1, i2 int) bool { return i1 < i2 },
	// }

	// rootNode := datastructs.RBNode[int, int]{
	// 	Data: 1,
	// 	Key:  1,
	// }
	// asd.Insert(&rootNode)

	// newNode := new(datastructs.RBNode[int, int])
	// newNode.Key = 7
	// //node15 := newNode
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 13
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 16
	// //node18 := newNode
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 21
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 24
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 16
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 28
	// node28 := newNode
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 5
	// asd.Insert(newNode)

	// fmt.Println(node15.Parent)
	// fmt.Println(node18.Parent)
	// fmt.Println(node28.Parent)
	// fmt.Println(newNode.Parent)

	// datastructs.InOrderTraversal(asd.Root)
	// fmt.Println()
	// datastructs.PreOrderTraversal(asd.Root)
	// fmt.Println()

	// rbTree := datastructs.RBTree[keyPair, datastructs.RBTreeData]{
	// 	Less: less,
	// }

	// dlList := datastructs.DoublyLinkedList{}

	// treeWithList := datastructs.RBTreeWithList[keyPair]{
	// 	RbTree: &rbTree,
	// 	DlList: &dlList,
	// }

	// treeWithList.Insert(keyPair{3, 0}, spreadPair{-1, 0}, traversal[keyPair])
	// treeWithList.Insert(keyPair{2, 0}, spreadPair{1, 0}, traversal[keyPair])
	// //treeWithList.Insert(keyPair{2, 0}, spreadPair{1, 0}, traversal[keyPair])
	// // treeWithList.Insert(keyPair{16, 0}, spreadPair{3, 0}, traversal[keyPair])
	// // treeWithList.Insert(keyPair{21, 0}, spreadPair{1, 0}, traversal[keyPair])

	// // treeWithList.Insert(keyPair{5, 0}, spreadPair{2, 0}, traversal[keyPair])
	// // treeWithList.PrintCurrentState()
	// // fmt.Println()
	// // treeWithList.Insert(keyPair{14, 0}, spreadPair{-2, 0}, traversal[keyPair])

	// treeWithList.PrintCurrentState()
}
