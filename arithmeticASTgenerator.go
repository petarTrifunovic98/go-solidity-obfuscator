package main

import (
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

type IntegerASTNode interface {
	getValue() int
	toString() string
}

type OperandASTNode struct {
	operator string
	left     IntegerASTNode
	right    IntegerASTNode
}

type ValueASTNode struct {
	value int
}

func (node OperandASTNode) getValue() int {
	if node.left == nil && node.right != nil {
		return node.right.getValue()
	} else if node.right == nil && node.left != nil {
		return node.left.getValue()
	} else if node.right == nil && node.left == nil {
		return 0
	}

	if node.operator == "+" {
		ret := node.left.getValue() + node.right.getValue()
		return ret
	} else {
		ret := node.left.getValue() * node.right.getValue()
		return ret
	}
}

func (node OperandASTNode) toString() string {
	var ret string
	if node.left == nil || node.right == nil {
		ret = ""
	}

	ret = "(" + node.left.toString() + " " + node.operator + " " + node.right.toString() + ")"
	return ret
}

func (node ValueASTNode) getValue() int {
	return node.value
}

func (node ValueASTNode) toString() string {
	ret := strconv.Itoa(node.value)
	return ret
}

func generateTargetAST(target int) IntegerASTNode {

	randomSeeder := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSeeder)

	numOperands := randomGenerator.Intn(10) + 5

	numMultiplications := 0
	var operator string
	selector := randomGenerator.Float64()
	if selector <= 0.5 {
		operator = "+"
	} else {
		operator = "*"
		numMultiplications++
	}

	rootNode := OperandASTNode{
		operator: operator,
	}

	currentNode := &rootNode

	for i := 1; i < numOperands; i++ {
		valueNode := ValueASTNode{
			value: int(randomGenerator.Intn(16)),
		}

		currentNode.left = &valueNode

		if i < numOperands-1 {
			selector := randomGenerator.Float64()
			if selector <= 0.5 || numMultiplications >= numOperands/2 {
				operator = "+"
			} else {
				operator = "*"
				numMultiplications++
			}

			operandNode := OperandASTNode{
				operator: operator,
			}
			currentNode.right = &operandNode
			currentNode = &operandNode
		} else {
			valueNode = ValueASTNode{
				value: int(randomGenerator.Intn(20)),
			}
			currentNode.right = &valueNode
		}
	}

	currentNode = &rootNode
	currentValue := rootNode.getValue()
	newRootNode := OperandASTNode{
		operator: "+",
	}
	newRootNode.right = &rootNode

	targetValueNode := ValueASTNode{
		value: target - currentValue,
	}
	newRootNode.left = &targetValueNode

	return newRootNode
}

func getLiterals(jsonAST map[string]interface{}) []string {

	nodes := jsonAST["nodes"]
	literalsList := make([]string, 0)
	literalsList = storeLiterals(nodes, literalsList)
	return literalsList
}

func storeLiterals(node interface{}, literalsList []string) []string {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			literalsList = storeLiterals(element, literalsList)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "Literal" {
				if value, ok := nodeMap["value"]; ok {
					literalsList = append(literalsList, value.(string))
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					literalsList = storeLiterals(value, literalsList)
				}
			}
		}
	}

	return literalsList
}

func replaceLiterals(literalsList []string, sourceString string) string {

	for _, literal := range literalsList {
		re, _ := regexp.Compile("\\b" + literal + "\\b")
		intValue, _ := strconv.Atoi(literal)
		arithmeticExpr := generateTargetAST(intValue)
		sourceString = re.ReplaceAllString(sourceString, arithmeticExpr.toString())
	}

	return sourceString
}
