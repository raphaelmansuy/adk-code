# Chapter 5: File I/O and Error Handling

This chapter covers how to interact with files on your computer and how to gracefully handle errors that might occur during program execution.

## 1. File Input/Output (I/O)

Working with files is a common task in programming, allowing you to read data from files or write data to them.

### Opening Files

To work with a file, you first need to open it using the `open()` function. It takes two main arguments: the file path and the mode.

```python
# Syntax:
# open(file_path, mode)
```

**A Note on File Paths:** In the examples, we use simple file names, which assume the file is in the same directory as your Python script. For files in other locations, you'll need to provide a full (absolute) path or a relative path (e.g., `"data/my_file.txt"`). Always be mindful of file paths, especially when your application moves or is deployed.

**Common Modes:**

- `"r"`: Read mode (default). Opens a file for reading. Raises `FileNotFoundError` if the file doesn't exist.
- `"w"`: Write mode. Opens a file for writing. Creates the file if it doesn't exist, or truncates (empties) it if it does.
- `"a"`: Append mode. Opens a file for appending. Creates the file if it doesn't exist. Data is added to the end of the file.
- `"x"`: Exclusive creation mode. Creates a new file, but raises an `FileExistsError` if the file already exists.
- `"t"`: Text mode (default). Handles text data (strings).
- `"b"`: Binary mode. Handles binary data (bytes objects), used for non-text files like images, executables, etc.

  ```python
  # Ensure the directory exists for the binary file example
  directory = "output_data"
  if not os.path.exists(directory):
      os.makedirs(directory)

  binary_file_path = os.path.join(directory, "binary_data.bin")
  # Writing binary data
  with open(binary_file_path, "wb") as file:
      file.write(b"\x01\x02\x03\x04") # b'' prefix denotes a bytes literal
  print(f"Binary data written to {binary_file_path}")

  # Reading binary data
  with open(binary_file_path, "rb") as file:
      data = file.read()
      print(f"Read binary data: {data}") # Output: b'\x01\x02\x03\x04'
  ```

## Working with File Paths and OS Module

The `os` module provides functions for interacting with the operating system, including file system operations. `os.path` is a submodule for path manipulations.

**Important:** Always use `os.path.join()` when constructing file paths to ensure your code works correctly across different operating systems (Windows, macOS, Linux) which use different path separators.

```python
import os

# Get current working directory
current_dir = os.getcwd()
print(f"Current directory: {current_dir}")

# Joining paths (platform-independent)
folder = "my_data"
file_name = "report.txt"
full_path = os.path.join(folder, file_name)
print(f"Joined path: {full_path}")

# Extracting parts of a path
print(f"Basename: {os.path.basename(full_path)}") # Output: report.txt
print(f"Dirname: {os.path.dirname(full_path)}")   # Output: my_data
print(f"Split path: {os.path.split(full_path)}") # Output: ('my_data', 'report.txt')
print(f"Absolute path: {os.path.abspath(full_path)}") # Output: /path/to/current_dir/my_data/report.txt

# Checking if a path exists
if os.path.exists(full_path):
    print(f"{full_path} exists.")
else:
    print(f"{full_path} does not exist.")

# Creating directories (if they don't exist)
if not os.path.exists(folder):
    os.makedirs(folder) # Creates intermediate directories if needed
    print(f"Created directory: {folder}")

# Deleting a file (use with caution!)
# Example: os.remove("path/to/my_file.txt")
# This would delete the file if it exists.

# Renaming a file
# Example: os.rename("old_name.txt", "new_name.txt")
# This would rename 'old_name.txt' to 'new_name.txt'.
```

### Writing to Files (Using the `with` statement)

When working with files, the `with` statement is the **recommended best practice**. It acts as a context manager, ensuring that the file is automatically closed after the block of code is executed, even if errors occur. This prevents resource leaks and potential data corruption, making your code more robust and reliable. You no longer need to explicitly call `file.close()`.

```python
# Example: Ensuring a directory exists before writing

directory = "output_data"
if not os.path.exists(directory):
    os.makedirs(directory) # Create the directory if it doesn't exist

file_path = os.path.join(directory, "my_file.txt")

# Writing to a new file (or overwriting an existing one)
with open(file_path, "w") as file:
    file.write("Hello, Python world!\n")
    file.write("This is a new line.\n")

# Appending to an existing file
with open(file_path, "a") as file:
    file.write("Adding another line at the end.\n")

print(f"Content written to {file_path}")
```

### Reading from Files

```python
# Reading the entire content
with open(file_path, "r") as file:
    content = file.read()
    print("--- Entire Content ---")
    print(content)

# Reading line by line
with open(file_path, "r") as file:
    print("--- Line by Line ---")
    for line in file:
        print(line.strip()) # .strip() removes leading/trailing whitespace, including newline

# Reading a specific number of characters
with open(file_path, "r") as file:
    first_10_chars = file.read(10)
    print("--- First 10 Characters ---")
    print(first_10_chars)
```

### Advanced File I/O: Custom Context Managers (Optional)

You can create your own context managers using the `contextlib` module. This is useful for managing any resource that needs to be set up and torn down, not just files.

```python
from contextlib import contextmanager

@contextmanager
def custom_manager():
    print("Entering the context...")
    yield # Code inside the 'with' block executes here
    print("Exiting the context.")

with custom_manager():
    print("Inside the context.")
```

## 2. Error Handling (Exceptions)

Errors, also known as exceptions, are events detected during execution that interrupt the normal flow of a program. Python provides a way to handle these errors gracefully using `try`, `except`, `else`, and `finally` blocks.

### `try` and `except`

The `try` block contains the code that might raise an exception. The `except` block catches and handles the exception.

```python
try:
    # This code might cause an error
    num1 = int(input("Enter a number: "))
    num2 = int(input("Enter another number: "))
    result = num1 / num2
    print(f"The result is: {result}")
except ValueError:
    print("Invalid input. Please enter a valid number.")
except ZeroDivisionError:
    print("Error: Cannot divide by zero!")
except Exception as e: # Catch any other unexpected error. Always catch specific exceptions first.
    print(f"An unexpected error occurred: {e}")

print("Program continues after error handling.")
```

**Common Built-in Exceptions:**
Python has many built-in exceptions to signal different types of errors. Here are a few common ones:

- `ValueError`: Raised when an operation receives an argument that has the right type but an inappropriate value.
- `TypeError`: Raised when an operation or function is applied to an object of inappropriate type.
- `NameError`: Raised when a local or global name is not found.
- `IndexError`: Raised when a sequence subscript (index) is out of range.
- `KeyError`: Raised when a dictionary key is not found.
- `FileNotFoundError`: Raised when a file or directory is requested but doesn't exist.
- `ZeroDivisionError`: Raised when the second operand of a division or modulo operation is zero.

**Best Practice for `except` blocks:** It's generally good practice to catch more specific exceptions first, followed by more general ones. This allows you to handle different error conditions appropriately. For instance, catch `ValueError` or `ZeroDivisionError` before `Exception`. Catching `Exception` (the base class for all exceptions) should typically be done last, if at all, to avoid masking unexpected errors and making debugging harder.

### `else` Block

The `else` block is executed if no exceptions are raised in the `try` block.

```python
# Example of else block with try-except for file operations

# Ensure the directory exists for the example file
directory = "output_data"
if not os.path.exists(directory):
    os.makedirs(directory)

file_to_check = os.path.join(directory, "existing_file_for_else.txt")

# First, create the file so the else block can execute
with open(file_to_check, "w") as f:
    f.write("This file exists for the else example.")

try:
    with open(file_to_check, "r") as file: # Use with statement inside try
        content = file.read()
        print(f"Content of {file_to_check}: {content}")
except FileNotFoundError:
    print(f"The file {file_to_check} was not found (this should not happen here).")
else:
    print(f"File '{os.path.basename(file_to_check)}' opened and read successfully (else block executed).")
    # No need for file.close() because of 'with' statement

print("-" * 30)

# Example where file is not found (and else block does NOT execute)
try:
    with open(os.path.join(directory, "non_existent_file.txt"), "r") as file:
        content = file.read()
except FileNotFoundError:
    print("Attempted to open a non-existent file: File not found (except block executed).")
else:
    print("This message will NOT be printed if FileNotFoundError occurs.")
```

### `finally` Block

The `finally` block is always executed, regardless of whether an exception occurred or not. It's often used for cleanup operations (like closing files).

```python
try:
    x = 10 / 0 # This will cause a ZeroDivisionError
except ZeroDivisionError:
    print("Caught a ZeroDivisionError.")
finally:
    print("This 'finally' block always executes.")

print("End of program.")
```

Here's a flowchart illustrating the `try-except-finally` flow:

```mermaid
graph TD
    A[Start]
    A --> B{try block}
    B -- No Exception --> C{else block?}
    C -- Yes --> D[else block (If no errors)]
    C -- No --> E[finally block]
    D --> E
    B -- Exception Occurs --> F{except block?}
    F -- Yes --> G[except block (Handle error)]
    F -- No --> E
    G --> E
    E --> H[End]
```

### `raise` Statement

You can explicitly raise an exception in your code using the `raise` statement. This is useful for signaling that an error condition has occurred.

```python
def validate_age(age):
    if not isinstance(age, (int, float)):
        raise TypeError("Age must be a number.")
    if age < 0:
        raise ValueError("Age cannot be negative.")
    if age < 18:
        raise ValueError("Must be at least 18 years old.")
    print("Age is valid.")

try:
    validate_age(20) # This will print "Age is valid."
except (TypeError, ValueError) as e:
    print(f"Validation Error: {e}")

try:
    validate_age(-5) # This will raise a ValueError
except (TypeError, ValueError) as e:
    print(f"Validation Error: {e}")

try:
    validate_age("abc") # This will raise a TypeError
except (TypeError, ValueError) as e:
    print(f"Validation Error: {e}")
```

## Best Practices and Common Pitfalls

To ensure your file operations and error handling are robust and efficient, consider these best practices and avoid common pitfalls:

### Best Practices

1.  **Always Use `with` for File Operations:** The `with` statement guarantees that file resources are properly closed, even if errors occur. This prevents resource leaks and potential data corruption.
2.  **Handle Specific Exceptions First:** In `try-except` blocks, catch more specific exceptions (e.g., `FileNotFoundError`, `ValueError`) before more general ones (e.g., `Exception`). This allows for targeted error recovery.
3.  **Use `os.path.join()` for Paths:** Always construct file paths using `os.path.join()` to ensure your code is platform-independent and works correctly on Windows, macOS, and Linux.
4.  **Check for File/Directory Existence:** Before attempting to read or write, use `os.path.exists()` to check if a file or directory exists. Similarly, use `os.makedirs()` with `exist_ok=True` to safely create directories.
5.  **Provide User-Friendly Error Messages:** When an error occurs, provide clear and helpful messages to the user, guiding them on how to resolve the issue.
6.  **Log Detailed Errors:** For debugging and monitoring, log detailed error information (e.g., using Python's `logging` module) to a file or a monitoring system.

### Common Pitfalls

1.  **Forgetting to Close Files:** Not using the `with` statement can lead to files remaining open, consuming system resources, and potentially causing data loss or corruption if the program crashes.
2.  **Broad `except` Clauses:** Using a bare `except:` or `except Exception:` too broadly can hide unexpected errors, making debugging extremely difficult. Only catch `Exception` as a last resort, and always log it.
3.  **Incorrect File Modes:** Opening a file with the wrong mode (e.g., trying to write to a file opened in `"r"` mode, or reading from a file opened in `"w"` mode) will result in errors.
4.  **Permission Errors:** Attempting to read from or write to files in directories where your program doesn't have the necessary permissions will raise `PermissionError`.
5.  **`FileNotFoundError`:** Trying to open a file for reading that doesn't exist is a common error. Always anticipate this and handle it gracefully.
6.  **Path Inconsistencies:** Hardcoding path separators (e.g., `"folder\\file.txt"` on Windows, `"folder/file.txt"` on Linux) can break your code on different operating systems. Use `os.path.join()`.

## Key Takeaways

- File I/O allows programs to read from and write to files using `open()` with various modes (`"r"`, `"w"`, `"a"`, etc.).
- The `with` statement is crucial for safe file handling, ensuring files are properly closed.
- The `os` module provides functions for interacting with the file system (e.g., `os.path.join`, `os.makedirs`, `os.remove`).
- Error handling uses `try`, `except`, `else`, and `finally` blocks to manage exceptions gracefully.
- `try` contains code that might raise an error; `except` catches specific errors.
- `else` executes if no exceptions occur in `try`; `finally` always executes for cleanup.
- Specific exceptions should be caught before general ones (`Exception`).
- The `raise` statement is used to explicitly trigger exceptions.

## Exercise 1: Log File Processor

Write a Python script that performs the following tasks:

1.  **Create a log file:** Write several lines of text to a file named `application.log`. Each line should simulate a log entry (e.g., "[INFO] User logged in", "[ERROR] Database connection failed").
2.  **Read and filter:** Read the `application.log` file line by line.
3.  **Error Report:** Create a new file called `error_report.txt`. If a line in `application.log` contains the word "ERROR", write that entire line to `error_report.txt`.
4.  **Handle potential errors:** Use `try-except` blocks to handle `FileNotFoundError` if `application.log` doesn't exist when trying to read it.

**Hint:** You'll need to use `open()` with different modes, the `with` statement, string methods like `in` or `find()` to check for "ERROR", and `try-except` for error handling. Remember to use `os.makedirs()` to ensure your output directory exists before writing files.

## Conclusion

This tutorial has covered the fundamentals of Python programming, from basic syntax and data types to control flow, functions, file I/O, and error handling. This knowledge forms a strong foundation for you to build more complex and robust Python applications. Keep practicing, explore Python's extensive libraries, and happy coding!
