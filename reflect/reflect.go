package reflect

import (
	"container/list"
	"fmt"

	"unsafe"
)

type Operator uint16
type Operand uint16

/***
OP_add 加
OP_sub 减
OP_mul 乘
OP_div 除
 */
const (
	OP_add Operator = 1
	OP_sub Operator = 2
	OP_mul Operator = 3
	OP_div Operator = 4
)

/***
  基础元素
	Type 操作数类型
	Operand 值
 */
type Element struct {
	Type    Operator
	Operand int64
}

type Stack struct {
	list *list.List
}

func NewStack() *Stack {
	list := list.New()
	return &Stack{list}
}

func (stack *Stack) Push(value interface{}) {
	stack.list.PushBack(value)
}

func (stack *Stack) Pop() interface{} {
	e := stack.list.Back()
	if e != nil {
		stack.list.Remove(e)
		return e.Value
	}
	return nil
}

func (stack *Stack) Peak() interface{} {
	e := stack.list.Back()
	if e != nil {
		return e.Value
	}

	return nil
}

func (stack *Stack) Len() int {
	return stack.list.Len()
}

func (stack *Stack) Empty() bool {
	return stack.list.Len() == 0
}

type tflag uint8
type nameOff int32 // offset to a name
type typeOff int32 // offset to an *rtype
type textOff int32 // offset from top of text section
// a copy of runtime.typeAlg
type typeAlg struct {
	// function for hashing objects of this type
	// (ptr to object, seed) -> hash
	hash func(unsafe.Pointer, uintptr) uintptr
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer) bool
}
type rtype struct {
	size       uintptr
	ptrdata    uintptr  // number of bytes in the type that can contain pointers
	hash       uint32   // hash of type; avoids computation in hash tables
	tflag      tflag    // extra type information flags
	align      uint8    // alignment of variable with this type
	fieldAlign uint8    // alignment of struct field with this type
	kind       uint8    // enumeration for C
	alg        *typeAlg // algorithm table
	gcdata     *byte    // garbage collection data
	str        nameOff  // string form
	ptrToThis  typeOff  // type for pointer to this type, may be zero
}

// emptyInterface is the header for an interface{} value.
type EmptyInterface struct {
	typ  *rtype
	word unsafe.Pointer
}




func main() {
	stack := NewStack()
	ele := Element{Type: OP_add, Operand: 102}

	for i := 0; i < 10; i++ {
		var ii Operator = OP_add
		stack.Push(ii)
		var iii Operand = (Operand)(i)
		stack.Push(iii)
	}

	ele.Type = OP_mul
	var e interface{}
	e = stack.Pop()

	for ; e != nil; e = stack.Pop() {
		eface := *(*EmptyInterface)(unsafe.Pointer(&e))

		fmt.Printf("%d => %T \t %d \n", e, e, eface.typ.str)
		//el, ok := e.(Operand)
		//if !ok {
		//	continue
		//}
		//fmt.Println(el)

	}

}
