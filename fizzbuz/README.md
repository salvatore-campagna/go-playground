# FizzBuzz

## Introduction

FizzBuzz is a classic programming challenge used to test basic programming skills. The goal is to print a sequence of numbers with some substitutions:
- For multiples of 3, print "Fizz" instead of the number.
- For multiples of 5, print "Buzz" instead of the number.
- For numbers that are multiples of both 3 and 5, print "FizzBuzz" instead of the number.
  
The name "FizzBuzz" reflects these substitutions, which are often used as a fun way to introduce loops, conditionals, and basic logic.

## Instructions

The task is to generate a sequence of FizzBuzz values up to a specified number.

For each integer from 1 up to the given number `n`:
- If the number is divisible by 3, add "Fizz" to the output.
- If the number is divisible by 5, add "Buzz" to the output.
- If the number is divisible by both 3 and 5, add "FizzBuzz" to the output.
- If the number is not divisible by 3 or 5, add the number itself as a string.

Each entry in the sequence should be followed by a newline character (`\n`).

## Example Usage

If `n = 5`, the result should be:

```text
1
2
Fizz
4
Buzz
