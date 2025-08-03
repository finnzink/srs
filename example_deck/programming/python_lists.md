<!-- FSRS: due:2025-08-18T21:43:52-07:00, stability:15.69, difficulty:3.22, elapsed_days:0, scheduled_days:16, reps:1, lapses:0, state:Review -->
# What are different ways to create lists in Python?

---

## List Creation Methods

### 1. Literal syntax:
```python
# Empty list
empty = []

# List with elements
numbers = [1, 2, 3, 4, 5]
mixed = [1, "hello", 3.14, True]
```

### 2. Using `list()` constructor:
```python
# From string
chars = list("hello")  # ['h', 'e', 'l', 'l', 'o']

# From range
numbers = list(range(5))  # [0, 1, 2, 3, 4]
```

### 3. List comprehensions:
```python
# Basic comprehension
squares = [x**2 for x in range(5)]  # [0, 1, 4, 9, 16]

# With condition
evens = [x for x in range(10) if x % 2 == 0]  # [0, 2, 4, 6, 8]
```

### 4. Using `*` operator:
```python
zeros = [0] * 5  # [0, 0, 0, 0, 0]
```