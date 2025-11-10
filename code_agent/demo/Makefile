CC = gcc
CFLAGS = -Wall -Wextra -std=c99 -g

SRCS = prolog_data.c prolog_db.c prolog_unify.c prolog_query.c prolog_parser.c main.c
OBJS = $(SRCS:.c=.o)
TARGET = my_prolog_exe

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) $(OBJS) -o $(TARGET)

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -f $(OBJS) $(TARGET)
