#include <stdio.h>
#include <stdlib.h>

// Function to perform calculation
int calculate(double num1, char operator, double num2, double *result) {
    switch (operator) {
        case '+':
            *result = num1 + num2;
            break;
        case '-':
            *result = num1 - num2;
            break;
        case '*':
            *result = num1 * num2;
            break;
        case '/':
            if (num2 == 0) {
                fprintf(stderr, "Error: Division by zero\n");
                return 0; // Indicate error
            }
            *result = num1 / num2;
            break;
        case '%':
            if (num2 == 0) {
                fprintf(stderr, "Error: Modulo by zero\n");
                return 0; // Indicate error
            }
            *result = (double)((long long)num1 % (long long)num2);
            break;
        default:
            fprintf(stderr, "Error: Invalid operator '%c'\n", operator);
            return 0; // Indicate error
    }
    return 1; // Indicate success
}

int main(int argc, char *argv[]) {
    if (argc != 2) {
        fprintf(stderr, "Usage: calculate expression\n");
        return 1;
    }

    char *expression = argv[1];
    double num1, num2, result;
    char operator;
    int num_found = sscanf(expression, "%lf%c%lf", &num1, &operator, &num2);

    if (num_found != 3) {
        fprintf(stderr, "Error: Invalid expression format. Expected 'number operator number'.\n");
        return 1;
    }

    if (!calculate(num1, operator, num2, &result)) {
        return 1; // Calculation failed
    }

    printf("%f\n", result);
    return 0;
}
