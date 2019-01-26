package main

import (
	"fmt"
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
	OP_ADD Operator = '+'
	OP_SUB Operator = '-'
	OP_MUL Operator = '*'
	OP_DIV Operator = '/'
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

	stack := NewStack()
	var input = []byte("a+b*c+(d*e+f)*g")

	for _, ele := range input {
		if curOperator, ok := OperatorMap[(Operator)(ele)]; ok {
			//操作符号
			var topOpcode Operator
			var ok bool
			topOpcode, ok = stack.Peak().(Operator)
			if !ok {
				stack.Push(curOperator)
				continue
			}
			switch curOperator {
			case ')':
				for stack.Peak() != nil && curOperator != '(' {
					stack.Pop()
					fmt.Printf("%c", curOperator)
					if topOpcode, ok = stack.Peak().(Operator); !ok {
						break
					}
				}

				if curOperator == '(' {
					stack.Pop()
				} else {
					//fmt.Printf("表达式有问题 %c\n", curOperator)
				}
				break

			default:
				for stack.Peak() != nil && OperatorLvl[curOperator] >= OperatorLvl[topOpcode] {
					if topOpcode == '(' {
						break
					}

					stack.Pop()
					fmt.Printf("%c", topOpcode)
					if topOpcode, ok = stack.Peak().(Operator); !ok {
						break
					}
				}
				//fmt.Printf("入栈 %c\n", curOperator)
				stack.Push(curOperator)
			}

		} else {
			//fmt.Printf("Operand %c \n", ele)
			fmt.Printf("%c", ele)
		}
	}

	var e interface{}
	e = stack.Pop()
	fmt.Println("\n stack:")
	for ; e != nil; e = stack.Pop() {
		el, ok := e.(Operator)
		if !ok {
			continue
		}
		fmt.Printf("Operand %c \n", el)
	}

}
