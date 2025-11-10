
.PHONY: all clean help run test

# Compiler and flags
CC = gcc
CFLAGS = -Wall -Wextra -std=c99 -I.
LDFLAGS = 

VPATH = demo

# Source files
SRCS = main.c term.c clause.c knowledge_base.c substitution.c parser.c unification.c inference.c

# Object files
OBJS = $(SRCS:.c=.o)

# Executable name
TARGET = prolog

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[0;33m
RED = \033[0;31m
NC = \033[0m # No Color

all: $(TARGET)
	@echo "$(GREEN)Build successful!$(NC)"

$(TARGET): $(OBJS)
	@echo "$(YELLOW)Linking...$(NC)"
	$(CC) $(OBJS) $(LDFLAGS) -o $(TARGET)
	@echo "$(GREEN)Executable created: $(TARGET)$(NC)"

%.o: %.c
	@echo "$(YELLOW)Compiling $<...$(NC)"
	$(CC) $(CFLAGS) -c $< -o $@

run: $(TARGET)
	@echo "$(YELLOW)Running $(TARGET)...$(NC)"
	@./$(TARGET)

test: $(TARGET)
	@echo "$(YELLOW)Running tests with input_unification.txt...$(NC)"
	@./$(TARGET) input_unification.txt
	@echo "$(YELLOW)Running tests with input.txt...$(NC)"
	@./$(TARGET) input.txt

clean:
	@echo "$(RED)Cleaning up...$(NC)"
	@rm -f $(OBJS) $(TARGET)
	@echo "$(GREEN)Cleanup complete.$(NC)"

help:
	@echo "$(GREEN)Makefile for Prolog Interpreter Demo$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make             - Builds the $(TARGET) executable."
	@echo "  make run         - Runs the compiled Prolog interpreter."
	@echo "  make test        - Runs predefined tests using input files."
	@echo "  make clean       - Removes all compiled object files and the executable."
	@echo "  make help        - Displays this help message."
