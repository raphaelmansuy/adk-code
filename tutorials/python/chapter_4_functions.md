# Chapter 4: Functions

## What are Functions?
A function is a block of organized, reusable code that is used to perform a single, related action. Functions provide better modularity for your application and a high degree of code reusing. You've already used built-in functions like `print()` and `type()`.

## Defining a Function
In Python, you define a function using the `def` keyword, followed by the function name, parentheses `()`, and a colon `:`. The function body is then indented.

```python
def greet():
    """This function prints a greeting message."""
    print("Hello, welcome to the function tutorial!")

# Calling the function
greet()
```

Here's a conceptual diagram of a function:

```
        +-------------------+
        | Function Call     |
        | (e.g., greet())   |
        +--------+----------+
                 |
                 V
+-----------------------------------+
| def function_name(parameter1, ...): |
|    """Docstring"""                |
|    # Function Body                |
|    result = operation()           |
|    return result                  |
+-----------------------------------+
                 |
                 V
        +-------------------+
        | Return Value      |
        +-------------------+
```

*   **`def` keyword:** Marks the start of the function header.
*   **Function Name:** Must follow Python's naming conventions (lowercase, words separated by underscores).
*   **Parentheses `()`:** Can contain parameters (inputs) to the function.
*   **Colon `:`:** Marks the end of the function header.
*   **Docstring (optional but recommended):** A string literal used to document the purpose of the function. It's enclosed in triple quotes (`"""Docstring goes here"""`). Docstrings are crucial for code documentation and can be accessed using `help(function_name)` or `function_name.__doc__`.
*   **Function Body:** The block of code that the function executes, always indented.

## Function Parameters and Arguments
Functions can accept input values called parameters. When you call the function, you pass actual values, which are called arguments.

### 1. Positional Arguments
Arguments are passed to parameters in the order they are defined.

```python
def add_numbers(a, b):
    """This function takes two numbers and returns their sum."""
    sum_result = a + b
    print(f"The sum of {a} and {b} is {sum_result}")

add_numbers(10, 5) # a=10, b=5 (positional arguments)
add_numbers(3, 7)  # a=3, b=7
```

### 2. Keyword Arguments
You can pass arguments by specifying the parameter name, which allows you to pass them in any order. This improves readability, especially for functions with many parameters.

```python
def introduce(name, age):
    print(f"My name is {name} and I am {age} years old.")

introduce(name="Alice", age=30) # Using keyword arguments
introduce(age=25, name="Bob")   # Order doesn't matter with keyword arguments
```

### 3. Default Parameter Values
You can provide default values for parameters. If an argument is not provided for such a parameter during the function call, its default value is used. Parameters with default values must come after any non-default parameters.

```python
def say_hello(name="Guest", language="English"): # "Guest" and "English" are default values
    print(f"Hello, {name}! ({language})")

say_hello()                     # Output: Hello, Guest! (English)
say_hello("Charlie")            # Output: Hello, Charlie! (English)
say_hello("David", "Spanish")   # Output: Hello, David! (Spanish)
say_hello(language="French", name="Eve") # Can mix with keyword arguments
```

### 4. Arbitrary Arguments (`*args` and `**kwargs`)
Sometimes you don't know how many arguments will be passed to your function. Python provides `*args` (for non-keyword/positional arguments) and `**kwargs` (for keyword arguments).

*   `*args` collects all extra positional arguments into a tuple.
*   `**kwargs` collects all extra keyword arguments into a dictionary.

```python
def sum_all_numbers(*args):
    """Sums all positional arguments."""
    total = 0
    for num in args:
        total += num
    return total

print(sum_all_numbers(1, 2, 3))        # Output: 6
print(sum_all_numbers(10, 20, 30, 40)) # Output: 100

def print_info(**kwargs):
    """Prints key-value pairs from keyword arguments."""
    for key, value in kwargs.items():
        print(f"{key}: {value}")

print_info(name="Alice", age=30, city="New York")
# Output:
# name: Alice
# age: 30
# city: New York
```

## Type Hinting (Optional but Recommended)

Python is dynamically typed, meaning you don't have to declare variable types. However, for better readability, maintainability, and to help tools like IDEs catch potential errors, Python introduced type hints (PEP 484). These don't affect how the code runs but provide useful metadata.

```python
def add(a: int, b: int) -> int:
    """Adds two integers and returns their sum."""
    return a + b

def greet_user(name: str) -> None:
    """Greets a user by name."""
    print(f"Hello, {name}!")

result = add(5, 3)
print(result) # Output: 8

greet_user("Bob") # Output: Hello, Bob!
```

## The `return` Statement
The `return` statement is used to send a value back from the function to the caller. If a function doesn't have a `return` statement, it implicitly returns `None`.

```python
def multiply(x, y):
    """This function multiplies two numbers and returns the result."""
    result = x * y
    return result

product = multiply(4, 6)
print(f"The product is: {product}") # Output: The product is: 24

def get_full_name(first, last):
    return f"{first} {last}"

full_name = get_full_name("John", "Doe")
print(full_name)
```

## Scope of Variables

*   **Local Scope:** Variables defined inside a function are local to that function and cannot be accessed from outside.
    ```python
    def my_function():
        local_var = 10
        print(local_var)

    my_function()
    # print(local_var) # This would cause an error (NameError)
    ```
*   **Global Scope:** Variables defined outside any function are global and can be accessed from anywhere in the program.
    ```python
global_var = 20

def another_function():
    print(global_var) # Can access global_var

another_function()
print(global_var)
    ```
    If you need to *modify* a global variable from within a function, you must use the `global` keyword:
    ```python
global_counter = 0

def increment_counter():
    global global_counter # Declare intent to modify the global variable
    global_counter += 1
    print(f"Inside function: {global_counter}")

print(f"Before call: {global_counter}") # Output: Before call: 0
increment_counter()                     # Output: Inside function: 1
print(f"After call: {global_counter}")  # Output: After call: 1
    ```
    Without `global`, `global_counter += 1` inside `increment_counter()` would create a new local variable named `global_counter` instead of modifying the global one.

## Lambda Functions (Anonymous Functions)

Lambda functions are small, anonymous functions defined with the `lambda` keyword. They can take any number of arguments but can only have one expression. They are often used for short, throw-away functions, especially as arguments to higher-order functions (functions that take other functions as arguments).

```python
# A regular function
def add_five(x):
    return x + 5

# Equivalent lambda function
add_five_lambda = lambda x: x + 5

print(add_five(10))        # Output: 15
print(add_five_lambda(10)) # Output: 15

# Using lambda with built-in functions like sorted()
pairs = [(1, 'one'), (2, 'two'), (3, 'three'), (4, 'four')]
pairs.sort(key=lambda pair: pair[1]) # Sort by the second element of each tuple
print(pairs)
# Output: [(4, 'four'), (1, 'one'), (3, 'three'), (2, 'two')]
```

In the final chapter, we will cover file input/output and error handling, essential skills for building robust applications.

## Key Takeaways
*   Functions are reusable blocks of code that perform specific tasks.
*   Define functions using `def`, followed by a name, parameters, and a colon. The function body is indented.
*   Docstrings (triple-quoted strings) explain a function's purpose.
*   Function arguments can be positional, keyword, have default values, or be arbitrary (`*args`, `**kwargs`).
*   The `return` statement sends a value back to the caller; functions without `return` implicitly return `None`.
*   Variables have either local (within function) or global (accessible everywhere) scope.
*   Use `global` keyword to modify global variables inside a function.
*   Type hints improve code readability and help with static analysis.
*   Lambda functions are small, anonymous functions suitable for simple, single expressions.

## Exercise 4: Simple Math Operations

Create a Python script that defines the following functions:

1.  `add(a, b)`: Takes two numbers and returns their sum.
2.  `subtract(a, b)`: Takes two numbers and returns their difference.
3.  `multiply(a, b)`: Takes two numbers and returns their product.
4.  `divide(a, b)`: Takes two numbers and returns their quotient. Handle division by zero by returning an error message (e.g., a string like "Error: Cannot divide by zero.")

Then, use these functions to perform a few operations and print their results. Include type hints for all functions.

**Example Usage:**
```python
print(f"Addition: {add(10, 5)}")
print(f"Subtraction: {subtract(10, 5)}")
print(f"Multiplication: {multiply(10, 5)}")
print(f"Division: {divide(10, 5)}")
print(f"Division by zero: {divide(10, 0)}")
```

**Expected Output (approximately):**
```
Addition: 15
Subtraction: 5
Multiplication: 50
Division: 2.0
Division by zero: Error: Cannot divide by zero.
```

**Hint:** Remember to add type hints like `a: int, b: int) -> int` to your function definitions. For division by zero, use an `if` statement.