# trealla-go [![GoDoc](https://godoc.org/github.com/trealla-prolog/go?status.svg)](https://godoc.org/github.com/trealla-prolog/go)
`import "github.com/trealla-prolog/go/trealla"`

Prolog interface for Go using [Trealla Prolog](https://github.com/trealla-prolog/trealla) and [wazero](https://wazero.io/).
It's pretty fast. Not as fast as native Trealla, but pretty dang fast (about 2x slower than native).

**Development Status**: inching closer to stability

### Caveats

- Beta status, <s>API will probably change</s>
  - API is relatively stable now.

## Install

```bash
go get github.com/trealla-prolog/go
```

**Note**: the module is under `github.com/trealla-prolog/go`, **not** `[...]/go/trealla`.
go.dev is confused about this and will pull a very old version if you try to `go get` the `trealla` package.

## Usage

This library uses WebAssembly to run Trealla, executing Prolog queries in an isolated environment.

```go
import "github.com/trealla-prolog/go/trealla"

func main() {
	// load the interpreter and (optionally) grant access to the current directory
	pl, err := trealla.New(trealla.WithPreopenDir("."))
	if err != nil {
		panic(err)
	}
	// run a query; cancel context to abort it
	ctx := context.Background()
	query := pl.Query(ctx, "member(X, [1, foo(bar), c]).")

	// calling Close is not necessary if you iterate through the whole result set
	// but it doesn't hurt either
	defer query.Close()

	// iterate through answers
	for query.Next(ctx) {
		answer := query.Current()
		x := answer.Solution["X"]
		fmt.Println(x) // 1, trealla.Compound{Functor: "foo", Args: [trealla.Atom("bar")]}, "c"
	}

	// make sure to check the query for errors
	if err := query.Err(); err != nil {
		panic(err)
	}
}
```

### Single query

Use `QueryOnce` when you only want a single answer.

```go
pl := trealla.New()
answer, err := pl.QueryOnce(ctx, "succ(41, N).")
if err != nil {
	panic(err)
}

fmt.Println(answer.Stdout)
// Output: hello world
```

### Binding variables

You can bind variables in the query using the `WithBind` and `WithBinding` options.
This is a safe and convenient way to pass data into the query.
It is OK to pass these multiple times.

```go
pl := trealla.New()
answer, err := pl.QueryOnce(ctx, "write(X)", trealla.WithBind("X", trealla.Atom("hello world")))
if err != nil {
	panic(err)
}

fmt.Println(answer.Stdout)
// Output: hello world
```

### Scanning solutions

You can scan an answer's substitutions directly into a struct or map, similar to ichiban/prolog.

Use the `prolog:"VariableName"` struct tag to manually specify a variable name.
Otherwise, the field's name is used.

```go
answer, err := pl.QueryOnce(ctx, `X = 123, Y = abc, Z = ["hello", "world"].`)
if err != nil {
	panic(err)
}

var result struct {
	X  int
	Y  string
	Hi []string `prolog:"Z"`
}
// make sure to pass a pointer to the struct!
if err := answer.Solution.Scan(&result); err != nil {
	panic(err)
}

fmt.Printf("%+v", result)
// Output: {X:123 Y:abc Hi:[hello world]}
```

#### Struct compounds

Prolog compounds can destructure into Go structs. A special field of type `trealla.Functor` will be set to the functor.
The compound's arguments are matched with the exported struct fields in order.
These structs can also be used to bind variables in queries.

```prolog
?- findall(kv(Flag, Value), current_prolog_flag(Flag, Value), Flags).
   Flags = [kv(double_quotes,chars),kv(char_conversion,off),kv(occurs_check,false),kv(character_escapes,true),...]
```

```go
// You can embed trealla.Functor to represent Prolog compounds using Go structs.

// kv(Flag, Value)
type pair struct {
	trealla.Functor `prolog:"-/2"` // tag is optional, but can be used to specify the functor/arity
	Flag            trealla.Atom   // 1st arg
	Value           trealla.Term   // 2nd arg
}
var result struct {
	Flags []pair // Flags variable
}

ctx := context.Background()
pl, err := trealla.New()
if err != nil {
	panic(err)
}
answer, err := pl.QueryOnce(ctx, `
	findall(Flag-Value, (member(Flag, [double_quotes, encoding, max_arity]), current_prolog_flag(Flag, Value)), Flags).
`)
if err != nil {
	panic(err)
}
if err := answer.Solution.Scan(&result); err != nil {
	panic(err)
}
fmt.Printf("%v\n", result.Flags)
// Output: [{- double_quotes chars} {- encoding 'UTF-8'} {- max_arity 255}]
```

## Documentation

See **[package trealla's documentation](https://pkg.go.dev/github.com/trealla-prolog/go#section-directories)** for more details and examples.

## Builtins

These additional predicates are built in:

- `crypto_data_hash/3`
- `http_consult/1`
  - Argument can be URL string, or `my_module_name:"https://url.example"`

## WASM binary

This library embeds the Trealla WebAssembly binary in itself, so you can use it without any external dependencies.
The binaries are currently sourced from [guregu/trealla](https://github.com/guregu/trealla).

### Building the WASM binary

If you'd like to build `libtpl.wasm` yourself:

```bash
# The submodule in src/trealla points to the current version
git submodule update --init --recursive
# Use the Makefile in the root of this project
make clean wasm
# Run the tests and benchmark to make sure it works
go test -v ./trealla -bench=.
```

## Thanks

- Andrew Davison ([@infradig](https://github.com/infradig)) and other contributors to [Trealla Prolog](https://github.com/trealla-prolog/trealla).
- Nuno Cruces ([@ncruces](https://github.com/ncruces)) for the wazero port.
- Jos De Roo ([@josd](https://github.com/josd)) for test cases and encouragement.
- Aram Panasenco ([@panasenco](https://github.com/panasenco)) for his JSON library.

## License

MIT. See ATTRIBUTION as well.

## See also

- [trealla-js](https://github.com/guregu/trealla-js) is Trealla for Javascript.
- [ichiban/prolog](https://github.com/ichiban/prolog) is a pure Go Prolog.
- [guregu/pengine](https://github.com/guregu/pengine) is a Pengines (SWI-Prolog) library for Go.
