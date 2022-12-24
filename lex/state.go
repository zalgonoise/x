package lex

type StateFn[C comparable, T any, I Item[C, T]] func(l Lexer[C, T, I]) StateFn[C, T, I]
