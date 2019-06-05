package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jakub-gawlas/expr/compiler"
	"github.com/jakub-gawlas/expr/parser"
	"github.com/jakub-gawlas/expr/vm"
	"github.com/sanity-io/litter"
)

var (
	bytecode bool
	debug    bool
	run      bool
	ast      bool
	dot      bool
	repl     bool
)

func init() {
	flag.BoolVar(&bytecode, "bytecode", false, "disassemble bytecode")
	flag.BoolVar(&debug, "debug", false, "debug program")
	flag.BoolVar(&run, "run", false, "run program")
	flag.BoolVar(&ast, "ast", false, "print ast")
	flag.BoolVar(&dot, "dot", false, "dot format")
	flag.BoolVar(&repl, "repl", false, "start repl")
}

func main() {
	flag.Parse()

	if ast {
		printAst()
		os.Exit(0)
	}
	if bytecode {
		printDisassemble()
		os.Exit(0)
	}
	if run {
		runProgram()
		os.Exit(0)
	}
	if debug {
		debugger()
		os.Exit(0)
	}
	if repl {
		startRepl()
		os.Exit(0)
	}

	flag.Usage()
	os.Exit(2)
}

func input() string {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func check(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func printAst() {
	tree, err := parser.Parse(input())
	check(err)
	if !dot {
		litter.Dump(tree.Node)
		return
	}

	dotAst(tree.Node)
}

func printDisassemble() {
	node, err := parser.Parse(input())
	check(err)

	program, err := compiler.Compile(node)
	check(err)

	_, _ = fmt.Fprintf(os.Stdout, program.Disassemble())
}

func runProgram() {
	tree, err := parser.Parse(input())
	check(err)

	program, err := compiler.Compile(tree)
	check(err)

	out, err := vm.Run(program, nil, nil)
	check(err)

	litter.Dump(out)
}

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	prompt()

	for scanner.Scan() {
		line := scanner.Text()

		tree, err := parser.Parse(line)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		program, err := compiler.Compile(tree)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		out, err := vm.Run(program, nil, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		fmt.Printf("%v\n", litter.Sdump(out))
		prompt()
	}
}

func prompt() {
	fmt.Print("> ")
}
