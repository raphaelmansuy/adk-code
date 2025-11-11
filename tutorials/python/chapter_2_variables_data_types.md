# Chapter 2: Variables and Data Types

## What are Variables?
In programming, a variable is a named storage location that holds a value. Think of it as a container where you can store different types of information. In Python, you don't need to explicitly declare the type of a variable; Python infers it based on the value you assign.

## Declaring and Assigning Variables
You assign a value to a variable using the `=` operator. Python is dynamically typed, meaning you can reassign a variable to a new value of a different type.

### Variable Naming Conventions (PEP 8)
Python has guidelines for writing readable code, known as PEP 8. For variable names, the convention is:
*   Use lowercase letters.
*   Separate words with underscores (`_`) for readability (e.g., `user_name`, `total_price`).
*   Avoid using Python's reserved keywords (like `if`, `for`, `class`).
*   Be descriptive! Choose names that clearly indicate the variable's purpose. **Good naming improves code readability and maintainability, especially in larger projects or when collaborating with others.**

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
    # Example: Number of students in a class
    num_students = 30
    # Example: Year of birth
    birth_year = 1990
    print(f"Number of students: {num_students}") # Output: Number of students: 30
    print(f"Birth year: {birth_year}")         # Output: Birth year: 1990
    ```
*   **float (Floating-Point Numbers):** Numbers with a decimal point.
    ```python
    # Example: Price of a product
    product_price = 29.99
    # Example: Average temperature
    avg_temperature = 23.5
    print(f"Product price: ${product_price}") # Output: Product price: $29.99
    print(f"Average temperature: {avg_temperature}°C") # Output: Average temperature: 23.5°C

    # Real-world example: Calculating total price with sales tax
    item_price = 50.00
    sales_tax_rate = 0.075 # 7.5%
    total_cost = item_price + (item_price * sales_tax_rate)
    print(f"Item price: ${item_price:.2f}")
    print(f"Sales tax rate: {sales_tax_rate:.1%}") # Formatted as percentage
    print(f"Total cost with tax: ${total_cost:.2f}") # Output: Total cost with tax: $53.75
    ```
*   **complex (Complex Numbers):** Represents numbers with a real and an imaginary part (e.g., `1 + 2j`). While less common in introductory programming, they are used in specialized mathematical and engineering applications and will not be covered in detail here.
### 2. Boolean Type
*   **bool (Boolean):** Represents truth values. It can only be `True` or `False`. Booleans are fundamental for conditional logic.
    ```python
    # Example: Is a user logged in?
    is_logged_in = True
    # Example: Is a file saved?
    file_saved = False
    print("User logged in:", is_logged_in)   # Output: User logged in: True
    print("File saved:", file_saved)       # Output: File saved: False

    # Booleans are often the result of comparison operations
    age = 20 # Define 'age' for this example
    is_adult = (age >= 18)
    print("Is adult (age >= 18):", is_adult)
    ```

### 3. Sequence Types (Ordered Collections)
*   **str (Strings):** A sequence of characters. Strings are **immutable** (cannot be changed after creation). This means once a string is created, you cannot modify individual characters within it. Any operation that seems to modify a string actually creates a new string.
    ```python
    # Example: Storing a user's name and a greeting
    first_name = "Alice"
    last_name = "Smith"
    full_name = first_name + " " + last_name # String concatenation
    greeting = "Hello, " + full_name + "!"

    print(f"Full Name: {full_name}") # Output: Full Name: Alice Smith
    print(f"Greeting: {greeting}")   # Output: Hello, Alice Smith!
    print(f"Length of greeting: {len(greeting)}") # Output: Length of greeting: 24
    print(f"First character: {greeting[0]}") # Output: First character: H (indexing)
    print(f"Slice (0-5): {greeting[0:5]}") # Output: Slice (0-5): Hello (slicing)

    # Real-world example: Formatting an address
    street = "123 Main St"
    city = "Anytown"
    state = "CA"
    zip_code = "90210"
    formatted_address = f"{street}, {city}, {state} {zip_code}"
    print(f"Formatted Address: {formatted_address}") # Output: 123 Main St, Anytown, CA 90210

    # Attempting to modify a character (will raise an error)
    try:
        # Strings are immutable - this will fail
        greeting[0] = 'h'
    except TypeError as e:
        print(f"Error trying to modify string: {e}")
    ```
    You can use single quotes (`'`) or double quotes (`"`) for strings. Multi-line strings can be created using triple quotes (`'''` or `"""`).

    **F-Strings (Formatted String Literals):** A powerful and readable way to embed expressions inside string literals.
    ```python
    product = "Laptop"
    price = 1200.50
    description = f"The {product} costs ${price:.2f}."
    print(description) # Output: The Laptop costs $1200.50.
    ```

*   **list:** An ordered, **mutable** sequence of items. Lists can contain items of different data types and their contents can be changed after creation (add, remove, modify elements). They are defined using square brackets `[]`.
    ```python
    # Example: A list of tasks to complete
    tasks = ["Learn Python", "Practice coding", "Build a project"]
    print(f"Initial tasks: {tasks}")

    # Add a new task
    tasks.append("Review concepts")
    print(f"After append: {tasks}") # Output: ['Learn Python', 'Practice coding', 'Build a project', 'Review concepts']

    # Modify a task
    tasks[0] = "Master Python Basics"
    print(f"After modification: {tasks}") # Output: ['Master Python Basics', 'Practice coding', 'Build a project', 'Review concepts']

    # Remove a task by value
    tasks.remove("Practice coding")
    print(f"After remove: {tasks}") # Output: ['Master Python Basics', 'Build a project', 'Review concepts']

    # Remove a task by index (and get its value)
    completed_task = tasks.pop(1) # Removes 'Build a project'
    print(f"Completed task: {completed_task}, Remaining tasks: {tasks}") # Output: Completed task: Build a project, Remaining tasks: ['Master Python Basics', 'Review concepts']

    # A list can contain mixed data types
    mixed_data = ["apple", 1, True, 3.14]
    print(f"Mixed data list: {mixed_data}")

    # Real-world example: Managing product prices
    product_prices = [29.99, 10.50, 5.00, 10.50, 99.99]
    print(f"Original product prices: {product_prices}")

    product_prices.sort() # Sorts the list in ascending order
    print(f"Sorted prices: {product_prices}") # Output: [5.0, 10.5, 10.5, 29.99, 99.99]

    count_of_10_50 = product_prices.count(10.50)
    print(f"Count of $10.50 items: {count_of_10_50}") # Output: 2
    ```

*   **tuple:** An ordered, **immutable** sequence of items. Tuples are similar to lists but, like strings, their contents cannot be changed after creation. This makes them useful for data that should not be altered, such as coordinates, fixed configurations, or when returning multiple values from a function. They are defined using parentheses `()`.
    ```python
    # Example: RGB color values (should not change)
    rgb_color = (255, 0, 128)
    print(f"RGB Color: {rgb_color}")
    print(f"Green component: {rgb_color[1]}") # Accessing elements

    # Attempting to modify a tuple (will raise an error)
    try:
        # Tuples are immutable - this will fail
        rgb_color[0] = 200
    except TypeError as e:
        print(f"Error trying to modify tuple: {e}")

    # Real-world example: Storing geographical coordinates (fixed values)
    coordinates = (34.0522, -118.2437) # (latitude, longitude)
    print(f"City coordinates: Latitude={coordinates[0]}, Longitude={coordinates[1]}")

    # Example: Function returning multiple values as a tuple
    def get_user_info():
        return "Charlie", 35, "Engineer"

    name, age, occupation = get_user_info()
    print(f"User Info: Name={name}, Age={age}, Occupation={occupation}")
    ```

### 4. Mapping Type (Key-Value Pairs)
*   **dict (Dictionary):** A mutable collection of key-value pairs. Keys must be unique and immutable (like strings or numbers), while values can be of any data type. Dictionaries are defined using curly braces `{}` with `key: value` pairs.

    **Note on Order:** Since Python 3.7, dictionaries are **ordered**, meaning they retain the insertion order of items. You can rely on this behavior.
    ```python
    # Example: Storing user preferences
    user_profile = {"username": "johndoe", "email": "john@example.com", "is_premium": False}
    print(f"Initial profile: {user_profile}")

    # Accessing values
    print(f"Username: {user_profile["username"]}") # Output: Username: johndoe

    # Modifying values
    user_profile["is_premium"] = True
    print(f"After updating premium status: {user_profile}") # Output: ... 'is_premium': True}

    # Adding a new key-value pair
    user_profile["last_login"] = "2023-10-26"
    print(f"After adding last login: {user_profile}")

    # Removing a key-value pair
    del user_profile["email"]
    print(f"After removing email: {user_profile}")

    # Safely accessing a key using .get()
    country = user_profile.get("country", "Unknown") # Returns 'Unknown' if 'country' key doesn't exist
    print(f"User's country: {country}")

    # Iterating through a dictionary
    print("\nUser Profile Details:")
    for key, value in user_profile.items():
        print(f"  {key}: {value}")

    # Real-world example: Counting word frequencies
    sentence = "the quick brown fox jumps over the lazy dog the quick brown fox"
    words = sentence.split()
    word_counts = {}
    for word in words:
        word_counts[word] = word_counts.get(word, 0) + 1
    print(f"\nWord frequencies: {word_counts}")
    # Output: {'the': 4, 'quick': 2, 'brown': 2, 'fox': 2, 'jumps': 1, 'over': 1, 'lazy': 1, 'dog': 1}
    ```

### 5. Set Types (Unordered Collections of Unique Items)
*   **set:** An unordered, **mutable** collection of unique items. Sets do not allow duplicate values and are useful for operations like checking membership, removing duplicates from a list, and performing mathematical set operations (union, intersection, difference).
    ```python
    # Example: Tracking unique visitors to a website (duplicates are automatically handled)
    website_visitors = {"Alice", "Bob", "Charlie", "Alice"} # "Alice" is automatically deduplicated
    print(f"Unique website visitors: {website_visitors}") # Output: {'Bob', 'Alice', 'Charlie'} (order may vary)

    # Sets are great for quickly checking if an item is present
    print(f"Is 'Charlie' a visitor? {'Charlie' in website_visitors}") # Output: True
    print(f"Is 'Frank' a visitor? {'Frank' in website_visitors}")   # Output: False

    # Adding and removing elements
    website_visitors.add("David")
    website_visitors.remove("Bob")
    print(f"Visitors after add/remove: {website_visitors}")

    # Demonstrating common set operations (useful for data analysis)
    # Imagine two groups of students and their enrolled courses
    math_students = {"Alice", "Bob", "David"}
    science_students = {"Charlie", "David", "Eve"}

    print(f"\nMath students: {math_students}")
    print(f"Science students: {science_students}")
    print(f"All students (Union): {math_students.union(science_students)}") # Students in Math OR Science (or both)
    print(f"Students in both (Intersection): {math_students.intersection(science_students)}") # Students in Math AND Science
    print(f"Students only in Math (Difference): {math_students.difference(science_students)}") # Students in Math but NOT Science
    print(f"Students not in both (Symmetric Difference): {math_students.symmetric_difference(science_students)}") # Students in Math OR Science, but NOT both

    # Real-world example: Finding unique items in a list
    data_list = [1, 2, 2, 3, 4, 4, 5, 1]
    unique_items = set(data_list)
    print(f"\nOriginal list: {data_list}")
    print(f"Unique items from list: {unique_items}") # Output: {1, 2, 3, 4, 5}
    ```
*   **frozenset:** An **immutable** version of a set. Once created, you cannot add or remove elements. Frozensets can be used as keys in dictionaries or as elements in other sets (because they are hashable), which regular (mutable) sets cannot.
    ```python
    # Example: Using a frozenset as a dictionary key
    immutable_tags = frozenset(["python", "programming"])
    article_metadata = {immutable_tags: "Introduction to Python"}
    print(f"Article metadata with frozenset key: {article_metadata}") # Output: frozenset({'programming', 'python'}): 'Introduction to Python'}

    # Attempting to modify a frozenset (will raise an AttributeError)
    try:
        immutable_tags.add("tutorial")
    except AttributeError as e:
        print(f"Error trying to modify frozenset: {e}")
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

# Real-world scenario: User input is always a string
# You need to convert it to a numeric type for calculations
user_age_str = input("Enter your age: ") # input() always returns a string
try:
    user_age_int = int(user_age_str)
    print(f"Next year you will be {user_age_int + 1} years old.")
except ValueError:
    print("That's not a valid age!")
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

## Common Mistakes and How to Avoid Them
Understanding common pitfalls can save a lot of debugging time. Here are a few to watch out for:

1.  **Incorrect Variable Naming:** Not following PEP 8 or using non-descriptive names.
    ```python
    # Bad:
    a = 10 # What does 'a' mean?
    UserName = "Bob" # Not snake_case (and typically reserved for class names)

    # Good:
    user_age = 10
    user_name = "Bob"
    ```

2.  **Modifying Immutable Types:** Attempting to change a string or tuple in place.
    ```python
    my_string = "Python"
    try:
        my_string[0] = 'p' # TypeError
    except TypeError as e:
        print(f"Error: {e}")

    my_tuple = (1, 2, 3)
    try:
        my_tuple[0] = 0 # TypeError
    except TypeError as e:
        print(f"Error: {e}")

    # Instead, create a new object:
    my_string = "p" + my_string[1:] # Creates a new string "python"
    print(f"New string: {my_string}")
    ```

3.  **`ValueError` During Type Conversion:** Trying to convert a string that doesn't represent a valid number.
    ```python
    invalid_num_str = "abc"
    try:
        int_val = int(invalid_num_str)
    except ValueError as e:
        print(f"Error: {e}") # Output: Error: invalid literal for int() with base 10: 'abc'
    ```

4.  **Misunderstanding Set Uniqueness/Order:** Expecting sets to maintain insertion order or allow duplicates.
    ```python
    my_set = {1, 2, 3, 2, 1}
    print(my_set) # Output will be {1, 2, 3} (duplicates removed, order not guaranteed)
    ```

## Key Takeaways
*   Variables are named storage locations for values, and Python infers their type.
*   Python has several built-in data types, including numeric (int, float, complex), boolean (bool), sequence (str, list, tuple), mapping (dict), and set (set, frozenset) types.
*   Strings and tuples are **immutable**, meaning their content cannot be changed after creation.
*   Lists, dictionaries, and sets are **mutable**, meaning their content can be modified.
*   The `type()` function checks a variable's data type, and functions like `int()`, `float()`, `str()` can convert between types.
*   F-strings provide an efficient and readable way to format strings.

## Exercises

### Exercise 1: Personal Greeting
Declare a variable for your favorite color and print it.

### Exercise 2: Data Type Exploration
Create variables for different data types (string, integer, float, boolean) and print their types using the `type()` function.

### Exercise 3: Movie List
Create a list of your top 5 favorite movies and print it.

### Exercise 4: Squares with List Comprehension
Use list comprehension to generate a list of squares from 0 to 9 and print the result.

### Exercise 5: Student Profile

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

## Common Pitfalls with Data Types
- **String Manipulation**: Strings are immutable in Python. When trying to change a character in a string, it will raise a TypeError.
- **Type Casting**: Be cautious when converting strings to integers or floats. Ensure that the string represents a valid number to avoid ValueError.