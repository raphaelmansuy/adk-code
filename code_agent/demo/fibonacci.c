#include <stdio.h>

/**
 * fibonacci - Calculates the nth Fibonacci number using recursion
 * @n: The position in the Fibonacci sequence (0-indexed)
 * 
 * Returns: The nth Fibonacci number
 * 
 * Note: This is a naive recursive implementation with O(2^n) time complexity.
 *       For large values of n, consider using memoization or iterative approach.
 */
int fibonacci(int n) {
  if (n <= 1) {
    return n;
  }
  return fibonacci(n - 1) + fibonacci(n - 2);
}

/**
 * main - Entry point of the program
 * 
 * Prints the first 10 numbers in the Fibonacci sequence.
 * 
 * Returns: 0 on success
 */
int main() {
  int n = 10;
  printf("Fibonacci sequence up to %d:\n", n);
  for (int i = 0; i < n; i++) {
    printf("%d ", fibonacci(i));
  }
  printf("\n");
  return 0;
}