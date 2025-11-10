# Chapter 5: Modules, Packages, and Object-Oriented Programming (OOP) Basics

## Learning Objectives
In this chapter, you will:
*   Learn to organize your code using modules and packages.
*   Understand the core concepts of Object-Oriented Programming (OOP): classes, objects, attributes, and methods.
*   Get an introduction to encapsulation and inheritance in Python.
*   Appreciate the benefits of modularity and OOP for larger projects.

This final chapter introduces you to organizing your Python code with modules and packages, and gives you a fundamental understanding of Object-Oriented Programming (OOP).

## Modules and Packages

As your programs grow, you'll want to organize your code into separate files for better readability, reusability, and maintainability. This is where modules and packages come in.

### Modules

A **module** is simply a Python file (`.py`) containing Python definitions and statements. When you import a module, you can use the functions, classes, and variables defined within it.

Let's create a simple module. Save the following as `my_math.py`:

```python
# my_math.py
def add(a, b):
    return a + b

def subtract(a, b):
    return a - b

PI = 3.14159
```

Now, in another Python file (e.g., `main.py`) in the same directory, you can import and use it:

```python
# main.py
import my_math

result_add = my_math.add(10, 5)
print(f"Addition: {result_add}") # Output: Addition: 15

result_sub = my_math.subtract(10, 5)
print(f"Subtraction: {result_sub}") # Output: Subtraction: 5

print(f"Value of PI: {my_math.PI}") # Output: Value of PI: 3.14159

# You can also import specific items from a module
from my_math import add, PI
print(f"Add using direct import: {add(2, 3)}") # Output: Add using direct import: 5
print(f"PI using direct import: {PI}") # Output: PI using direct import: 3.14159

# Or import all (generally discouraged for larger projects)
# from my_math import *
```

### Packages

A **package** is a way of organizing related modules into a directory hierarchy. A package is essentially a directory containing a special file named `__init__.py` (which can be empty) and other module files or sub-packages.

Consider this structure:

```
my_project/
├── main.py
└── calculations/
    ├── __init__.py
    ├── basic_ops.py
    └── advanced_ops.py
```

The `__init__.py` file: This special file tells Python that the directory it contains should be treated as a package. It can be an empty file, but it's often used to perform package-level initialization, like importing sub-modules into the package's namespace or defining `__all__` for `from package import *` statements.

For example, if you wanted `from calculations import add` to work, you could add `from .basic_ops import add` to your `calculations/__init__.py`.

`calculations/basic_ops.py`:
```python
# basic_ops.py
def add(a, b):
    return a + b
```

`calculations/advanced_ops.py`:
```python
# advanced_ops.py
def power(base, exp):
    return base ** exp
```

`main.py`:
```python
# main.py
from calculations import basic_ops
from calculations.advanced_ops import power

print(basic_ops.add(7, 3)) # Output: 10
print(power(2, 4))       # Output: 16
```

## Object-Oriented Programming (OOP) Basics

Object-Oriented Programming (OOP) is a programming paradigm based on the concept of "objects", which can contain data (attributes) and code (methods). It helps in structuring programs to be more modular, reusable, and maintainable.

### Classes and Objects

*   **Class:** A blueprint or a template for creating objects. It defines the attributes (data) and methods (functions) that objects of that class will have.
*   **Object:** An instance of a class. When you create an object, you are creating a specific entity based on the class blueprint.

To define a class, use the `class` keyword:

```python
class Dog:
    # Class attribute (shared by all instances)
    species = "Canis familiaris"

    # The __init__ method is a special method (constructor)
    # It's called automatically when a new object is created
    def __init__(self, name, age):
        self.name = name # Instance attribute
        self.age = age   # Instance attribute

    # Instance method
    def bark(self):
        return f"{self.name} says Woof!"

    # Another instance method
    def get_age_in_dog_years(self):
        return self.age * 7

# Creating objects (instances) of the Dog class
my_dog = Dog("Buddy", 3)
your_dog = Dog("Lucy", 5)

# Accessing attributes
print(f"My dog's name is {my_dog.name} and he is {my_dog.age} years old.")
print(f"Your dog's name is {your_dog.name} and she is {your_dog.age} years old.")

# Accessing class attribute
print(f"Buddy is a {my_dog.species}.")
print(f"Lucy is a {Dog.species}.")

# Calling methods
print(my_dog.bark()) # Output: Buddy says Woof!
print(your_dog.get_age_in_dog_years()) # Output: 35

# The 'self' parameter:
# When you call a method like my_dog.bark(), Python automatically passes the object (my_dog) as the first argument to the bark method.
# This argument is conventionally named 'self', and it allows the method to access the object's attributes and other methods.
```

### Encapsulation: Public and "Private" Attributes

Encapsulation is the bundling of data (attributes) and methods that operate on the data into a single unit (the class). It also restricts direct access to some of an object's components, which is a means of preventing accidental interference and misuse of the data.

In Python, there isn't strict "private" access like in some other languages. By convention:
*   **Public attributes/methods:** Accessible from anywhere.
*   **Protected attributes/methods:** Start with a single underscore (e.g., `_attribute`). Indicates to developers that it's intended for internal use within the class or its subclasses, but can still be accessed.
*   **Private attributes/methods:** Start with double underscores (e.g., `__attribute`). Python *mangles* these names to make them harder to access directly from outside the class, but it's not truly private.

```python
class BankAccount:
    def __init__(self, balance):
        self.__balance = balance # "Private" attribute by convention

    def deposit(self, amount):
        if amount > 0:
            self.__balance += amount
            print(f"Deposited {amount}. New balance: {self.__balance}")
        else:
            print("Deposit amount must be positive.")

    def get_balance(self):
        return self.__balance

account = BankAccount(100)
account.deposit(50)
# print(account.__balance) # This would raise an AttributeError due to name mangling
print(account.get_balance()) # Access balance via a public method
```

### Inheritance: Building on Existing Classes

Inheritance is a mechanism that allows a new class (subclass/derived class) to inherit attributes and methods from an existing class (superclass/base class). This promotes code reuse and establishes a natural hierarchy. The following class diagram illustrates this concept:

```text
      +--------+
      | Animal |
      +--------+
      | + name |
      | + speak() |
      +----^---^
           |   |
           |   |
  +--------+---+--------+
  | Dog    |   | Cat   |
  +--------+   +-------+
  | + speak()|   | + speak()|
  +--------+   +-------+
```

```python
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        raise NotImplementedError("Subclass must implement abstract method")

class Dog(Animal):
    def speak(self):
        return f"{self.name} barks."

class Cat(Animal):
    def speak(self):
        return f"{self.name} meows."

my_dog = Dog("Buddy")
my_cat = Cat("Whiskers")

print(my_dog.speak())  # Output: Buddy barks.
print(my_cat.speak())  # Output: Whiskers meows.

# Polymorphism (briefly):
# The ability of different objects to respond to the same method call in their own way.
# Here, my_dog.speak() and my_cat.speak() both call 'speak', but produce different results.
```

## Exercises

1.  **Create a Module:** Create a module `geometry.py` with functions `area_circle(radius)` and `circumference_circle(radius)`. Then, in a `main.py` file, import and use these functions.
2.  **Simple Class:** Create a class `Car` with attributes `make`, `model`, and `year`. Add a method `display_info()` that prints the car's details. Create a `Car` object and call its method.
3.  **Inheritance Challenge:** Create a base class `Shape` with a method `area()` that raises a `NotImplementedError`. Then create two subclasses, `Rectangle` (with `width` and `height`) and `Circle` (with `radius`), each implementing their own `area()` method.
4.  **Package Structure:** Create a package named `utilities` with two modules: `string_utils.py` (with a function `reverse_string(s)`) and `list_utils.py` (with a function `get_max(l)`). Demonstrate importing and using them in a `main.py` file.
5.  **Private Attribute Simulation:** In the `BankAccount` example, add a `withdraw` method. Ensure that the balance cannot go below zero and that the `__balance` is only modified via methods.

This concludes Chapter 5 and your introductory Python tutorial. You've learned the fundamentals of Python programming, from basic syntax to organizing your code and understanding the basics of OOP. Keep practicing and exploring to deepen your knowledge!