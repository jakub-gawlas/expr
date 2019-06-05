package compiler_test

import (
	"github.com/jakub-gawlas/expr/compiler"
	"github.com/jakub-gawlas/expr/parser"
	"github.com/jakub-gawlas/expr/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestCompile_debug(t *testing.T) {
	input := `false && true && true`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	_, err = compiler.Compile(tree)
	require.NoError(t, err)
}

func TestCompile(t *testing.T) {
	type test struct {
		input   string
		program vm.Program
	}
	var tests = []test{
		{
			`1`,
			vm.Program{
				Bytecode: []byte{
					vm.OpPush, 1, 0,
				},
			},
		},
		{
			`65535`,
			vm.Program{
				Bytecode: []byte{
					vm.OpPush, 255, 255,
				},
			},
		},
		{
			`65536`,
			vm.Program{
				Constants: []interface{}{
					int64(math.MaxUint16 + 1),
				},
				Bytecode: []byte{
					vm.OpConst, 0, 0,
				},
			},
		},
		{
			`.5`,
			vm.Program{
				Constants: []interface{}{
					float64(.5),
				},
				Bytecode: []byte{
					vm.OpConst, 0, 0,
				},
			},
		},
		{
			`true`,
			vm.Program{
				Bytecode: []byte{
					vm.OpTrue,
				},
			},
		},
		{
			`Name`,
			vm.Program{
				Constants: []interface{}{
					"Name",
				},
				Bytecode: []byte{
					vm.OpFetch, 0, 0,
				},
			},
		},
		{
			`"string"`,
			vm.Program{
				Constants: []interface{}{
					"string",
				},
				Bytecode: []byte{
					vm.OpConst, 0, 0,
				},
			},
		},
		{
			`"string" == "string"`,
			vm.Program{
				Constants: []interface{}{
					"string",
				},
				Bytecode: []byte{
					vm.OpConst, 0, 0,
					vm.OpConst, 0, 0,
					vm.OpEqual,
				},
			},
		},
		{
			`1000000 == 1000000`,
			vm.Program{
				Constants: []interface{}{
					int64(1000000),
				},
				Bytecode: []byte{
					vm.OpConst, 0, 0,
					vm.OpConst, 0, 0,
					vm.OpEqual,
				},
			},
		},
		{
			`-1`,
			vm.Program{
				Bytecode: []byte{
					vm.OpPush, 1, 0,
					vm.OpNegate,
				},
			},
		},
		{
			`true && true || true`,
			vm.Program{
				Bytecode: []byte{
					vm.OpTrue,
					vm.OpJumpIfFalse, 2, 0,
					vm.OpPop,
					vm.OpTrue,
					vm.OpJumpIfTrue, 2, 0,
					vm.OpPop,
					vm.OpTrue,
				},
			},
		},
	}

	for _, test := range tests {
		node, err := parser.Parse(test.input)
		require.NoError(t, err)

		program, err := compiler.Compile(node)
		require.NoError(t, err, test.input)

		assert.Equal(t, test.program.Disassemble(), program.Disassemble(), test.input)
	}
}
