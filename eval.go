package expr

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
)

// Eval parses and evaluates given input.
func Eval(input string, env interface{}) (interface{}, error) {
	node, err := Parse(input)
	if err != nil {
		return nil, err
	}
	return Run(node, env)
}

// Run evaluates given ast.
func Run(node Node, env interface{}) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	v, err := node.Eval(env)
	if err != nil {
		return nil, err
	}
	if v.IsValid() && v.CanInterface() {
		return v.Interface(), nil
	}
	return nil, nil
}

// eval functions

var null = reflect.ValueOf(nil)

func (n nilNode) Eval(env interface{}) (reflect.Value, error) {
	return null, nil
}

func (n identifierNode) Eval(env interface{}) (reflect.Value, error) {
	return reflect.ValueOf(n.value), nil
}

func (n numberNode) Eval(env interface{}) (reflect.Value, error) {
	return reflect.ValueOf(n.value), nil
}

func (n boolNode) Eval(env interface{}) (reflect.Value, error) {
	return reflect.ValueOf(n.value), nil
}

func (n textNode) Eval(env interface{}) (reflect.Value, error) {
	return reflect.ValueOf(n.value), nil
}

func (n nameNode) Eval(env interface{}) (reflect.Value, error) {
	v, ok := extract(reflect.ValueOf(env), reflect.ValueOf(n.name))
	if !ok {
		return null, fmt.Errorf("undefined: %v", n)
	}
	return v, nil
}

func (n unaryNode) Eval(env interface{}) (reflect.Value, error) {
	val, err := n.node.Eval(env)
	if err != nil {
		return null, err
	}

	switch n.operator {
	case "not", "!":
		return reflect.ValueOf(!toBool(n, val)), nil
	}

	v := toNumber(n, val)
	switch n.operator {
	case "-":
		return reflect.ValueOf(-v), nil
	case "+":
		return reflect.ValueOf(+v), nil
	}

	return null, fmt.Errorf("implement unary %q operator", n.operator)
}

func (n binaryNode) Eval(env interface{}) (reflect.Value, error) {
	left, err := n.left.Eval(env)
	if err != nil {
		return null, err
	}

	switch n.operator {
	case "or", "||":
		if toBool(n.left, left) {
			return reflect.ValueOf(true), nil
		}
		right, err := n.right.Eval(env)
		if err != nil {
			return null, err
		}
		return reflect.ValueOf(toBool(n.right, right)), nil

	case "and", "&&":
		if toBool(n.left, left) {
			right, err := n.right.Eval(env)
			if err != nil {
				return null, err
			}
			return reflect.ValueOf(toBool(n.right, right)), nil
		}
		return reflect.ValueOf(false), nil
	}

	right, err := n.right.Eval(env)
	if err != nil {
		return null, err
	}

	switch n.operator {
	case "==":
		return reflect.ValueOf(equal(left, right)), nil

	case "!=":
		return reflect.ValueOf(!equal(left, right)), nil

	case "in":
		ok, err := contains(left, right)
		if err != nil {
			return null, err
		}
		return reflect.ValueOf(ok), nil

	case "not in":
		ok, err := contains(left, right)
		if err != nil {
			return null, err
		}
		return reflect.ValueOf(!ok), nil

	case "~":
		return reflect.ValueOf(toText(n.left, left) + toText(n.right, right)), nil
	}

	// Next goes operators on numbers

	l, r := toNumber(n.left, left), toNumber(n.right, right)

	switch n.operator {
	case "|":
		return reflect.ValueOf(int(l) | int(r)), nil

	case "^":
		return reflect.ValueOf(int(l) ^ int(r)), nil

	case "&":
		return reflect.ValueOf(int(l) & int(r)), nil

	case "<":
		return reflect.ValueOf(l < r), nil

	case ">":
		return reflect.ValueOf(l > r), nil

	case ">=":
		return reflect.ValueOf(l >= r), nil

	case "<=":
		return reflect.ValueOf(l <= r), nil

	case "+":
		return reflect.ValueOf(l + r), nil

	case "-":
		return reflect.ValueOf(l - r), nil

	case "*":
		return reflect.ValueOf(l * r), nil

	case "/":
		div := r
		if div == 0 {
			return null, fmt.Errorf("division by zero")
		}
		return reflect.ValueOf(l / div), nil

	case "%":
		numerator := int64(l)
		denominator := int64(r)
		if denominator == 0 {
			return null, fmt.Errorf("division by zero")
		}
		return reflect.ValueOf(float64(numerator % denominator)), nil

	case "**":
		return reflect.ValueOf(math.Pow(l, r)), nil

	case "..":
		return makeRange(int64(l), int64(r))
	}

	return null, fmt.Errorf("implement %q operator", n.operator)
}

func makeRange(min, max int64) (reflect.Value, error) {
	size := max - min + 1
	if size > 1e6 {
		return null, fmt.Errorf("range %v..%v exceeded max size of 1e6", min, max)
	}
	a := make([]float64, size)
	for i := range a {
		a[i] = float64(min + int64(i))
	}
	return reflect.ValueOf(a), nil
}

func (n matchesNode) Eval(env interface{}) (reflect.Value, error) {
	left, err := n.left.Eval(env)
	if err != nil {
		return null, err
	}

	if n.r != nil {
		return reflect.ValueOf(n.r.MatchString(toText(n.left, left))), nil
	}

	right, err := n.right.Eval(env)
	if err != nil {
		return null, err
	}

	matched, err := regexp.MatchString(toText(n.right, right), toText(n.left, left))
	if err != nil {
		return null, err
	}
	return reflect.ValueOf(matched), nil
}

func (n propertyNode) Eval(env interface{}) (reflect.Value, error) {
	v, err := n.node.Eval(env)
	if err != nil {
		return null, err
	}
	p, ok := extract(v, reflect.ValueOf(n.property))
	if !ok {
		if isNil(v) {
			return null, fmt.Errorf("%v is nil", n.node)
		}
		return null, fmt.Errorf("%v undefined (type %T has no field %v)", n, v, n.property)
	}
	return p, nil
}

func (n indexNode) Eval(env interface{}) (reflect.Value, error) {
	v, err := n.node.Eval(env)
	if err != nil {
		return null, err
	}
	i, err := n.index.Eval(env)
	if err != nil {
		return null, err
	}
	p, ok := extract(v, i)
	if !ok {
		return null, fmt.Errorf("cannot get %q from %T: %v", i, v, n)
	}
	return p, nil
}

func (n methodNode) Eval(env interface{}) (reflect.Value, error) {
	v, err := n.node.Eval(env)
	if err != nil {
		return null, err
	}

	method, ok := getFunc(v, reflect.ValueOf(n.method))
	if !ok {
		return null, fmt.Errorf("cannot get method %v from %T: %v", n.method, v, n)
	}

	in := make([]reflect.Value, 0)

	for _, a := range n.arguments {
		i, err := a.Eval(env)
		if err != nil {
			return null, err
		}
		in = append(in, reflect.ValueOf(i))
	}

	out := reflect.ValueOf(method).Call(in)

	if len(out) == 0 {
		return null, nil
	} else if len(out) > 1 {
		return null, fmt.Errorf("method %q must return only one value", n.method)
	}

	return out[0], nil

	return null, nil
}

func (n builtinNode) Eval(env interface{}) (reflect.Value, error) {
	switch n.name {
	case "len":
		if len(n.arguments) == 0 {
			return null, fmt.Errorf("missing argument: %v", n)
		}
		if len(n.arguments) > 1 {
			return null, fmt.Errorf("too many arguments: %v", n)
		}

		i, err := n.arguments[0].Eval(env)
		if err != nil {
			return null, err
		}

		switch reflect.TypeOf(i).Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			return reflect.ValueOf(float64(reflect.ValueOf(i).Len())), nil
		}
		return null, fmt.Errorf("invalid argument %v (type %T)", n, i)
	}

	return null, fmt.Errorf("unknown %q builtin", n.name)
}

func (n functionNode) Eval(env interface{}) (reflect.Value, error) {
	fn, ok := getFunc(reflect.ValueOf(env), reflect.ValueOf(n.name))
	if !ok {
		return null, fmt.Errorf("undefined: %v", n.name)
	}

	in := make([]reflect.Value, 0)

	for _, a := range n.arguments {
		i, err := a.Eval(env)
		if err != nil {
			return null, err
		}
		in = append(in, i)
	}

	out := reflect.ValueOf(fn).Call(in)

	if len(out) == 0 {
		return null, nil
	} else if len(out) > 1 {
		return null, fmt.Errorf("func %q must return only one value", n.name)
	}

	return out[0], nil

	return null, nil
}

func (n conditionalNode) Eval(env interface{}) (reflect.Value, error) {
	cond, err := n.cond.Eval(env)
	if err != nil {
		return null, err
	}

	// If
	if toBool(n.cond, cond) {
		// Then
		a, err := n.exp1.Eval(env)
		if err != nil {
			return null, err
		}
		return a, nil
	}
	// Else
	b, err := n.exp2.Eval(env)
	if err != nil {
		return null, err
	}
	return b, nil

}

func (n arrayNode) Eval(env interface{}) (reflect.Value, error) {
	array := make([]interface{}, 0)
	for _, node := range n.nodes {
		val, err := node.Eval(env)
		if err != nil {
			return null, err
		}
		array = append(array, val)
	}
	return reflect.ValueOf(array), nil
}

func (n mapNode) Eval(env interface{}) (reflect.Value, error) {
	m := make(map[interface{}]interface{})
	for _, pair := range n.pairs {
		key, err := pair.key.Eval(env)
		if err != nil {
			return null, err
		}
		value, err := pair.value.Eval(env)
		if err != nil {
			return null, err
		}
		m[key] = value
	}
	return reflect.ValueOf(m), nil
}
