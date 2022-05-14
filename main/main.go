package main

import (
	"fmt"
	"os"
	. "test1/tokenizer"
	"strconv"
	r "reflect"
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

// conver boolean value to 1 or 0
func btoi(b bool) int {
	if b {return 1}
	return 0
}


// take 2 interface{} and return comparision result
func compare(x, y interface{}, compOp string) bool {
	s1, xIsStr := x.(string)
	s2, yIsStr := y.(string)

	//fmt.Println(s1, s2)
	if s1 == "None" || s2 == "None" {
		if s1 == s2 || compOp == "!=" {
			return true
		}
		return false
	}

	if xIsStr && yIsStr {
		// switch
		switch compOp {
		case "==":
			return s1 == s2
		case "!=":
			return s1 != s2
		case "<":
			return s1 < s2
		case "<=":
			return s1 <= s2
		case ">":
			return s1 > s2
		case ">=":
			return s1 >= s2
		}
		
	} else if xIsStr || yIsStr {   // 1 string 1 number
		if compOp == "==" || compOp == "!=" {
			return false
		} else {
			panic("Compare string with number.")
		}
	}

	var xv float64
	var yv float64
	if r.ValueOf(x).CanFloat() {
		xv = r.ValueOf(x).Float()
	} else { 								// unsafe
		xv = float64(x.(int))
	}

	if r.ValueOf(y).CanFloat() {
		yv = r.ValueOf(y).Float()
	} else { 								// unsafe
		yv = float64(y.(int))
	}

	// switch
	switch compOp {
		case "==":
			return xv == yv
		case "!=":
			return xv != yv
		case "<":
			return xv < yv
		case "<=":
			return xv <= yv
		case ">":
			return xv > yv
		case ">=":
			return xv >= yv
		}

	return false // death code, never execute
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
	m := map[int]bool{PRINT:true,NAME:true,IF:true,WHILE:true,PASS:true,DEF:true,}
	for _,ok := m[token.Category]; ok; ok = m[token.Category] {
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
	switch token.Category {
	case NAME, PRINT, PASS, RETURN, GLOBAL:
		simplestmt()
		consume(NEWLINE)
	case WHILE, IF:
		compoundstmt()
	}
}

func simplestmt() {
	switch token.Category {
	case NAME: // can be assignstmt a = e or functioncall f()  
		if tokenlist[tokenIndex+1].Category == ASSIGNOP {
			assignmentstmt()
		} else if tokenlist[tokenIndex+1].Category == LEFTPARENT {
			functioncall()
		}
	case PRINT:
		printstmt()
	case PASS:
		passstmt()
	case RETURN:
		returnstmt()
	case GLOBAL:
		globalstmt()
	// default:
	// 	fmt.Println("Expecting statement, got", token.Lexeme)
	// 	os.Exit(1)
	}
}

func compoundstmt() {
	if token.Category == IF {
		ifstmt()
	} else if token.Category == WHILE {
		whilestmt()
	} else if token.Category == DEF {
		defstmt()
	} else {
		fmt.Println("Expecting IF or WHILE got", token.Lexeme, "at line", token.Row)
		os.Exit(1)
	}
}

// <assignmentstmt> -> NAME "=" <relexpr>
func assignmentstmt() {
	left := token.Lexeme
	advance() // simplestmt() already check this token is NAME
	consume(ASSIGNOP) 
	relexpr()
	symtab[left] = stackPop()
}

// <printstmt> -> "print" "(" [<relexpr> (COMMA <relexpr>)* [COMMA]] ")"
func printstmt() {
	advance()
	consume(LEFTPARENT)
	if token.Category != RIGHTPARENT {
		relexpr()
		v := stackPop()
		f,isFloat := v.(float64)
		if isFloat {
			fmt.Printf("%.1f", f)
		}else {
			fmt.Print(v)
		}
		for token.Category == COMMA {
			// there are 2 cases:  ,e OR ,)
			advance()
			if token.Category == RIGHTPARENT {
				break;
			}
			relexpr()
			v = stackPop()
			f,isFloat = v.(float64)
			if isFloat {
				fmt.Printf("%0f", f)
			}else {
				fmt.Print(" ",v)
			}
		}
	}
	fmt.Println()
	consume(RIGHTPARENT)
}

func passstmt() {
	consume(PASS)
}

// functioncall -> NAME "(" [relexpr ("," relexpr)*] ")"
func functioncall() {
	advance() // pass NAME
	consume(LEFTPARENT)
	if token.Category != RIGHTPARENT {
		relexpr()
		for token.Category == COMMA {
			consume(COMMA)
			relexpr()
		}
	}
	consume(RIGHTPARENT)
}

// returnstmt -> RETURN [<relexpr>]
func returnstmt() {
	advance()
	if token.Category != NEWLINE {
		relexpr()
	}
}

// globalstmt -> GLOBAL NAME ("," NAME)
func globalstmt() {
	advance()
	consume(NAME)
	for token.Category == COMMA {
		advance()
		consume(NAME)
	}
}

//<ifstmt> -> "if" <relexpr> ":" <codeblock> ["else" ":" <codeblock>]
func ifstmt() {
	advance()
	relexpr()
	consume(COLON)
	saveVal := stackPop()
	v,ok := saveVal.(int); 
	//fmt.Println(v, " ", ok)
	if ok && v == 1 {
		//fmt.Println("if is true")
		codeblock()
	} else {
		consume(NEWLINE)
		indentCol := token.Column
		consume(INDENT)
		for {
			if token.Category == DEDENT && token.Column < indentCol {
				advance()
				break
			}
			advance()
		}
	}

	if token.Category == ELSE {
		advance()
		consume(COLON)
		if v, ok := saveVal.(int) ; !ok || v != 1 {
			codeblock()
		} else {
			consume(NEWLINE)
			indentCol := token.Column
			consume(INDENT)
			for {
				if token.Category == DEDENT && token.Column < indentCol {
					advance()
					break
				}
				advance()
			}
		}
	}
}

// <whilestmt> -> "while" "(" <relexpr> ")" ":" <codeblock>
func whilestmt() {
	advance()
	saveIndex := tokenIndex
	for {
		relexpr()
		consume(COLON)
		v,ok := stackPop().(int)
		//fmt.Println(v, " ", ok) // 1, true and 1 true
		if ok && v == 1 {
			//fmt.Println("While hit")
			codeblock()
			tokenIndex = saveIndex
			token = tokenlist[tokenIndex]
		} else {
			break
		}
	}
	consume(NEWLINE)
	indentCol := token.Column
	consume(INDENT)
	for {
		if token.Category == DEDENT && token.Column < indentCol {
			advance()
			break
		}
		advance()
	}
}

// defstmt -> DEF NAME "(" [NAME ("," NAME)*] ")" ":" <codeblock>
func defstmt() {
	advance()
	consume(NAME)
	consume(LEFTPARENT)
	if token.Category != RIGHTPARENT {
		consume(NAME)
		for token.Category == COMMA {
			advance()
			consume(NAME)
		}
	}
	consume(RIGHTPARENT)
	consume(COLON)
	codeblock()
}

// <codeblock> -> NEWLINE INDENT stmt+ DEDENT
func codeblock() {
	consume(NEWLINE)
	consume(INDENT)
	//stmt() // must have at least 1 stmt
	m := map[int]bool{PRINT:true,NAME:true,IF:true,WHILE:true,PASS:true, GLOBAL:true, RETURN:true,}
    for _,ok := m[token.Category]; ok; ok = m[token.Category] {
    	stmt()
    }
	consume(DEDENT)
}

// <expr> -> <term> (("+" | "-") <term>)*
func expr() {
	term()
	for token.Category == PLUS || token.Category == MINUS {
		saveCat := token.Category
		advance()
		term()
		rightop := stackPop()
		leftop := stackPop()

		ri, risInt := rightop.(int)
		li, lisInt := leftop.(int)
		rs, risStr := rightop.(string)
		ls, lisStr := leftop.(string)
		rf, risFloat := rightop.(float64)
		lf, lisFloat := leftop.(float64)
		if saveCat == PLUS {
			if risInt && lisInt {
				stack = append(stack, ri + li)
			} else if risFloat && lisFloat {
				stack = append(stack, rf + lf)
			} else if risStr && lisStr {
				stack = append(stack, ls + rs)  // string concat
			} else if lisInt && risFloat {
				stack = append(stack, float64(li) + rf)
			} else if lisFloat && risInt {
				stack = append(stack, lf + float64(ri))
			} else if risStr || lisStr {
				panic("addition between string and number")
			}
		} else if saveCat == MINUS {
			if risInt && lisInt {
				stack = append(stack, li - ri)
			} else if risFloat && lisFloat {
				stack = append(stack, lf - rf)
			} else if risStr && lisStr {
				panic("minus two strings")
			} else if lisInt && risFloat {
				stack = append(stack, float64(li) - rf)
			} else if lisFloat && risInt {
				stack = append(stack, lf - float64(ri))
			} else if lisStr || risStr {
				panic("minus between string and number")
			}
		}
	}
}

// <relexpr> -> <expr> [CONDITIONALOP <expr>]
func relexpr() {
	m := map[int]bool{EQUAL:true,NOTEQUAL:true,LESSTHAN:true,LESSEQUAL:true,
					  GREATERTHAN:true,GREATEREQUAL:true,}
	expr()
	if _,ok := m[token.Category]; ok {
		saveCat := token.Category
		advance()
		expr()
		rightop := stackPop()
		leftop := stackPop()
		switch saveCat {
		case EQUAL:
			stack = append(stack, btoi(compare(leftop, rightop, "==")))
		case NOTEQUAL:
			stack = append(stack, btoi(compare(leftop, rightop, "!=")))
		case LESSTHAN:
			stack = append(stack, btoi(compare(leftop, rightop, "<")))
		case LESSEQUAL:
			stack = append(stack, btoi(compare(leftop, rightop, "<=")))
		case GREATERTHAN:
			stack = append(stack, btoi(compare(leftop, rightop, ">")))
		case GREATEREQUAL:
			stack = append(stack, btoi(compare(leftop, rightop, ">=")))
		}
	}
}

// <term> -> <factor> (("*" | "/") <factor>)*
func term() {
	sign = 1
	factor()
	//fmt.Println("token lexeme in term ", token.Lexeme)
	for token.Category == TIMES || token.Category == DIV {
		saveCat := token.Category
		advance()
		sign = 1
		factor()
		rightop := stackPop()
		leftop := stackPop()

		ri, risInt := rightop.(int)
		li, lisInt := leftop.(int)
		rs, risStr := rightop.(string)
		ls, lisStr := leftop.(string)
		rf, risFloat := rightop.(float64)
		lf, lisFloat := leftop.(float64)
		if saveCat == TIMES {
			if risInt && lisInt {
				stack = append(stack, ri * li)
			} else if risFloat && lisFloat {
				stack = append(stack, rf * lf)
			} else if risStr && lisStr {
				panic("multiply string sequense")
			} else if lisInt && risFloat {
				stack = append(stack, float64(li) * rf)
			} else if lisFloat && risInt {
				stack = append(stack, lf * float64(ri))
			} else if risStr || lisStr {
				// string multiplication, "a" * 3 = "aaa"
				if risStr {
					buf := ""
					for i := 0; i < li; i++ {
						buf += rs
					}
					stack = append(stack, buf)
				} else if lisStr {
					buf := ""
					for i := 0; i < ri ; i++ {
						buf += ls
					}
					stack = append(stack, buf)
				}
			}
		} else if saveCat == DIV {
			if risInt && lisInt {
				stack = append(stack, li /ri)
			} else if risFloat && lisFloat {
				stack = append(stack, lf / rf)
			} else if risStr && lisStr {
				panic("divide string sequense")
			} else if lisInt && risFloat {
				stack = append(stack, float64(li) / rf)
			} else if lisFloat && risInt {
				stack = append(stack, lf / float64(ri))
			} else if risStr || lisStr {
				panic("divide string and number")
			}
		}
		
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
  <factor> -> INPUT "(" STRING ")"
  <factor> -> INT "(" <relexpr> ")"
  <factor> -> <functioncal>          // how to know when a factor is a NAME or function call
  									 // 1 + a or 1 + a()
*/
func factor() {
	if token.Category == PLUS {					// change to switch
		advance()
		factor()
	} else if token.Category == MINUS {
		sign = -sign
		advance()
		factor()
	} else if token.Category == TIMES {
		advance()
		factor()
	} else if token.Category == DIV {
		advance()
		factor()
	} else if token.Category == UNSIGNEDINT {
		i, _ := strconv.Atoi(token.Lexeme)       // need handle error here
		stack = append(stack, sign * i)
		advance()
	} else if token.Category == UNSIGNEDFLOAT {
		f, _ := strconv.ParseFloat(token.Lexeme, 64)
		stack = append(stack, float64(sign) * f)
		advance()
	} else if token.Category == NAME {
		// check if this var is declared (in symtab)
		// if it declared push value from symtab to stack
		
		// 2 cases NAME or function call
		if tokenlist[tokenIndex+1].Category == LEFTPARENT {
			functioncall()
		} else if tokenlist[tokenIndex+1].Category != LEFTPARENT {

			if v, ok := symtab[token.Lexeme]; ok {
				i, isInt := v.(int)
				f, isFloat := v.(float64)
				if isInt {
					stack = append(stack, sign * i)
				} else if isFloat {
					stack = append(stack, float64(sign) * f)
				}
			} else {
				fmt.Printf("Name %s not declared\n", token.Lexeme)
				os.Exit(1)
			}
			advance()
		}
	} else if token.Category == LEFTPARENT {
		saveSign := sign
		advance()
		relexpr()	
		if saveSign == -1 {
			i, isInt := stack[len(stack)-1].(int)
			f, isFloat := stack[len(stack)-1].(float64)
			if isInt {
				stack[len(stack)-1] = -i
			} else if isFloat {
				stack[len(stack)-1] = -f
			}
		}
		consume(RIGHTPARENT)
	} else if token.Category == STRING {
		stack = append(stack, token.Lexeme) // stack now contain string
		advance()
	} else if token.Category == TRUE {
		stack = append(stack, 1)			// 1 represent true
		advance()
	} else if token.Category == FALSE {
		stack = append(stack, 0)			// 0 represent false
		advance()
	} else if token.Category == NONE {
		// None only == None and everything else is false
		stack = append(stack, "None")
		advance()
	} else if token.Category == INPUT {
		advance()
		consume(LEFTPARENT)
		consume(STRING)
		consume(RIGHTPARENT)
	} else if token.Category == INT {
		advance()
		consume(LEFTPARENT)
		relexpr()
		consume(RIGHTPARENT)
	} else {
		fmt.Println("Expecting factor")
		os.Exit(1)
	}
	//fmt.Println(stack)
}

/*
	i2 -> i3

	now that p3 is done, next work on i2, check if i1 still work by uncomment
*/

/*
	One feature at a time
	t4 grammar -> need to build t4 tokenizer
	for, range, len, float convert, support "" for string

	t5 grammar -> support class, list [], and dict {} 
*/


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



















