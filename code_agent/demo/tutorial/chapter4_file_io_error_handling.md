# Chapter 4: File I/O and Error Handling

## Learning Objectives
In this chapter, you will:
*   Learn to read from and write to files using various modes.
*   Understand the importance of the `with` statement for file operations.
*   Explore basic file path manipulation with the `os` module.
*   Master `try`, `except`, and `finally` blocks for robust error handling.
*   Get an introduction to creating and raising custom exceptions.

This chapter covers how to interact with files on your computer (File Input/Output) and how to gracefully handle errors that might occur during your program's execution.

## File I/O: Reading and Writing Files

Working with files is a common task in programming, whether it's reading configuration, processing data, or saving results.

### Opening Files

To interact with a file, you first need to open it using the `open()` function. This function returns a file object.

`open(file_path, mode)`

Common modes:
*   `"r"`: Read (default). The file must exist.
*   `"w"`: Write. Creates a new file or truncates (empties) an existing one. **Use with caution!**
*   `"a"`: Append. Creates a new file or adds content to the end of an existing one.
*   `"x"`: Create. Creates a new file, raises an error if the file already exists.
*   `"t"`: Text mode (default). Reads/writes strings.
*   `"b"`: Binary mode. Reads/writes bytes.

### Writing to Files

```python
# Writing to a new file (or overwriting an existing one)
with open("my_file.txt", "w") as file:
    file.write("Hello, Python world!\n")
    file.write("This is a new line.\n")

# Appending to a file
with open("my_file.txt", "a") as file:
    file.write("This line was appended.\n")

print("Content written to my_file.txt")
```

**The `with` statement:** This is the recommended way to work with files. It ensures that the file is properly closed even if errors occur, preventing resource leaks.

### Reading from Files

```python
# Reading the entire file content
with open("my_file.txt", "r") as file:
    content = file.read()
    print("\n--- Entire File Content ---")
    print(content)

# Reading line by line
with open("my_file.txt", "r") as file:
    print("\n--- Reading Line by Line ---")
    for line in file:
        print(line.strip()) # .strip() removes leading/trailing whitespace, including newline characters

# Reading a specific number of characters
with open("my_file.txt", "r") as file:
    print("\n--- Reading 10 characters ---")
    first_10_chars = file.read(10)
    print(first_10_chars)

# Reading all lines into a list
with open("my_file.txt", "r") as file:
    lines = file.readlines()
    print("\n--- Reading all lines into a list ---")
    for line in lines:
        print(line.strip())
```

### Working with File Paths (`os` module)

The `os` module provides a way of using operating system dependent functionality, like reading or writing to the file system. It's particularly useful for path manipulation to ensure your code works across different operating systems (Windows, macOS, Linux).

```python
import os

# Get current working directory
current_dir = os.getcwd()
print(f"Current working directory: {current_dir}")

# Joining paths (platform-independent)
file_name = "report.txt"
full_path = os.path.join(current_dir, "data", file_name)
print(f"Full path to report: {full_path}")

# Checking if a path exists
print(f"Does 'my_file.txt' exist? {os.path.exists("my_file.txt")}")

# Getting the directory name and base name
path_example = "/users/documents/file.txt"
dir_name = os.path.dirname(path_example)
base_name = os.path.basename(path_example)
print(f"Directory: {dir_name}, File: {base_name}")

# Creating directories
# os.makedirs("new_directory/sub_directory", exist_ok=True)
# print("Directories created (if they didn't exist).")
```

## Error Handling: `try`, `except`, `finally`

Errors (or exceptions) are events that disrupt the normal flow of a program. Python provides a robust mechanism to handle these errors gracefully, preventing your program from crashing.

### `try` and `except`

The `try` block lets you test a block of code for errors. The `except` block lets you handle the error.

```python
# Example 1: Division by zero
try:
    result = 10 / 0
    print(result)
except ZeroDivisionError:
    print("Error: Cannot divide by zero!")

# Example 2: Invalid file access
try:
    with open("non_existent_file.txt", "r") as file:
        content = file.read()
        print(content)
except FileNotFoundError:
    print("Error: The file was not found.")

# Example 3: Handling multiple specific exceptions
try:
    num1 = int(input("Enter a number: "))
    num2 = int(input("Enter another number: "))
    division = num1 / num2
    print(f"The division result is: {division}")
except ValueError:
    print("Invalid input. Please enter integers only.")
except ZeroDivisionError:
    print("Cannot divide by zero!")
except Exception as e: # Catch-all for any other exceptions
    print(f"An unexpected error occurred: {e}")
```

### Custom Exceptions

Sometimes, you might want to define your own exception types to make your error handling more specific and descriptive. Custom exceptions are typically created by inheriting from Python's built-in `Exception` class.

```python
class InvalidInputError(Exception):
    """Custom exception for invalid user input."""
    def __init__(self, message="Invalid input provided."):
        self.message = message
        super().__init__(self.message)

def process_data(value):
    if not isinstance(value, int) or value < 0:
        raise InvalidInputError("Input must be a non-negative integer.")
    return value * 2

try:
    # process_data("hello")
    # process_data(-5)
    result = process_data(10)
    print(f"Processed result: {result}")
except InvalidInputError as e:
    print(f"Custom Error: {e}")
except Exception as e:
    print(f"An unexpected error occurred: {e}")
```

It's good practice to handle specific exceptions rather than using a broad `except Exception` unless absolutely necessary, as it can hide other important errors.

### `finally` Block

The `finally` block is always executed, regardless of whether an exception occurred in the `try` block or not. It's often used for cleanup operations (like closing files or releasing resources) that must happen.

```python
try:
    f = open("another_file.txt", "w")
    f.write("Some content.")
except IOError:
    print("Error: Could not write to file.")
finally:
    if 'f' in locals() and not f.closed: # Check if file object was created and is open
        f.close()
        print("File closed in finally block.")
    else:
        print("File was not opened or already closed.")

# A better way to ensure file closure is using 'with' statement, as seen above.
# The 'with' statement implicitly handles the finally block for file objects.
```

## Exercises

1.  **File Copy:** Write a Python script that reads the content of `my_file.txt` (created in this chapter) and writes it to a new file named `my_file_copy.txt`.
2.  **Number Averager:** Write a program that asks the user to enter numbers, one per line. When the user enters "done", calculate and print the average of the numbers. Handle `ValueError` if the user enters non-numeric input.
3.  **Safe File Read:** Create a function `read_safe(filepath)` that attempts to read a file. If `FileNotFoundError` occurs, it should print a user-friendly message and return an empty string. Otherwise, it should return the file's content.
4.  **Custom Age Error:** Define a custom exception `InvalidAgeError`. Write a function `set_age(age)` that raises this error if `age` is not a positive integer. Test it with valid and invalid inputs.
5.  **Log File Creator:** Write a script that appends a timestamped log message to a file named `app_log.txt` each time it's run. Make sure the file is always properly closed.

This concludes Chapter 4. You now understand how to perform file operations and build more robust applications by handling potential errors gracefully.