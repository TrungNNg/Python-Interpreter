package main

import (
	"fmt"
	"os"
	. "test1/tokenizer"
)

// global var
var symtab = make(map[string]interface{}) // symbol table to hold var name and their values
var tokenlist []Token 		              // store Token need to test if work first
var tokenIndex int = -1 				  // index of the current token in tokenlist
var token Token                           // current token

// keywords and their category
var keyWords = map[string]int {
    "print" : PRINT, "None" : NONE, "True" : TRUE,
    "False" : FALSE, "pass" : PASS, "if" : IF,
    "else" : ELSE, "while" : WHILE,
    //start of t3 
    "def" : DEF, "return" : RETURN, "global" : GLOBAL, 
    "input" : INPUT, "int" : INT,
}

// one-character tokens and their category
var smallTokens = map[string]int {
    "=" : ASSIGNOP, "(" : LEFTPARENT, ")" : RIGHTPARENT, "+" : PLUS,
    "-" : MINUS, "*" : TIMES, "\n" : NEWLINE, "" : EOF, "==" : EQUAL,
    "<" : LESSTHAN, "<=" : LESSEQUAL, ">" : GREATERTHAN, ">=" : GREATEREQUAL,
    "!" : ERROR, "!=" : NOTEQUAL, "," : COMMA, ":" : COLON, "/" : DIV,
}

/*#################
### PARSER CODE ###
#################*/

// check if the current token in tokenlist has the same category as expectedCategory
// if not then the grammar is invalid
func consume(expectedCategory int) {
	if tokenlist[tokenIndex].Category != expectedCategory {
		fmt.Println("Not match, expecting",expectedCategory, "on line", tokenlist[tokenIndex].Row)
		os.Exit(1)
	} else {
		// no error advance to next token
		advance()
	}
}

// advance to the next token and save current token
func advance() {
	tokenIndex += 1
	if tokenIndex >= len(tokenlist) {
		fmt.Println("Unexpected end of file")
		os.Exit(1)
	}
	token = tokenlist[tokenIndex]
}

// parse all valid stmt, 
// <program> -> <stmt>* EOF
func program() {
	for token.Category == PRINT || token.Category == NAME {
		stmt()
	}
	// can not call consume(EOF) because advance() will yield error
	if token.Category != EOF {
		fmt.Println("Expecting EOF")
		os.Exit(1)
	}
}

// <stmt> -> <simplestmt> NEWLINE
// <simplestmt> -> <assignmentstmt> | <printstmt>
func stmt() {
	simplestmt()
	consume(NEWLINE)
}

func simplestmt() {
	if token.Category == NAME {
		assignmentstmt()
	} else if token.Category == PRINT {
		printstmt()
	} else {
		fmt.Println("Expecting statment, got", token.Lexeme)
	}
}

// <assignmentstmt> -> NAME "=" <expr>
func assignmentstmt() {
	advance() // simplestmt() already check this token is NAME
	consume(ASSIGNOP) 
	expr()
}

// <printstmt> -> "print" "(" <expr> ")"
func printstmt() {
	advance()
	consume(LEFTPARENT)
	expr()
	consume(RIGHTPARENT)
}

// <expr> -> <term> ("+" <term>)*
func expr() {
	term()
	for token.Category == PLUS {
		advance()
		term()
	}
}

// <term> -> <factor> ("*" <factor>)*
func term() {
	factor()
	for token.Category == TIMES {
		advance()
		factor()
	}
}

/*
  <factor> -> "+" <factor>
  <factor> -> "-" <factor>
  <factor> -> UNSIGNEDINT
  <factor> -> NAME
  <factor> -> "(" <expr> ")"
*/
func factor() {
	if token.Category == PLUS {
		advance()
		factor()
	} else if token.Category == MINUS {
		advance()
		factor()
	} else if token.Category == UNSIGNEDINT {
		advance()
	} else if token.Category == NAME {
		advance()
	} else if token.Category == LEFTPARENT {
		advance()
		expr()
		consume(RIGHTPARENT)
	} else {
		fmt.Println("Expecting factor")
		os.Exit(1)
	}
}


func main() {
    // tokenize input file and store in tokenlist
	BuildTokenList()
	tokenlist = TokenList
	//PRINT token list
	// for _, token := range tokenlist {
 //        fmt.Printf("%d %d %s\n", token.Row, token.Column, token.Lexeme)
 //    }
 //    fmt.Println("length of tokenlist" , len(tokenlist))
	

	// advance to initialize token to tokenlist[0] 
	advance()

	// start the parser
	program() // should print nothing
}



















