package main

import (
	"fmt"
	"os"
	. "test1/tokenizer"
	//"strconv"
)

// global var
var symtab = make(map[string]interface{}) // symbol table to hold var name and their values
var stack []interface{}					  // data stack use save expr value
var tokenlist []Token 		              // store Token need to test if work first
var tokenIndex int = -1 				  // index of the current token in tokenlist
var token Token                           // current token
var sign int 							  // use to track unary minus sign

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
//helper func to pop data stack
func stackPop() interface{} {
	save := stack[len(stack)-1]
	stack = stack[:len(stack)-1]	// slice to remove last element
	return save
}


// check if the current token in tokenlist has the same category as expectedCategory
// if not then the grammar is invalid
func consume(expectedCategory int) {
	if tokenlist[tokenIndex].Category != expectedCategory {
		fmt.Println("Not match, expecting",expectedCategory, "on line", tokenlist[tokenIndex].Row)
		fmt.Println("current token :", token.Lexeme)
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
	for token.Category == PRINT || token.Category == NAME || token.Category == IF || token.Category == WHILE || token.Category == PASS {
		stmt()
	}
	// can not call consume(EOF) because advance() will yield out of range error
	if token.Category != EOF {
		fmt.Println("Expecting EOF, got", token.Category)
		os.Exit(1)
	}
}

// <stmt> -> <simplestmt> NEWLINE | compoundstmt  
//*compoundstmt does not have NEWLINE because <codeblock> already consume it

// <simplestmt> -> <assignmentstmt> | <printstmt>
// <compoundstmt> -> <ifstmt> | <whilestmt>
func stmt() {
	if token.Category == NAME || token.Category == PRINT || token.Category == PASS {
		simplestmt()
		consume(NEWLINE)
	} else if token.Category == WHILE || token.Category == IF {
		compoundstmt()
	}
}

func simplestmt() {
	if token.Category == NAME {
		assignmentstmt()
	} else if token.Category == PRINT {
		printstmt()
	}else if token.Category == PASS {
		passstmt()
	} else {
		fmt.Println("Expecting NAME or PRINT or PASS, got", token.Lexeme)
	}
}

func compoundstmt() {
	if token.Category == IF {
		ifstmt()
	} else if token.Category == WHILE {
		whilestmt()
	} else {
		fmt.Println("Expecting IF or WHILE got", token.Lexeme, "at line", token.Row)
		os.Exit(1)
	}
}

// <assignmentstmt> -> NAME "=" <relexpr>
func assignmentstmt() {
	//left := token.Lexeme
	advance() // simplestmt() already check this token is NAME
	consume(ASSIGNOP) 
	relexpr()
	//symtab[left] = stackPop()
}

// <printstmt> -> "print" "(" [<relexpr> (COMMA <relexpr>)* [COMMA]] ")"
func printstmt() {
	advance()
	//fmt.Println("current token lexeme", token.Lexeme)
	consume(LEFTPARENT)
	//fmt.Println("current token lexeme after (", token.Lexeme)
	if token.Category != RIGHTPARENT {
		relexpr()
		for token.Category == COMMA {
			// there are 2 cases:  ,e OR ,)
			if tokenlist[tokenIndex + 1].Category == RIGHTPARENT {
				advance()	
			} else {
				advance()
				relexpr()
			}
		}
	}
	//fmt.Println(stackPop())
	consume(RIGHTPARENT)
}

func passstmt() {
	consume(PASS)
}

//<ifstmt> -> "if" <relexpr> ":" <codeblock> ["else" ":" <codeblock>]
func ifstmt() {
	advance()
	relexpr()
	consume(COLON)
	codeblock()
	if token.Category == ELSE {
		advance()
		consume(COLON)
		codeblock()
	}
}

// <whilestmt> -> "while" "(" <relexpr> ")" ":" <codeblock>
func whilestmt() {
	advance()
	relexpr()
	consume(COLON)
	codeblock()
}

// <codeblock> -> NEWLINE INDENT stmt* DEDENT
func codeblock() {
	consume(NEWLINE)
	consume(INDENT)
    for token.Category == PRINT || token.Category == NAME || token.Category == IF || token.Category == WHILE || token.Category == PASS {
    	stmt()
    }
	consume(DEDENT)
}

// <expr> -> <term> (("+" | "-") <term>)*
func expr() {
	term()
	for token.Category == PLUS || token.Category == MINUS {
		advance()
		term()
		/*
		rightop := stackPop()
		leftop := stackPop()
		if token.Category == PLUS{
			stack = append(stack, rightop.(int) + leftop.(int))
		} else {
			stack = append(stack, rightop.(int) - leftop.(int))
		}
		*/
	}
}

// <relexpr> -> <expr> [CONDITIONALOP <expr>]
func relexpr() {
	expr()

	switch token.Category {
	case EQUAL, NOTEQUAL, LESSTHAN, LESSEQUAL, GREATERTHAN, GREATEREQUAL:
		advance()
		expr()
	//default:
		// fmt.Println("Expecting relational operator, got", token.Lexeme)
		// os.Exit(1)
	}
}

// <term> -> <factor> (("*" | "/") <factor>)*
func term() {
	sign = 1
	factor()
	for token.Category == TIMES || token.Category == DIV {
		advance()
		sign = 1
		factor()
		/*
		rightop := stackPop()
		leftop := stackPop()
		if token.Category == TIMES {
			stack = append(stack, rightop.(int) * leftop.(int))
		} else {
			stack = append(stack, rightop.(int) / leftop.(int))   // integer div need float div
		}
		*/
	}
}

/*
  <factor> -> "+" <factor>
  <factor> -> "-" <factor>
  <factor> -> UNSIGNEDINT
  <factor> -> UNSIGNEDFLOAT
  <factor> -> NAME
  <factor> -> "(" <expr> ")"
  <factor> -> STRING
  <factor> -> TRUE
  <factor> -> FALSE
  <factor> -> NONE
*/
func factor() {
	if token.Category == PLUS {					// change to switch
		advance()
		factor()
	} else if token.Category == MINUS {
		sign = -sign
		advance()
		factor()
	} else if token.Category == UNSIGNEDINT {
		//i, _ := strconv.Atoi(token.Lexeme)       // need handle error here
		//stack = append(stack, sign * i)
		advance()
	} else if token.Category == UNSIGNEDFLOAT {
		advance()
	} else if token.Category == NAME {
		// check if this var is declared (in symtab)
		// if it declared push value from symtab to stack
		/*
		if v, ok := symtab[token.Lexeme]; ok {
			stack = append(stack, sign * v.(int))  // will panic if not int
		} else {
			fmt.Printf("Name %s not declared\n", token.Lexeme)
			os.Exit(1)
		}
		*/
		advance()
	} else if token.Category == LEFTPARENT {
		//saveSign := sign
		advance()
		relexpr()
		/*
		if saveSign == -1 {
			stack[len(stack)-1] = -stack[len(stack)-1].(int)
		}
		*/
		consume(RIGHTPARENT)
	} else if token.Category == STRING {
		advance()
	} else if token.Category == TRUE {
		advance()
	} else if token.Category == FALSE {
		advance()
	} else if token.Category == NONE {
		advance()
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
 //         fmt.Printf("%d %d %s\n", token.Row, token.Column, token.Lexeme)
 //    }
 //    fmt.Println("length of tokenlist" , len(tokenlist))
	

	// advance to initialize token to tokenlist[0] 
	advance()

	// start the interpreter
	program()
}



















