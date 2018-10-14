package expr_test

import (
	"testing"

	"github.com/antonmedv/expr"
)

type segment struct {
	Origin string
}
type passengers struct {
	Adults int
}
type request struct {
	Segments   []*segment
	Passengers *passengers
	Marker     string
}

type NameNode_request_Segments struct{}

func (n NameNode_request_Segments) Eval(env interface{}) (interface{}, error) {
	return env.(*request).Segments, nil
}

type NameNode_request_Passengers struct{}

func (n NameNode_request_Passengers) Eval(env interface{}) (interface{}, error) {
	return env.(*request).Passengers, nil
}

type NameNode_request_Marker struct{}

func (n NameNode_request_Marker) Eval(env interface{}) (interface{}, error) {
	return env.(*request).Marker, nil
}

var nodes = map[string]expr.Node{
	"Segments":   NameNode_request_Segments{},
	"Passengers": NameNode_request_Passengers{},
	"Marker":     NameNode_request_Marker{},
}

func Benchmark_expr(b *testing.B) {
	r := &request{
		Segments: []*segment{
			{Origin: "MOW"},
		},
		Passengers: &passengers{
			Adults: 2,
		},
		Marker: "test",
	}

	code := `Segments[0].Origin == "MOW" && Passengers.Adults == 2 && Marker == "test"`
	p, err := expr.Parse(code, expr.Env(&request{}), expr.Gen(nodes))
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		expr.Run(p, r)
	}
}
