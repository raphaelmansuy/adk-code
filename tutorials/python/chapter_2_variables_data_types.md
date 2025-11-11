# Chapter 2: Variables and Data Types

## What are Variables?
In programming, a variable is a named storage location that holds a value. Think of it as a container where you can store different types of information. In Python, you don't need to explicitly declare the type of a variable; Python infers it based on the value you assign.

## Declaring and Assigning Variables
You assign a value to a variable using the `=` operator.

### Variable Naming Conventions (PEP 8)
Python has guidelines for writing readable code, known as PEP 8. For variable names, the convention is:
*   Use lowercase letters.
*   Separate words with underscores (`_`) for readability (e.g., `user_name`, `total_price`).
*   Avoid using Python's reserved keywords (like `if`, `for`, `class`).
*   Be descriptive! Choose names that clearly indicate the variable's purpose.

```python
# Assigning an integer to a variable
age = 30

# Assigning a string to a variable
user_name = "Alice" # Using snake_case for readability

# Assigning a floating-point number
total_price = 19.99

# Assigning a boolean value
is_student = True

print(age)
print(user_name)
print(total_price)
print(is_student)
```

Here's a conceptual diagram of how variables store values:

```
+------------+       +-------------------+
|  Variable  |       |   Memory Values   |
+------------+       +-------------------+
|    age     |-----> | 30 (integer)      |
+------------+       +-------------------+
| user_name  |-----> | "Alice" (string)  |
+------------+       +-------------------+
| total_price|-----> | 19.99 (float)     |
+------------+       +-------------------+
| is_student |-----> | True (boolean)    |
+------------+       +-------------------+
```

## Python's Built-in Data Types
Python has several built-in data types to handle various kinds of data.

### 1. Numeric Types
*   **int (Integers):** Whole numbers (positive, negative, or zero) without a fractional part.
    ```python
    x = 10
    y = -5
    ```
*   **float (Floating-Point Numbers):** Numbers with a decimal point.
    ```python
    pi = 3.14
    temperature = 98.6
    ```
### 2. Boolean Type
*   **bool:** Represents truth values. It can only be `True` or `False`.
    ```python
    is_active = True
    has_permission = False
    ```

### 3. Sequence Types (Ordered Collections)
*   **str (Strings):** A sequence of characters. Strings are **immutable** (cannot be changed after creation). This means once a string is created, you cannot modify individual characters within it. Any operation that seems to modify a string actually creates a new string.
    ```python
    message = "Hello, Python!"
    city = 'New York'
    print(message[0]) # Accessing individual characters
    try:
        message[0] = 'h' # This would cause a TypeError because strings are immutable
    except TypeError as e:
        print(f"Error: {e}") # Output: Error: 'str' object does not support item assignment
    ```
    You can use single quotes (`'`) or double quotes (`"`) for strings.

    **F-Strings (Formatted String Literals):** A powerful and readable way to embed expressions inside string literals.
    ```python
    name = "Alice"
    age = 30
    greeting = f"Hello, {name}! You are {age} years old."
    print(greeting) # Output: Hello, Alice! You are 30 years old.
    ```

*   **list:** An ordered, **mutable** sequence of items. Lists can contain items of different data types and their contents can be changed after creation (add, remove, modify elements). They are defined using square brackets `[]`.
    ```python
    fruits = ["apple", "banana", "cherry"]
    numbers = [1, 2, 3, 4, 5]
    mixed_list = ["text", 10, True, 3.14]

    # Example of mutability with lists
    fruits[0] = "grape"
    fruits.append("mango")
    print(fruits) # Output: ['grape', 'banana', 'cherry', 'mango']
    ```

*   **tuple:** An ordered, **immutable** sequence of items. Tuples are similar to lists but, like strings, their contents cannot be changed after creation. This makes them useful for data that should not be altered, such as coordinates or fixed configurations. They are defined using parentheses `()`.
    ```python
    coordinates = (10.0, 20.0)
    colors = ("red", "green", "blue")
    try:
        coordinates[0] = 15.0 # This would cause a TypeError because tuples are immutable
    except TypeError as e:
        print(f"Error: {e}") # Output: Error: 'tuple' object does not support item assignment
    ```

### 4. Mapping Type (Key-Value Pairs)
*   **dict (Dictionary):** A mutable collection of key-value pairs. While historically considered unordered, since Python 3.7, dictionaries are **ordered**, meaning they retain the insertion order of items. Keys must be unique and immutable (like strings or numbers), while values can be of any data type. Dictionaries are defined using curly braces `{}` with `key: value` pairs.
    ```python
    person = {"name": "Bob", "age": 25, "city": "London"}
    scores = {"math": 90, "science": 85}
    # Accessing values
    print(person["name"]) # Output: Bob
    # Modifying values
    person["age"] = 26
    print(person) # Output: {'name': 'Bob', 'age': 26, 'city': 'London'}
    ```

### 5. Set Types (Unordered Collections of Unique Items)
*   **set:** An unordered collection of unique items. Sets are mutable, meaning you can add or remove elements. They do not allow duplicate values and are useful for operations like checking membership, removing duplicates, and performing mathematical set operations (union, intersection). Sets are defined using curly braces `{}` or the `set()` constructor.
    ```python
    unique_numbers = {1, 2, 3, 3, 4} # duplicates are automatically removed
    print(unique_numbers) # Output: {1, 2, 3, 4} (the order of elements is not guaranteed)
    print(3 in unique_numbers) # Output: True (checking for membership is very efficient with sets)
    
    # Demonstrating common set operations
    set_a = {1, 2, 3, 4}
    set_b = {3, 4, 5, 6}
    print(f"Set A: {set_a}")
    print(f"Set B: {set_b}")
    print(f"Union (A | B): {set_a.union(set_b)}")          # Elements in A or B or both
    print(f"Intersection (A & B): {set_a.intersection(set_b)}") # Elements common to A and B
    print(f"Difference (A - B): {set_a.difference(set_b)}")   # Elements in A but not in B
    ```
*   **frozenset:** An **immutable** version of a set. Once created, you cannot add or remove elements. Frozensets can be used as keys in dictionaries or as elements in other sets, which regular (mutable) sets cannot.not.
    ```python
    immutable_set = frozenset([1, 2, 3])
    print(immutable_set) # Output: frozenset({1, 2, 3})
    ```

### 6. None Type
*   **NoneType (None):** Represents the absence of a value or a null value. It's often used to indicate that a variable has not been assigned anything yet, or as the default return value for functions that don't explicitly return anything.
    ```python
    result = None
    print(result) # Output: None
    print(type(result)) # Output: <class 'NoneType'>

    def do_nothing():
        pass # This function implicitly returns None
    
    nothing_value = do_nothing()
    print(nothing_value) # Output: None
    ```

## Checking Data Types
You can use the `type()` function to check the data type of a variable.

```python
my_variable = 100
print(type(my_variable)) # Output: <class 'int'>

my_string = "Python"
print(type(my_string)) # Output: <class 'str'>

my_list = [1, 2, 3]
print(type(my_list)) # Output: <class 'list'>
```

## Type Conversion (Type Casting)
You can convert values from one data type to another using built-in functions like `int()`, `float()`, `str()`, `list()`, etc.

```python
# Convert int to float
num_int = 10
num_float = float(num_int)
print(num_float) # Output: 10.0

# Convert float to int (truncates decimal part)
num_float_2 = 12.7
num_int_2 = int(num_float_2)
print(num_int_2) # Output: 12

# Convert int to string
num_str = str(num_int)
print(type(num_str)) # Output: <class 'str'>

# Convert string to int
str_num = "45"
int_from_str = int(str_num)
print(int_from_str) # Output: 45

# Convert string to float
str_float = "3.14"
float_from_str = float(str_float)
print(float_from_str) # Output: 3.14
```

**Caution:** Be careful when converting strings to numbers. If the string doesn't represent a valid number, it will raise a `ValueError`.

```python
# Example of ValueError during type conversion
# invalid_num_str = "hello"
# int_from_invalid_str = int(invalid_num_str) # This would raise a ValueError
try:
    invalid_num_str = "hello"
    int_from_invalid_str = int(invalid_num_str)
except ValueError as e:
    print(f"Error: {e}") # Output: Error: invalid literal for int() with base 10: 'hello'
```

In the next chapter, we will explore how to control the flow of your program using conditional statements and loops.

## Key Takeaways
*   Variables are named storage locations for values, and Python infers their type.
*   Python has several built-in data types, including numeric (int, float, complex), boolean (bool), sequence (str, list, tuple), mapping (dict), and set (set, frozenset) types.
*   Strings and tuples are **immutable**, meaning their content cannot be changed after creation.
*   Lists, dictionaries, and sets are **mutable**, meaning their content can be modified.
*   The `type()` function checks a variable's data type, and functions like `int()`, `float()`, `str()` can convert between types.
*   F-strings provide an efficient and readable way to format strings.

## Exercise 2: Student Profile

Create variables to store the following information for a student:
*   Student's name (string)
*   Student's age (integer)
*   Student's GPA (float)
*   Whether the student is enrolled (boolean)
*   A list of courses they are taking (list of strings)
*   A dictionary with their grades for two courses (e.g., {"Math": "A", "Science": "B"})

Then, print all this information using f-strings for clear output.

**Example Output:**
```
Student Name: John Doe
Age: 20
GPA: 3.85
Enrolled: True
Courses: ['Calculus', 'Physics', 'Literature']
Grades: {'Math': 'A', 'Science': 'B'}
```

**Hint:** Remember to use appropriate data types for each piece of information and f-strings for printing.