package main

import (
	datastructs "solidity-obfuscator/dataStructs"
)

type keyPair struct {
	currentLine int
	reducedLine int
}

type spreadPair struct {
	realSpread      int
	increasedSpread int
}

func less(k1, k2 keyPair) bool {
	if k1.currentLine < 0 {
		return k1.reducedLine < k2.reducedLine
	} else {
		return k1.currentLine < k2.currentLine
	}
}

func traversal(node *datastructs.DLLNode) {
	prev := node.GetPrevious()
	if prev != nil {
		prevNodeValue := prev.GetValue().(spreadPair)
		nodeValue := node.GetValue().(spreadPair)
		newValue := spreadPair{
			realSpread:      nodeValue.realSpread,
			increasedSpread: nodeValue.realSpread + prevNodeValue.increasedSpread,
		}
		node.SetValue(newValue)
	}
}

// func updateKey(k keyPair, nodeData datastructs.RBTreeData) keyPair {

// }

func main() {

	// jsonFile, errJson := os.Open("../contract_examples/contract_example_0813_2.sol_json.ast")
	// defer jsonFile.Close()
	// if errJson != nil {
	// 	fmt.Println(errJson)
	// 	return
	// }

	// byteValue, _ := ioutil.ReadAll(jsonFile)
	// var jsonStringMap map[string]interface{}
	// json.Unmarshal([]byte(byteValue), &jsonStringMap)

	// sourceFile, errSource := os.Open("../contract_examples/contract_example_0813_2.sol")
	// defer sourceFile.Close()
	// if errSource != nil {
	// 	fmt.Println(errSource)
	// 	return
	// }

	// byteValue, _ = ioutil.ReadAll(sourceFile)
	// sourceString := string(byteValue)

	// sourceString = ManipulateCalledFunctionsBodies()
	// sourceString = ReplaceVarNames()
	// sourceString = ReplaceComments()
	// sourceString = ReplaceLiterals()

	// outputFile, errOutput := os.Create("../contract_examples/obfuscated.sol")
	// defer outputFile.Close()
	// if errOutput != nil {
	// 	fmt.Println(errOutput)
	// 	return
	// }

	// outputFile.WriteString(sourceString)
	//generateTargetAST(12)

	// asd := datastructs.RBTree[int, int]{
	// 	Less: func(i1, i2 int) bool { return i1 < i2 },
	// }

	// rootNode := datastructs.RBNode[int, int]{
	// 	Data: 1,
	// 	Key:  7,
	// }
	// asd.Insert(&rootNode)

	// newNode := new(datastructs.RBNode[int, int])
	// newNode.Key = 15
	// node15 := newNode
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 6
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 18
	// node18 := newNode
	// asd.Insert(newNode)

	// newNode = new(datastructs.RBNode[int, int])
	// newNode.Key = 12
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

	// treeWithList.Insert(keyPair{1, 1}, spreadPair{3, 3}, traversal)
	// treeWithList.Insert()

	// treeWithList.PrintCurrentState()
}
