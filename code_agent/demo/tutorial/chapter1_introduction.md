# Chapter 1: Introduction to Python

## Learning Objectives
In this chapter, you will:
*   Understand what Python is and why it's a popular programming language.
*   Learn how to set up your Python development environment.
*   Write and run your first Python program.
*   Familiarize yourself with basic data types and variables.

## What is Python?

Python is a high-level, interpreted, general-purpose programming language. Created by Guido van Rossum and first released in 1991, Python's design philosophy emphasizes code readability with its notable use of significant indentation. It is dynamically typed and garbage-collected. It supports multiple programming paradigms, including structured (particularly procedural), object-oriented, and functional programming.

## Why Learn Python?

Python's popularity stems from several key advantages:

*   **Simplicity and Readability:** Python's syntax is clear and intuitive, making it easy to learn for beginners and highly readable for experienced developers.
*   **Versatility:** It's used in web development (Django, Flask), data science (Pandas, NumPy, Scikit-learn), machine learning (TensorFlow, PyTorch), artificial intelligence, automation, scripting, scientific computing, game development, and more.
*   **Large Community and Ecosystem:** Python boasts a vast and active community, leading to abundant resources, tutorials, and libraries. The Python Package Index (PyPI) hosts over 350,000 projects, providing tools for almost any task.
*   **Cross-Platform:** Python runs on various operating systems, including Windows, macOS, and Linux.

## Setting Up Your Environment

While Python can be installed directly, using a distribution like Anaconda or Miniconda is highly recommended, especially for data science and machine learning. These distributions come with Python and many essential libraries pre-installed.

1.  **Download:** Visit the Anaconda or Miniconda website and download the installer for your operating system.
2.  **Install:** Follow the installation instructions.
3.  **Verify:** Open your terminal or command prompt and type `python --version` to ensure Python is installed correctly. You should see a version number (e.g., Python 3.9.7).

For coding, a good Integrated Development Environment (IDE) or text editor is crucial. Visual Studio Code (VS Code) with the Python extension is a popular and powerful choice.

## The Interactive Python Interpreter

Before writing your first program, it's helpful to know about the interactive Python interpreter. You can use it to execute Python code line by line and see immediate results. It's an excellent tool for experimenting and testing small snippets of code.

To access it, simply open your terminal or command prompt and type `python` (or `python3` on some systems) and press Enter.

```bash
python
```

You will see a prompt like `>>>`. You can now type Python code directly:

```python
>>> print("Hello from the interpreter!")
Hello from the interpreter!
>>> 2 + 3
5
>>> name = "Alice"
>>> name
'Alice'
```

To exit the interpreter, type `exit()` and press Enter, or press `Ctrl+D` (on Linux/macOS) or `Ctrl+Z` then Enter (on Windows).

## Your First Python Program: "Hello, World!"

Let's write the classic "Hello, World!" program. The following diagram illustrates the simple workflow:

```text
+-------------------+     +--------------------+     +------------------------+
|       Start       | --> |  Open Text Editor  | --> | Write print("Hello...")|
+-------------------+     +--------------------+     +------------------------+
          |                                                       |
          v                                                       v
+-------------------+     +--------------------+     +------------------------+
|  Save as hello.py | --> |   Open Terminal    | --> |  Navigate to Directory |
+-------------------+     +--------------------+     +------------------------+
          |                                                       |
          v                                                       v
+-------------------+     +--------------------+     +------------------------+
| Run: python hello.py | --> | See "Hello, World!" Output | --> |          End           |
+-------------------+     +--------------------+     +------------------------+
```

Open a text editor, type the following line, and save it as `hello.py`.

```python
print("Hello, World!")
```

To run it, open your terminal or command prompt, navigate to the directory where you saved `hello.py`, and type:

```bash
python hello.py
```

You should see:

```
Hello, World!
```

The `print()` function is a built-in Python function that outputs text to the console.

## Basic Data Types and Variables

Python handles various types of data. Here are some fundamental ones:

*   **Integers (`int`):** Whole numbers (e.g., `10`, `-5`, `0`).
*   **Floating-Point Numbers (`float`):** Numbers with a decimal point (e.g., `3.14`, `-0.5`, `2.0`).
*   **Strings (`str`):** Sequences of characters, enclosed in single or double quotes (e.g., `"Hello"`, `'Python'`, `"123"`).
*   **Booleans (`bool`):** Represent truth values, either `True` or `False`.

**Variables** are used to store data. In Python, you don't need to declare the type of a variable; Python infers it dynamically.

```python
# Integer variable
age = 30
print(f"Age: {age}, Type: {type(age)}") # Output: Age: 30, Type: <class 'int'>

# Float variable
pi = 3.14159
print(f"PI: {pi}, Type: {type(pi)}") # Output: PI: 3.14159, Type: <class 'float'>

# String variable
name = "Alice"
message = 'Hello, ' + name + '!'
print(f"Message: {message}, Type: {type(message)}") # Output: Message: Hello, Alice!, Type: <class 'str'>

# Boolean variable
is_student = True
print(f"Is student: {is_student}, Type: {type(is_student)}") # Output: Is student: True, Type: <class 'bool'>

# Demonstrating type conversion
num_str = "42"
num_int = int(num_str)
print(f"Converted string to int: {num_int}, Type: {type(num_int)}") # Output: Converted string to int: 42, Type: <class 'int'>

float_num = 15.7
int_from_float = int(float_num)
print(f"Converted float to int: {int_from_float}, Type: {type(int_from_float)}") # Output: Converted float to int: 15, Type: <class 'int'>

# Complex assignment and operations
a = 10
b = 20
c = a + b
print(f"Sum of a and b: {c}") # Output: Sum of a and b: 30

greeting = "Good"
time = "Morning"
full_greeting = greeting + " " + time
print(f"Full greeting: {full_greeting}") # Output: Full greeting: Good Morning

# More examples of basic data types
price = 19.99
quantity = 3
product_name = "Laptop"
is_available = True

print(f"Product: {product_name}, Price: ${price}, Quantity: {quantity}, Available: {is_available}")

# Reassigning variables
age = 31
print(f"New age: {age}")

# Multiple assignment
x, y, z = 10, 20, 30
print(f"x: {x}, y: {y}, z: {z}")

# Swapping values
x, y = y, x
print(f"Swapped x: {x}, swapped y: {y}")
```

## Exercises

1.  **Experiment with the Interpreter:** Open your Python interpreter and try printing your name, performing some simple arithmetic operations, and assigning values to new variables.
2.  **Create a Greeting Program:** Write a Python program named `greeting.py` that stores your favorite color in a variable and then prints a message like: "My favorite color is [your color]!"
3.  **Variable Types:** Predict the output of `type()` for the following values: `100`, `"25"`, `10.5 + 2`, `False`.

This concludes Chapter 1. You've taken your first steps into the world of Python programming!