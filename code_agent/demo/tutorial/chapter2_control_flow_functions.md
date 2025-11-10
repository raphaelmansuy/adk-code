# Chapter 2: Control Flow and Functions

## Learning Objectives
In this chapter, you will:
*   Master conditional statements (`if`, `elif`, `else`) to make decisions in your programs.
*   Learn to use `for` and `while` loops for repetitive tasks.
*   Understand `break` and `continue` to control loop execution.
*   Define and call functions with arguments and return values.
*   Discover how to document your functions using docstrings.

This chapter will introduce you to controlling the flow of your Python programs using conditional statements and loops, and how to organize your code into reusable blocks with functions.

## Conditional Statements: `if`, `elif`, `else`

Conditional statements allow your program to make decisions based on certain conditions. Python uses `if`, `elif` (else if), and `else`. The following flowchart illustrates a typical `if-elif-else` structure:

```text
        +-------+
        | Start |
        +---v---+
            |
            v
      +-------------+
      | Condition 1?|
      +------+------+
             |   True
        False|-------> +----------+
             |          | Action 1 |
             v          +----v-----+
      +-------------+        |
      | Condition 2?|        v
      +------+------+
             |   True
        False|-------> +----------+
             |          | Action 2 |
             v          +----v-----+
      +-------------+        |
      | Action 3    |<-------+
      |   (Else)    |
      +-------^-----+
              |
              v
            +---+
            | End |
            +-----+
```

```python
# Example 1: Basic if statement
x = 10
if x > 5:
    print("x is greater than 5")

# Example 2: if-else statement
y = 3
if y % 2 == 0:
    print("y is an even number")
else:
    print("y is an odd number")

# Example 3: if-elif-else chain
score = 85
if score >= 90:
    print("Grade A")
elif score >= 80:
    print("Grade B")
elif score >= 70:
    print("Grade C")
else:
    print("Grade F")
```

**Important:** Python uses indentation (usually 4 spaces) to define code blocks. This is crucial for readability and correctness.

### The `pass` Statement

The `pass` statement is a null operation; nothing happens when it executes. It's useful as a placeholder when a statement is syntactically required but you don't want any code to execute. This can be helpful when you're planning out your code structure.

```python
def future_function():
    pass # TODO: Implement this later

if True:
    pass # Do nothing for now
```

## Loops: `for` and `while`

Loops allow you to execute a block of code multiple times.

### `for` Loop

The `for` loop is used for iterating over a sequence (like a list, tuple, string, or range).

```python
# Iterating over a list
fruits = ["apple", "banana", "cherry"]
for fruit in fruits:
    print(fruit)

# Iterating using range()
# range(5) generates numbers from 0 to 4
for i in range(5):
    print(i)

# range(start, stop)
for i in range(2, 6):
    print(i) # Output: 2, 3, 4, 5

# range(start, stop, step)
for i in range(1, 10, 2):
    print(i) # Output: 1, 3, 5, 7, 9
```

### `while` Loop

The `while` loop repeatedly executes a block of code as long as a given condition is true.

```python
count = 0
while count < 5:
    print(f"Count is {count}") # f-string for formatted output
    count += 1 # Increment count
```

**Caution:** Be careful with `while` loops to avoid infinite loops. Ensure the condition eventually becomes false.

## `break` and `continue` Statements

*   **`break`:** Terminates the current loop entirely and execution resumes at the statement immediately following the loop.
*   **`continue`:** Skips the rest of the current iteration and moves to the next iteration of the loop.

```python
# break example
for i in range(10):
    if i == 5:
        break
    print(i) # Output: 0, 1, 2, 3, 4

# continue example
for i in range(10):
    if i % 2 == 0: # Skip even numbers
        continue
    print(i) # Output: 1, 3, 5, 7, 9
```

## Defining Functions

Functions are blocks of organized, reusable code that perform a single, related action. They help break down your program into smaller, manageable, and modular chunks.

To define a function, use the `def` keyword:

```python
def greet():
    print("Hello there!")

# Calling the function
greet()
```

## Docstrings: Documenting Your Functions

Docstrings (documentation strings) are multiline strings used to document modules, classes, functions, and methods. They are enclosed in triple quotes (`"""Docstring goes here"""`) and should explain what the code does, its arguments, and what it returns. Good docstrings are essential for writing readable and maintainable code.

```python
def add_numbers(a, b):
    """
    This function takes two numbers as input and returns their sum.

    Args:
        a (int or float): The first number.
        b (int or float): The second number.

    Returns:
        (int or float): The sum of the two numbers.
    """
    return a + b

# You can access a function's docstring using help() or the __doc__ attribute
print(add_numbers.__doc__)
help(add_numbers)
```

## Function Arguments and Return Values

### Arguments

Functions can accept arguments (inputs) to work with.

```python
def greet_name(name):
    print(f"Hello, {name}!")

greet_name("Alice") # Output: Hello, Alice!
greet_name("Bob")   # Output: Hello, Bob!
```

**Default Arguments:** You can provide default values for arguments.

```python
def greet_with_default(name="Guest"):
    print(f"Hello, {name}!")

greet_with_default()      # Output: Hello, Guest!
greet_with_default("Charlie") # Output: Hello, Charlie!
```

**Keyword Arguments:** You can pass arguments using their names, which improves readability and allows you to pass them in any order.

```python
def describe_pet(animal_type, pet_name):
    print(f"I have a {animal_type} named {pet_name}.")

describe_pet(animal_type="dog", pet_name="Buddy")
describe_pet(pet_name="Whiskers", animal_type="cat")
```

### Return Values

Functions can return values using the `return` statement. If no `return` statement is present, the function implicitly returns `None`.

```python
def add(a, b):
    return a + b

result = add(5, 3)
print(result) # Output: 8

def multiply(x, y):
    product = x * y
    return product

result2 = multiply(4, 6)
print(result2) # Output: 24
```

## Exercises

1.  **Even or Odd:** Write a program that takes an integer as input and prints whether it's even or odd.
2.  **Sum of List Elements:** Create a function `calculate_sum(numbers)` that takes a list of numbers and returns their sum. Use a `for` loop.
3.  **Factorial Function:** Write a function `factorial(n)` that calculates the factorial of a given non-negative integer `n`. (Hint: `n! = n * (n-1) * ... * 1`)
4.  **Countdown:** Use a `while` loop to print a countdown from 5 to 1, then print "Go!".
5.  **Docstring Practice:** Add a comprehensive docstring to your `factorial` function from Exercise 3.

This concludes Chapter 2. You now have a solid understanding of how to control the flow of your programs and write reusable functions.