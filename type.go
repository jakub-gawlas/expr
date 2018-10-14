package expr

import (
	"fmt"
	"reflect"
)

// Type is a reflect.Type alias.
type Type = reflect.Type

type typesTable map[string]Type

var (
	nilType       = reflect.TypeOf(nil)
	boolType      = reflect.TypeOf(true)
	numberType    = reflect.TypeOf(float64(0))
	textType      = reflect.TypeOf("")
	arrayType     = reflect.TypeOf([]interface{}{})
	mapType       = reflect.TypeOf(map[interface{}]interface{}{})
	interfaceType = reflect.TypeOf(new(interface{})).Elem()
)

type typed interface {
	Type(*parser) (Type, error)
}

func (p *parser) Type(node *Node) (Type, error) {
	ntype, err := (*node).(typed).Type(p)
	if err != nil {
		return nil, err
	}

	// Replace generated nodes.
	switch (*node).(type) {
	case *nameNode:
		genNode, ok := p.nameNodes[(*node).(*nameNode).name]
		if ok {
			*node = genNode
		}
	}

	return ntype, nil
}

func (n *nilNode) Type(p *parser) (Type, error) {
	return nil, nil
}

func (n *identifierNode) Type(p *parser) (Type, error) {
	return textType, nil
}

func (n *numberNode) Type(p *parser) (Type, error) {
	return numberType, nil
}

func (n *boolNode) Type(p *parser) (Type, error) {
	return boolType, nil
}

func (n *textNode) Type(p *parser) (Type, error) {
	return textType, nil
}

func (n *nameNode) Type(p *parser) (Type, error) {
	if t, ok := p.types[n.name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("unknown name %v", n)
}

func (n *unaryNode) Type(p *parser) (Type, error) {
	ntype, err := p.Type(&n.node)
	if err != nil {
		return nil, err
	}

	switch n.operator {
	case "!", "not":
		if isBoolType(ntype) || isInterfaceType(ntype) {
			return boolType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched type %v)`, n, ntype)
	}

	return interfaceType, nil
}

func (n *binaryNode) Type(p *parser) (Type, error) {
	var err error
	ltype, err := p.Type(&n.left)
	if err != nil {
		return nil, err
	}
	rtype, err := p.Type(&n.right)
	if err != nil {
		return nil, err
	}
	switch n.operator {
	case "==", "!=":
		if isComparable(ltype, rtype) {
			return boolType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	case "or", "||", "and", "&&":
		if (isBoolType(ltype) || isInterfaceType(ltype)) && (isBoolType(rtype) || isInterfaceType(rtype)) {
			return boolType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	case "in", "not in":
		if (isStringType(ltype) || isInterfaceType(ltype)) && (isStructType(rtype) || isInterfaceType(rtype)) {
			return boolType, nil
		}
		if isArrayType(rtype) || isMapType(rtype) || isInterfaceType(rtype) {
			return boolType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	case "<", ">", ">=", "<=":
		if (isNumberType(ltype) || isInterfaceType(ltype)) && (isNumberType(rtype) || isInterfaceType(rtype)) {
			return boolType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	case "/", "+", "-", "*", "**", "|", "^", "&", "%":
		if (isNumberType(ltype) || isInterfaceType(ltype)) && (isNumberType(rtype) || isInterfaceType(rtype)) {
			return numberType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	case "..":
		if (isNumberType(ltype) || isInterfaceType(ltype)) && (isNumberType(rtype) || isInterfaceType(rtype)) {
			return arrayType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	case "~":
		if (isStringType(ltype) || isInterfaceType(ltype)) && (isStringType(rtype) || isInterfaceType(rtype)) {
			return textType, nil
		}
		return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)

	}

	return interfaceType, nil
}

func (n *matchesNode) Type(p *parser) (Type, error) {
	var err error
	ltype, err := p.Type(&n.left)
	if err != nil {
		return nil, err
	}
	rtype, err := p.Type(&n.right)
	if err != nil {
		return nil, err
	}
	if (isStringType(ltype) || isInterfaceType(ltype)) && (isStringType(rtype) || isInterfaceType(rtype)) {
		return boolType, nil
	}
	return nil, fmt.Errorf(`invalid operation: %v (mismatched types %v and %v)`, n, ltype, rtype)
}

func (n *propertyNode) Type(p *parser) (Type, error) {
	ntype, err := p.Type(&n.node)
	if err != nil {
		return nil, err
	}
	if t, ok := fieldType(ntype, n.property); ok {
		return t, nil
	}
	return nil, fmt.Errorf("%v undefined (type %v has no field %v)", n, ntype, n.property)
}

func (n *indexNode) Type(p *parser) (Type, error) {
	ntype, err := p.Type(&n.node)
	if err != nil {
		return nil, err
	}
	_, err = p.Type(&n.index)
	if err != nil {
		return nil, err
	}
	if t, ok := indexType(ntype); ok {
		return t, nil
	}
	return nil, fmt.Errorf("invalid operation: %v (type %v does not support indexing)", n, ntype)
}

func (n *methodNode) Type(p *parser) (Type, error) {
	ntype, err := p.Type(&n.node)
	if err != nil {
		return nil, err
	}
	for _, node := range n.arguments {
		_, err := p.Type(&node)
		if err != nil {
			return nil, err
		}
	}
	if t, ok := methodType(ntype, n.method); ok {
		if f, ok := funcType(t); ok {
			return f, nil
		}
	}

	return nil, fmt.Errorf("%v undefined (type %v has no method %v)", n, ntype, n.method)
}

func (n *builtinNode) Type(p *parser) (Type, error) {
	for _, node := range n.arguments {
		_, err := p.Type(&node)
		if err != nil {
			return nil, err
		}
	}
	switch n.name {
	case "len":
		// TODO: Add arguments type checks.
		return numberType, nil
	}
	return nil, fmt.Errorf("%v undefined", n)
}

func (n *functionNode) Type(p *parser) (Type, error) {
	for _, node := range n.arguments {
		_, err := p.Type(&node)
		if err != nil {
			return nil, err
		}
	}
	if t, ok := p.types[n.name]; ok {
		if f, ok := funcType(t); ok {
			return f, nil
		}
	}
	return nil, fmt.Errorf("unknown func %v", n)
}

func (n *conditionalNode) Type(p *parser) (Type, error) {
	ctype, err := p.Type(&n.cond)
	if err != nil {
		return nil, err
	}
	if !isBoolType(ctype) && !isInterfaceType(ctype) {
		return nil, fmt.Errorf("non-bool %v (type %v) used as condition", n.cond, ctype)
	}

	t1, err := p.Type(&n.exp1)
	if err != nil {
		return nil, err
	}
	t2, err := p.Type(&n.exp2)
	if err != nil {
		return nil, err
	}

	if t1 == nil && t2 != nil {
		return t2, nil
	}
	if t1 != nil && t2 == nil {
		return t1, nil
	}
	if t1 == nil && t2 == nil {
		return nilType, nil
	}
	if t1.AssignableTo(t2) {
		return t1, nil
	}
	return interfaceType, nil
}

func (n *arrayNode) Type(p *parser) (Type, error) {
	for _, node := range n.nodes {
		_, err := p.Type(&node)
		if err != nil {
			return nil, err
		}
	}
	return arrayType, nil
}

func (n *mapNode) Type(p *parser) (Type, error) {
	for _, node := range n.pairs {
		_, err := node.Type(p)
		if err != nil {
			return nil, err
		}
	}
	return mapType, nil
}

func (n *pairNode) Type(p *parser) (Type, error) {
	var err error
	_, err = p.Type(&n.key)
	if err != nil {
		return nil, err
	}
	_, err = p.Type(&n.value)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// helper funcs for reflect

func isComparable(l Type, r Type) bool {
	l = dereference(l)
	r = dereference(r)

	if l == nil || r == nil {
		return true // It is possible to compare with nil.
	}

	if isNumberType(l) && isNumberType(r) {
		return true
	} else if l.Kind() == reflect.Interface {
		return true
	} else if r.Kind() == reflect.Interface {
		return true
	} else if l == r {
		return true
	}
	return false
}

func isInterfaceType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.Interface:
			return true
		}
	}
	return false
}

func isNumberType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.Float32, reflect.Float64:
			fallthrough
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fallthrough
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return true
		}
	}
	return false
}

func isBoolType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.Bool:
			return true
		}
	}
	return false
}

func isStringType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.String:
			return true
		}
	}
	return false
}

func isArrayType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			return true
		}
	}
	return false
}

func isMapType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.Map:
			return true
		}
	}
	return false
}

func isStructType(t Type) bool {
	t = dereference(t)
	if t != nil {
		switch t.Kind() {
		case reflect.Struct:
			return true
		}
	}
	return false
}

func fieldType(ntype Type, name string) (Type, bool) {
	ntype = dereference(ntype)
	if ntype != nil {
		switch ntype.Kind() {
		case reflect.Interface:
			return interfaceType, true
		case reflect.Struct:
			// First check all struct's fields.
			for i := 0; i < ntype.NumField(); i++ {
				f := ntype.Field(i)
				if !f.Anonymous && f.Name == name {
					return f.Type, true
				}
			}

			// Second check fields of embedded structs.
			for i := 0; i < ntype.NumField(); i++ {
				f := ntype.Field(i)
				if f.Anonymous {
					if t, ok := fieldType(f.Type, name); ok {
						return t, true
					}
				}
			}
		case reflect.Map:
			return ntype.Elem(), true
		}
	}

	return nil, false
}

func methodType(ntype Type, name string) (Type, bool) {
	ntype = dereference(ntype)
	if ntype != nil {
		switch ntype.Kind() {
		case reflect.Interface:
			return interfaceType, true
		case reflect.Struct:
			// First check all struct's methods.
			for i := 0; i < ntype.NumMethod(); i++ {
				m := ntype.Method(i)
				if m.Name == name {
					return m.Type, true
				}
			}

			// Second check all struct's fields.
			for i := 0; i < ntype.NumField(); i++ {
				f := ntype.Field(i)
				if !f.Anonymous && f.Name == name {
					return f.Type, true
				}
			}

			// Third check fields of embedded structs.
			for i := 0; i < ntype.NumField(); i++ {
				f := ntype.Field(i)
				if f.Anonymous {
					if t, ok := methodType(f.Type, name); ok {
						return t, true
					}
				}
			}
		case reflect.Map:
			return ntype.Elem(), true
		}
	}

	return nil, false
}

func indexType(ntype Type) (Type, bool) {
	ntype = dereference(ntype)
	if ntype == nil {
		return nil, false
	}

	switch ntype.Kind() {
	case reflect.Interface:
		return interfaceType, true
	case reflect.Map, reflect.Array, reflect.Slice:
		return ntype.Elem(), true
	}

	return nil, false
}

func funcType(ntype Type) (Type, bool) {
	ntype = dereference(ntype)
	if ntype == nil {
		return nil, false
	}

	switch ntype.Kind() {
	case reflect.Interface:
		return interfaceType, true
	case reflect.Func:
		if ntype.NumOut() > 0 {
			return ntype.Out(0), true
		}
		return nilType, true
	}

	return nil, false
}

func dereference(ntype Type) Type {
	if ntype == nil {
		return nil
	}
	if ntype.Kind() == reflect.Ptr {
		ntype = dereference(ntype.Elem())
	}
	return ntype
}
