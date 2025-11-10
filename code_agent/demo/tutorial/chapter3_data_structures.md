# Chapter 3: Data Structures

## Learning Objectives
In this chapter, you will:
*   Understand and utilize Python's fundamental data structures: Lists, Tuples, Dictionaries, and Sets.
*   Learn how to create, access, modify, and manipulate these data structures.
*   Discover practical use cases for each data structure.
*   Get an introduction to list and dictionary comprehensions for concise data manipulation.

Python offers several built-in data structures that are essential for organizing and managing data efficiently. This chapter will cover the most commonly used ones: Lists, Tuples, Dictionaries, and Sets.

## Lists

A list is an ordered, mutable (changeable) collection of items. Lists are defined by enclosing elements in square brackets `[]`, with elements separated by commas. They can contain items of different data types.

### Creation and Access

Lists are ordered collections, meaning items have a defined order, and that order will not change. You can access list items by referring to their index number. Indices start at 0 for the first item.

```text
+---+---+---+-------+--------+------+
| 1 | 2 | 3 | apple | banana | True |
+---+---+---+-------+--------+------+
  0   1   2     3        4      5
```

```python
# Creating a list
my_list = [1, 2, 3, "apple", "banana", True]
print(my_list) # Output: [1, 2, 3, 'apple', 'banana', True]

# Accessing elements (indexing starts from 0)
print(my_list[0]) # Output: 1
print(my_list[3]) # Output: apple
print(my_list[-1]) # Output: True (last element)

# Slicing a list [start:end:step]
print(my_list[1:4]) # Output: [2, 3, 'apple'] (elements from index 1 up to, but not including, 4)
print(my_list[:3])  # Output: [1, 2, 3] (from beginning to index 2)
print(my_list[3:])  # Output: ['apple', 'banana', True] (from index 3 to end)
print(my_list[::2]) # Output: [1, 'apple', True] (every second element)
```

### Modification

```python
my_list[0] = 100
print(my_list) # Output: [100, 2, 3, 'apple', 'banana', True]

# Adding elements
my_list.append("orange") # Adds to the end
print(my_list) # Output: [100, 2, 3, 'apple', 'banana', True, 'orange']

my_list.insert(1, "new_item") # Inserts at a specific index
print(my_list) # Output: [100, 'new_item', 2, 3, 'apple', 'banana', True, 'orange']

# Removing elements
my_list.remove("apple") # Removes the first occurrence of the value
print(my_list) # Output: [100, 'new_item', 2, 3, 'banana', True, 'orange']

popped_item = my_list.pop() # Removes and returns the last item
print(popped_item) # Output: orange
print(my_list) # Output: [100, 'new_item', 2, 3, 'banana', True]

del my_list[0] # Deletes item at a specific index
print(my_list) # Output: ['new_item', 2, 3, 'banana', True]
```

### Common List Operations

Python provides several built-in functions and methods for working with lists:

```python
numbers = [10, 1, 8, 3, 5]

print(len(numbers))    # Output: 5 (number of elements)
print(min(numbers))    # Output: 1 (minimum value)
print(max(numbers))    # Output: 10 (maximum value)
print(sum(numbers))    # Output: 27 (sum of all elements)

numbers.sort()         # Sorts the list in-place (modifies the original list)
print(numbers)         # Output: [1, 3, 5, 8, 10]

sorted_numbers = sorted([10, 1, 8, 3, 5]) # Returns a new sorted list, doesn't modify original
print(sorted_numbers)  # Output: [1, 3, 5, 8, 10]

# Checking if an item is in a list
print(3 in numbers)    # Output: True
print(9 in numbers)    # Output: False
```

## Tuples

A tuple is an ordered, immutable (unchangeable) collection of items. Tuples are defined by enclosing elements in parentheses `()`. Once a tuple is created, its elements cannot be changed, added, or removed.

```python
# Creating a tuple
my_tuple = (1, 2, "hello", 3.14)
print(my_tuple) # Output: (1, 2, 'hello', 3.14)

# Accessing elements (same as lists)
print(my_tuple[0]) # Output: 1

# Attempting to modify a tuple will result in an error
# my_tuple[0] = 10 # TypeError: 'tuple' object does not support item assignment

# Use cases: fixed collections of items, function return values
def get_coordinates():
    return (10, 20)

x, y = get_coordinates()
print(f"X: {x}, Y: {y}") # Output: X: 10, Y: 20
```

### Tuple Unpacking

Tuple unpacking allows you to assign the elements of a tuple (or any iterable) to multiple variables in a single statement. This is a very convenient feature.

```python
coordinates = (100, 200)
x_coord, y_coord = coordinates
print(f"X Coordinate: {x_coord}, Y Coordinate: {y_coord}") # Output: X Coordinate: 100, Y Coordinate: 200

# Swapping variables easily
a = 5
b = 10
a, b = b, a
print(f"a: {a}, b: {b}") # Output: a: 10, b: 5

# Unpacking with a function call
def get_user_info():
    return ("Alice", 30, "Engineer")

name, age, job = get_user_info()
print(f"Name: {name}, Age: {age}, Job: {job}")
```

## Dictionaries

A dictionary is an unordered, mutable collection of key-value pairs. Each key must be unique and immutable (e.g., strings, numbers, tuples). Values can be of any data type.

### Creation and Access

Dictionaries are collections of key-value pairs, where each unique key maps to a specific value. They are unordered and mutable.

```text
+----------+-------+
|   Key    | Value |
+----------+-------+
| "name"   | "Alice" |
| "age"    | 30    |
| "city"   | "New York" |
+----------+-------+
```

```python
# Creating a dictionary
my_dict = {
    "name": "Alice",
    "age": 30,
    "city": "New York"
}
print(my_dict) # Output: {'name': 'Alice', 'age': 30, 'city': 'New York'}

# Accessing values using keys
print(my_dict["name"]) # Output: Alice

# Using .get() method (safer, returns None if key not found)
print(my_dict.get("age")) # Output: 30
print(my_dict.get("country")) # Output: None
print(my_dict.get("country", "Unknown")) # Output: Unknown (default value if key not found)
```

### Modification

```python
# Adding a new key-value pair
my_dict["occupation"] = "Engineer"
print(my_dict) # Output: {'name': 'Alice', 'age': 30, 'city': 'New York', 'occupation': 'Engineer'}

# Modifying an existing value
my_dict["age"] = 31
print(my_dict) # Output: {'name': 'Alice', 'age': 31, 'city': 'New York', 'occupation': 'Engineer'}

# Removing a key-value pair
del my_dict["city"]
print(my_dict) # Output: {'name': 'Alice', 'age': 31, 'occupation': 'Engineer'}

popped_value = my_dict.pop("occupation") # Removes and returns the value for the given key
print(popped_value) # Output: Engineer
print(my_dict) # Output: {'name': 'Alice', 'age': 31}
```

## Sets

A set is an unordered collection of unique items. Sets are useful for mathematical set operations like union, intersection, and difference, and for efficiently checking membership.

### Creation and Operations

```python
# Creating a set
my_set = {1, 2, 3, 3, 4}
print(my_set) # Output: {1, 2, 3, 4} (duplicates are automatically removed)

# Creating an empty set
empty_set = set() # {} creates an empty dictionary, not an empty set
print(empty_set)

# Adding elements
my_set.add(5)
print(my_set) # Output: {1, 2, 3, 4, 5}

# Removing elements
my_set.remove(3)
print(my_set) # Output: {1, 2, 4, 5}

# Set operations
set_a = {1, 2, 3, 4}
set_b = {3, 4, 5, 6}

print(set_a.union(set_b))        # Output: {1, 2, 3, 4, 5, 6}
print(set_a.intersection(set_b)) # Output: {3, 4}
print(set_a.difference(set_b))   # Output: {1, 2}
print(set_b.difference(set_a))   # Output: {5, 6}

# Checking membership (very efficient)
print(2 in set_a) # Output: True
print(7 in set_a) # Output: False
```

## List Comprehensions (Brief Introduction)

List comprehensions provide a concise way to create lists. They consist of brackets containing an expression followed by a `for` clause, then zero or more `for` or `if` clauses.

```python
# Traditional loop to create a list of squares
squares = []
for i in range(5):
    squares.append(i * i)
print(squares) # Output: [0, 1, 4, 9, 16]

# List comprehension for the same task
squares_comp = [i * i for i in range(5)]
print(squares_comp) # Output: [0, 1, 4, 9, 16]

# List comprehension with a condition
even_numbers = [i for i in range(10) if i % 2 == 0]
print(even_numbers) # Output: [0, 2, 4, 6, 8]
```

## Dictionary Comprehensions (Brief Introduction)

Similar to list comprehensions, dictionary comprehensions provide a concise way to create dictionaries.

```python
# Creating a dictionary of squares
squares_dict = {i: i*i for i in range(5)}
print(squares_dict) # Output: {0: 0, 1: 1, 2: 4, 3: 9, 4: 16}

# Dictionary comprehension with a condition
even_squares_dict = {i: i*i for i in range(10) if i % 2 == 0}
print(even_squares_dict) # Output: {0: 0, 2: 4, 4: 16, 6: 36, 8: 64}
```

## Exercises

1.  **List Manipulation:** Create a list of your 5 favorite movies. Add a new movie, remove one, and then print the list sorted alphabetically.
2.  **Tuple Unpacking:** Create a tuple containing a person's name, age, and city. Unpack these values into separate variables and print them.
3.  **Dictionary of Students:** Create a dictionary where keys are student names (strings) and values are their scores (integers). Add a new student, update an existing student's score, and print the names of all students who scored above 90.
4.  **Set Operations:** Given two sets of numbers, find their union, intersection, and difference.
5.  **List Comprehension Challenge:** Use a list comprehension to create a list of all numbers from 1 to 20 that are divisible by both 2 and 3.

This concludes Chapter 3. You now have a solid foundation in Python's fundamental data structures, which are crucial for building complex applications.