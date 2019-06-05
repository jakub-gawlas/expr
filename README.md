<p align="center"><img src="docs/images/logo-small.png" width="150" alt="expr logo"></p>

# Expr 
[![Build Status](https://travis-ci.org/jakub-gawlas/expr.svg?branch=master)](https://travis-ci.org/jakub-gawlas/expr) 
[![Go Report Card](https://goreportcard.com/badge/github.com/jakub-gawlas/expr)](https://goreportcard.com/report/github.com/jakub-gawlas/expr) 
[![Code Coverage](https://scrutinizer-ci.com/g/jakub-gawlas/expr/badges/coverage.png?b=master)](https://scrutinizer-ci.com/g/jakub-gawlas/expr/?branch=master) 
[![GoDoc](https://godoc.org/github.com/jakub-gawlas/expr?status.svg)](https://godoc.org/github.com/jakub-gawlas/expr)

**Expr** package provides an engine that can compile and evaluate expressions. 
An expression is a one-liner that returns a value (mostly, but not limited to, booleans).
It is designed for simplicity, speed and safety.

The purpose of the package is to allow users to use expressions inside configuration for more complex logic. 
It is a perfect candidate for the foundation of a _business rule engine_. 
The idea is to let configure things in a dynamic way without recompile of a program:

```coffeescript
# Get the special price if
user.Group in ["good_customers", "collaborator"]

# Promote article to the homepage when
len(article.Comments) > 100 and article.Category not in ["misc"]

# Send an alert when
product.Stock < 15
```

## Features

* Seamless integration with Go.
* Static typing ([example](https://godoc.org/github.com/jakub-gawlas/expr#example-Env)).
  ```go
  out, err := expr.Eval("'hello' + 10")
  // err: invalid operation + (mismatched types string and int64)
  // | 'hello' + 10
  // | ........^
  ```
* User-friendly error messages.
* Reasonable set of basic operators.
* Builtins `all`, `none`, `any`, `one`, `filter`, `map`.
  ```coffeescript
  all(Tweets, {.Size < 140})
  ```
* Fast ([benchmarks](https://github.com/jakub-gawlas/golang-expression-evaluation-comparison)).

## Install

```
go get github.com/jakub-gawlas/expr
```

<a href="https://www.patreon.com/jakub-gawlas">
	<img src="https://c5.patreon.com/external/logo/become_a_patron_button@2x.png" width="160">
</a>

## Documentation

* See [docs](docs) page for developer documentation.
* See [The Expression Syntax](docs/The-Expression-Syntax.md) page to learn the syntax.

## Examples

Executing arbitrary expressions.

```go
env := map[string]interface{}{
    "foo": 1,
    "bar": struct{Value int}{1},
}

out, err := expr.Eval("foo + bar.Value", env)
```

Static type checker with struct as environment.

```go
type Env struct {
	Foo int
	Bar *Bar
}

type Bar struct {
	Value int
}

program, err := expr.Compile("Foo + Bar.Value", expr.Env(&Env{}))

out, err := expr.Run(program, &Env{1, &Bar{2}})
```

Using env's methods as functions inside expressions.

```go
type Env struct {
	Name string
}

func (e *Env) Title() string {
	return strings.Title(e.Name)
}

program, err := expr.Compile(`"Hello " + Title()`, expr.Env(&Env{}))

out, err := expr.Run(program, &Env{"world"})
```

## Contributing

**Expr** consist of a few packages for parsing source code to AST, type checking AST, compiling to bytecode and VM for running bytecode program.

Also expr provides powerful tool [exp](cmd/exp) for debugging. It has interactive terminal debugger for our bytecode virtual machine.

<p align="center">
    <img src="docs/images/debug.gif" alt="debugger" width="605">
</p>
    

## Who is using Expr?

* [Aviasales](https://aviasales.ru) are actively using Expr for different parts of the search engine.

## License

[MIT](LICENSE)
