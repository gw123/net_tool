package main

import (
	"fmt"
	"github.com/gw123/net_tool/worker/tree/stack"
)

type Operator uint16
type Operand float64

/***
OP_add 加
OP_sub 减
OP_mul 乘
OP_div 除
 */
const (
	OP_ADD      Operator = '+'
	OP_SUB      Operator = '-'
	OP_MUL      Operator = '*'
	OP_DIV      Operator = '/'
	OP_LBRACKET Operator = '('
	OP_RBRACKET Operator = ')'
)

var OperatorMap map[Operator]Operator
var OperatorLvl map[Operator]uint8

func init() {
	OperatorMap = map[Operator]Operator{
		'+': '+',
		'-': '-',
		'*': '*',
		'/': '/',
		'%': '%',
		'(': '(',
		')': ')',
		'{': '{',
		'}': '}',
		'[': '[',
		']': ']',
	}

	OperatorLvl = map[Operator]uint8{
		'+': 1,
		'-': 1,
		'*': 2,
		'/': 2,
		'%': 2,
		'(': 3,
		')': 3,
		'{': 3,
		'}': 3,
		'[': 3,
		']': 3,
	}
}

func main() {

	stack := stack.NewStack()
	var input = []byte("a+b*c+(d*e+f)*g")
	var result = make([]byte, 0)
	for _, ele := range input {
		if curOperator, ok := OperatorMap[(Operator)(ele)]; ok {
			//操作符号
			var topOpcode Operator
			var ok bool
			topOpcode, ok = stack.Peak().(Operator)
			if !ok {
				//fmt.Printf("topOpcode %c", topOpcode)
				stack.Push(curOperator)
				continue
			}

			switch curOperator {
			case OP_RBRACKET:
				for topOpcode != OP_LBRACKET {
					stack.Pop()
					result = append(result, byte(topOpcode))
					//fmt.Printf("%s , %c\n", string(result), byte(topOpcode))
					if topOpcode, ok = stack.Peak().(Operator); !ok {
						break
					}
				}

				if curOperator == OP_LBRACKET {
					stack.Pop()
				} else {
					//fmt.Printf("表达式有问题 %c\n", curOperator)
				}
				break

			default:
				for OperatorLvl[curOperator] <= OperatorLvl[topOpcode] {
					if topOpcode == OP_LBRACKET {
						break
					}
					stack.Pop()
					result = append(result, byte(topOpcode))
					//fmt.Printf("%s , %c\n", string(result), byte(topOpcode))
					if topOpcode, ok = stack.Peak().(Operator); !ok {
						break
					}
				}
				//fmt.Printf("入栈 %c\n", curOperator)
				stack.Push(curOperator)
			}
		} else {
			result = append(result, ele)
			//fmt.Printf("Operand %c \n", ele)
		}
	}

	var e interface{}
	e = stack.Pop()

	for ; e != nil; e = stack.Pop() {
		el, ok := e.(Operator)
		if !ok {
			continue
		}
		if el == OP_LBRACKET {
			continue
		}
		result = append(result, byte(el))
	}

	fmt.Println("\n stack:", string(result))

}
