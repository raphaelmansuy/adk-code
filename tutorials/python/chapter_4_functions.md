# Chapter 4: Functions

## What are Functions?
A function is a block of organized, reusable code that is used to perform a single, related action. Functions provide better modularity for your application and a high degree of code reusing. You've already used built-in functions like `print()` and `type()`.

Beyond simply reusing code, functions are crucial for **breaking down complex problems into smaller, manageable pieces** (decomposition) and **combining these smaller functions to build larger, more complex functionalities** (composition). This approach significantly improves code readability, maintainability, and makes debugging easier.

## Defining a Function
In Python, you define a function using the `def` keyword, followed by the function name, parentheses `()`, and a colon `:`. The function body is then indented.

```python
def greet(name="Guest"):
    """This function prints a personalized greeting message.
    
    Args:
        name (str): The name of the person to greet. Defaults to "Guest".
    """
    print(f"Hello, {name}! Welcome to the function tutorial!")

# Calling the function
greet()          # Output: Hello, Guest! Welcome to the function tutorial!
greet("Alice")   # Output: Hello, Alice! Welcome to the function tutorial!

# Real-world example: A simple calculation function
def calculate_area_rectangle(length, width):
    """Calculates the area of a rectangle.
    
    Args:
        length (float): The length of the rectangle.
        width (float): The width of the rectangle.
        
    Returns:
        float: The area of the rectangle.
    """
    area = length * width
    return area

# Using the calculation function
room_area = calculate_area_rectangle(5.0, 3.5)
print(f"The area of the room is: {room_area} square units.") # Output: The area of the room is: 17.5 square units.
```

Functions are fundamental for:
*   **Modularity:** Breaking down complex problems into smaller, manageable pieces.
*   **Reusability:** Writing code once and using it multiple times.
*   **Readability:** Making your code easier to understand and maintain.
*   **Abstraction:** Hiding complex implementation details behind a simple function call.

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
*   **Docstring (optional but recommended):** A string literal used to document the purpose of the function. It's enclosed in triple quotes (`"""Docstring goes here"""`) and should immediately follow the function header. Docstrings are crucial for code documentation and can be accessed using `help(function_name)` or `function_name.__doc__`.

    ```python
    def example_function():
        """This is an example docstring."""
        pass

    print(example_function.__doc__) # Output: This is an example docstring.
    # help(example_function) # Uncomment to see full help output in a live environment
    ```
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

# Real-world example: Calculating the area of a triangle
def calculate_triangle_area(base, height):
    """Calculates the area of a triangle given its base and height."""
    area = 0.5 * base * height
    return area

# Using positional arguments
triangle_area = calculate_triangle_area(10, 4) # base=10, height=4
print(f"Area of triangle: {triangle_area} square units.") # Output: Area of triangle: 20.0 square units.
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
Sometimes you don't know how many arguments will be passed to your function, or you want to create highly flexible functions. Python provides `*args` (for non-keyword/positional arguments) and `**kwargs` (for keyword arguments).

*   `*args` collects all extra positional arguments into a tuple. Use it when your function needs to accept an arbitrary number of positional arguments (e.g., a `sum` function that can sum any number of items).
*   `**kwargs` collects all extra keyword arguments into a dictionary. Use it when your function needs to accept an arbitrary number of keyword arguments (e.g., passing configuration options).

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

*   **Local Scope:** Variables defined inside a function (including its parameters) have local scope. They are only accessible from within that function.
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

Here's a simple illustration of variable scope:

```mermaid
graph LR
    A[Global Scope] --> B(global_var = 20)
    A --> C{another_function()}
    C --> D[Access global_var]

    E[my_function()] --> F(local_var = 10)
    F --> G[End my_function]
    E -- calls --> G

    style A fill:#f9f,stroke:#333,stroke-width:2px
    style E fill:#ccf,stroke:#333,stroke-width:2px
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

    **What happens without `global`?**
    If you omit the `global` keyword, Python assumes you are creating a *new local variable* with the same name, rather than modifying the global one. This can lead to unexpected behavior.
    ```python
    another_global_counter = 0

    def try_to_increment_without_global():
        another_global_counter = 1 # This creates a NEW local variable, doesn't modify the global one
        print(f"Inside function (local): {another_global_counter}")

    print(f"Before call: {another_global_counter}") # Output: Before call: 0
    try_to_increment_without_global()                  # Output: Inside function (local): 1
    print(f"After call: {another_global_counter}")  # Output: After call: 0 (Global variable remains unchanged)
    ```
*   **`nonlocal` Keyword (for Nested Functions):** While less common for beginners, `nonlocal` is used in nested functions to modify variables in the nearest enclosing (but non-global) scope. This allows a function to modify a variable in its outer function's scope.
    ```python
    def outer_function():
        x = 10
        def inner_function():
            nonlocal x
            x = 20
        inner_function()
        print(f"Outer function: {x}") # Output: Outer function: 20

    outer_function()
    ```

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

### Higher-Order Functions (Advanced Concept)

Higher-order functions are functions that can take one or more functions as arguments, or return a function as their result. This is a powerful concept in functional programming paradigms and allows for more flexible and reusable code. You've already seen an example of this with `sorted()` and `lambda` functions, where `sorted()` takes a `key` argument that can be a function.

```python
def apply_operation(func, x, y):
    """Applies a given function to two values."""
    return func(x, y)

# Define simple functions
def add(a, b):
    return a + b

def subtract(a, b):
    return a - b

# Use apply_operation with different functions
result_add = apply_operation(add, 10, 5)
print(f"Applying add: {result_add}") # Output: 15

result_subtract = apply_operation(subtract, 10, 5)
print(f"Applying subtract (lambda): {result_subtract}") # Output: 5

# Using a lambda function with a higher-order function
result_multiply = apply_operation(lambda a, b: a * b, 10, 5)
print(f"Applying multiply (lambda): {result_multiply}") # Output: 50
```

In the final chapter, we will cover file input/output and error handling, essential skills for building robust applications.

## Best Practices and Common Pitfalls with Functions

To write clean, maintainable, and robust functions, consider these best practices and common pitfalls:

### Best Practices

1.  **Meaningful Names:** Always choose descriptive names for your functions and parameters that clearly indicate their purpose. (e.g., `calculate_area` instead of `ca`).
2.  **Single Responsibility Principle (SRP):** Each function should do one thing and do it well. Avoid functions that try to accomplish too many unrelated tasks.
3.  **Docstrings:** Write clear and concise docstrings for every function. Docstrings explain what the function does, its arguments, and what it returns. Use triple quotes (`"""Docstring"""`).
4.  **Type Hinting:** Use type hints for function parameters and return types (`def greet(name: str) -> None:`). This improves readability, helps IDEs provide better suggestions, and allows static analysis tools to catch errors early.
5.  **Keep Functions Small:** Aim for functions that are relatively short. If a function becomes too long or complex, it's often a sign that it can be broken down into smaller, more focused functions.
6.  **Avoid Global Variables (where possible):** Relying heavily on global variables can make code harder to understand and debug. Pass necessary data as arguments instead. If modification of a global variable is truly needed, use the `global` keyword.

### Common Pitfalls

1.  **Incorrect Argument Passing:**
    *   **Missing arguments:** Calling a function without providing all required positional arguments will result in a `TypeError`.
    *   **Extra arguments:** Providing too many arguments will also result in a `TypeError`.
    *   **Incorrect types:** While Python doesn't enforce types at runtime (without explicit checks), passing arguments of unexpected types can lead to runtime errors or incorrect behavior. Type hints help mitigate this.
2.  **Mutable Default Arguments:** Be very cautious when using mutable objects (like lists or dictionaries) as default parameter values. They are created only once when the function is defined, not on each call, which can lead to unexpected shared state across function calls.
    ```python
def add_to_list(item, my_list=[]): # DANGER! my_list is created once
        my_list.append(item)
        return my_list

    print(add_to_list(1))    # Output: [1]
    print(add_to_list(2))    # Output: [1, 2] - Unexpected!
    print(add_to_list(3, [])) # Output: [3] - This works as expected
    ```
    **Solution:** Use `None` as a default and create the mutable object inside the function if `None` is passed.
    ```python
def add_to_list_safe(item, my_list=None):
        if my_list is None:
            my_list = []
        my_list.append(item)
        return my_list

    print(add_to_list_safe(1)) # Output: [1]
    print(add_to_list_safe(2)) # Output: [2] - Correct behavior
    ```
3.  **Forgetting `return`:** If a function performs a calculation but doesn't `return` a value, it will implicitly return `None`. Always explicitly `return` the value you intend to send back.
4.  **Scope Misunderstanding:** Confusing local and global variables, especially when trying to modify a global variable without using the `global` keyword.

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

## Exercise 1: Simple Math Operations

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