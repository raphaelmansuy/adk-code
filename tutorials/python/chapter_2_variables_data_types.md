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
*   **complex (Complex Numbers):** Numbers with a real and imaginary part (e.g., `3 + 4j`). These are primarily used in advanced mathematical and scientific computing.

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
    # message[0] = 'h' # This would cause a TypeError because strings are immutable
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
    # coordinates[0] = 15.0 # This would cause a TypeError because tuples are immutable
    ```

### 4. Mapping Type (Key-Value Pairs)
*   **dict (Dictionary):** An unordered collection of key-value pairs. Keys must be unique and immutable (like strings or numbers), while values can be of any data type. Dictionaries are defined using curly braces `{}` with `key: value` pairs.
    ```python
    person = {"name": "Bob", "age": 25, "city": "London"}
    scores = {"math": 90, "science": 85}
    ```

### 5. Set Types (Unordered Collections of Unique Items)
*   **set:** An unordered collection of unique items. Sets are mutable, meaning you can add or remove elements. They do not allow duplicate values and are useful for operations like checking membership, removing duplicates, and performing mathematical set operations (union, intersection). Sets are defined using curly braces `{}` or the `set()` constructor.
    ```python
    unique_numbers = {1, 2, 3, 3, 4} # will store {1, 2, 3, 4}
    print(unique_numbers) # Output: {1, 2, 3, 4} (order may vary)
    ```
*   **frozenset:** An **immutable** version of a set. Once created, you cannot add or remove elements. Frozensets can be used as keys in dictionaries or as elements in other sets, which regular (mutable) sets cannot.
    ```python
    immutable_set = frozenset([1, 2, 3])
    print(immutable_set) # Output: frozenset({1, 2, 3})
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