<!-- FSRS: due:2025-08-02T21:53:45-07:00, stability:3.17, difficulty:5.28, elapsed_days:0, scheduled_days:0, reps:1, lapses:0, state:Learning -->
## How do you declare a variable in Go? Feeeeeeeeeeen

---

## Variable Declaration

Using `var` keyword:
```go
var name string = "value"
var name = "value"  // type inference
var name string     // zero value
```

Short declaration (inside functions only):
```go
name := "value"
```

### Multiple variables:
```go
var a, b, c int
var x, y = 1, 2
```
