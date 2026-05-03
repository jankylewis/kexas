# 03 — Explicit types (no auto-inference)

## The rule

**Specify variable, parameter, and return types explicitly.**

Use `var x T = ...` everywhere it's syntactically possible. The `:=` short-declaration is allowed only in four constructs where `var` doesn't fit syntactically.

## The four allowed `:=` constructs

| Construct | Example | Why allowed |
|---|---|---|
| for-loop init | `for i := 0; i < n; i++` | for-loop init is a SimpleStmt; `var` is a Decl (syntax error) |
| range loop | `for _, x := range slice` | scope is loop-only; pre-declaring leaks scope |
| if-init | `if err := f(); err != nil { return err }` | scopes `err` to the if/else block |
| type switch | `switch v := x.(type) { ... }` | type binding requires `:=`; no syntactic alternative |

These four are the forms with **smaller block scope than `var`**, which is itself a fail-fast / readability win — you don't want a loop counter visible outside the loop, etc.

## Disallowed

```go
// WRONG — should be var
result := computeResult()

// WRONG
data := make(map[string]int)

// WRONG
err := someFunc()
if err != nil { return err }
```

## Correct (verbose) form

```go
var result *Response = computeResult()
var data map[string]int = make(map[string]int)
var err error = someFunc()
if err != nil { return err }
```

## Why we accept the verbosity

Two reasons stated by the project:

1. **Fail-fast culture.** Explicit types catch type errors at compile time more visibly. A reader inspecting one line can immediately see the type without having to mentally trace what `:=` resolves to.
2. **Faster readability.** When reading a variable, you understand its meaning instantly — you don't have to look up the function signature it came from to figure out the type.

## Trade-off acknowledged

This is the **least idiomatic-Go** rule in the project. Idiomatic Go strongly favors `:=`. New Go contributors will instinctively reach for `:=` and need correction.

We accept this cost for the fail-fast / readability benefits.

This rule **fights Rule 01** (230-line file cap) — explicit types add lines. When refactoring, plan for the line-count impact: a file at 200 lines using `:=` heavily may exceed 230 lines after rewriting with `var x T = ...`.

## Special cases

### Multi-return assignments

```go
// CORRECT — declare each variable separately
var result *Response
var err error
result, err = computeResult()

// NOT this:
result, err := computeResult()  // disallowed
```

### Comma-ok map lookups

```go
// CORRECT
var v SomeType
var ok bool
v, ok = m[key]

// NOT this:
v, ok := m[key]  // disallowed
```

### Type assertions (single value)

```go
// CORRECT
var page *kexas.Page = result.(*kexas.Page)

// NOT this:
page := result.(*kexas.Page)  // disallowed
```

### Function parameters and returns

Always explicit — Go enforces this anyway, no choice:

```go
func Launch(opts *LaunchOptions) (*Browser, error) { ... }
```

## Audit checklist

- [ ] No `:=` outside the four allowed constructs (for-init, range, if-init, type-switch)
- [ ] All `var` declarations include explicit type, even when inferable from RHS
- [ ] Multi-return assignments split into pre-declared `var`s
- [ ] Type assertions assigned to explicitly-typed `var`s
