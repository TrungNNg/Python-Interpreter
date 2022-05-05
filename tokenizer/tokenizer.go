// Tokenizer for syntax of the Python programming language
package main

import (
    "fmt"
    "os"
    "unicode"
)

// each token keep track of 4 pieces of information: row, col, category, and lexeme
type token struct {
    row int
    column int
    category int   // see token categories below
    lexeme string  // the literal string of the token
}

// constants to represent token categories
const (
    EOF = iota // end of file
    PRINT  // 'print' keyword
    UNSIGNEDINT // integer
    NAME // identifier that is not a keyword
    ASSIGNOP // '='
    LEFTPARENT // '('
    RIGHTPARENT // ')' 
    PLUS // '+'
    MINUS // '-'
    TIMES // '*'
    NEWLINE // newline character
    ERROR // if not any of the above error
    NONE
    TRUE
    FALSE
    PASS
    IF
    ELSE
    WHILE
    UNSIGNEDFLOAT
    STRING
    EQUAL
    NOTEQUAL
    LESSTHAN
    LESSEQUAL
    GREATERTHAN
    GREATEREQUAL
    DIV
    COMMA
    COLON
    INDENT
    DEDENT
    DEF     // start of t3
    RETURN
    GLOBAL
    INPUT
    INT
)

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

// package level var,
var tokenlist []token // store tokens
var source string     // string of source file
var sourceIndex int   // keep track of current character's index in source string
var prevChar byte = '\n'     // save previous character
var line int          // line number of current token
var column int        // column number of current token
var isBlankLine bool = true    // check if current line is a bankline 
var isInString bool
var parenLevel int    //

// return the current character pointed by sourceIndex and adjust row and column.
func getChar() byte {
    if prevChar == '\n' {
        line += 1
        column = 0
        isBlankLine = true
    }

    // if at the end of source file, change isEOF flag
    if sourceIndex >= len(source) {
        column = 1
        return 0
    }

    c := source[sourceIndex]
    sourceIndex += 1
    column += 1

    // check if # is in string or start of comment
    if c == '#' && !isInString {
        for {
            c = source[sourceIndex]
            sourceIndex += 1
            if c == '\n' {
                break
            }
        }
    }

    if !unicode.IsSpace(rune(c)) {
        isBlankLine = false
    }
    prevChar = c

    // if at the end of blankline, return ' ' instead of '\n'
    if c == '\n' && isBlankLine {
        return ' '
    } else {
        return c
    }
}

func tokenizer() {
    var currChar byte = ' '
    indentStack := []int{1}
    for {
        // skip white space but not newline
        for currChar != '\n' && unicode.IsSpace(rune(currChar)) {
            currChar = getChar()
        }

        var currToken = token{row: line, column: column, category: -1, lexeme: ""} 
        if unicode.IsDigit(rune(currChar)) || currChar == '.'{
            currToken.category = UNSIGNEDINT
            if currChar == '.' {
                currToken.category = UNSIGNEDFLOAT
            }
            for {
                currToken.lexeme += string(currChar)
                currChar = getChar()
                if currToken.category == UNSIGNEDINT && currChar == '.' {
                    currToken.category = UNSIGNEDFLOAT
                } else if !unicode.IsDigit(rune(currChar)) {
                    break
                }
            }
        } else if unicode.IsLetter(rune(currChar)) || currChar == '_' {
            for {
                currToken.lexeme += string(currChar)
                currChar = getChar()
                if !(unicode.IsLetter(rune(currChar)) || unicode.IsDigit(rune(currChar)) || currChar == '_') {
                    break
                }
            }

            // check the token's lexeme to see if it is keyword or user var name 
            if v, ok := keyWords[currToken.lexeme]; ok {
                currToken.category = v
            } else {
                currToken.category = NAME
            }
            
        } else if _, ok := smallTokens[string(currChar)]; ok {
            saveChar := currChar
            currChar = getChar()
            twoChar := string(saveChar) + string(currChar)
            if v, ok := smallTokens[twoChar]; ok {
                currToken.category = v
                currToken.lexeme = string(twoChar)
                currChar = getChar()
            } else {
                currToken.category = smallTokens[string(saveChar)]
                currToken.lexeme = string(saveChar)
            }
        } else if currChar == 39 {
            // current char is single quote ', indicate start of string data
            isInString = true
            for {
                currChar = getChar()
                if currChar == 0 || currChar == '\n'{
                    fmt.Println("unterminated string on", line)
                    os.Exit(1)
                }
                if currChar == 39 { // currChar == '
                    currChar = getChar() // advance pass last single quote
                    currToken.category = STRING
                    isInString = false
                    break
                }
                if currChar == 92 { // currChar == \
                    currChar = getChar()
                    if currChar == 'n' {
                        currToken.lexeme += "\n"
                    } else if currChar == 't' {
                        currToken.lexeme += "\t"
                    } else if currChar == '\n' {
                        currToken.lexeme += "n"
                    } else {
                        currToken.lexeme += string(currChar)
                    }
                } else {
                    currToken.lexeme += string(currChar)
                }
            }

        } else if currChar == 0 {
            currToken.category = EOF
            currToken.lexeme = ""
        } else {
            currToken.category = ERROR
            currToken.lexeme = string(currChar)
            fmt.Println("error occur with token " + string(currChar) + " on line", line)
            os.Exit(1)
        }

        if len(tokenlist) == 0 || tokenlist[len(tokenlist)-1].category == NEWLINE {
            if indentStack[len(indentStack)-1] < currToken.column{
                indentStack = append(indentStack, currToken.column)
                var indentToken = token{row: currToken.row, column: currToken.column, category: INDENT, lexeme : "{"}
                tokenlist = append(tokenlist, indentToken) 
            } else if indentStack[len(indentStack)-1] > currToken.column {
                for {
                    var dedentToken = token{row: currToken.row, column: currToken.column, category : DEDENT, lexeme : "}"}
                    tokenlist = append(tokenlist, dedentToken)
                    indentStack = indentStack[:len(indentStack)-1]
                    if indentStack[len(indentStack)-1] == currToken.column {
                        break
                    } else if indentStack[len(indentStack)-1] < currToken.column {
                        fmt.Println("indetation err on line", line)
                        os.Exit(1)
                    }
                }
            }
        }

        tokenlist = append(tokenlist, currToken)
        if currToken.category == EOF {
            break
        }
    }
}

// a function to read in souce file from command line argument
// and return the source string
func readSourceFile() string {
    // check to see if valid number of cmd line args
    if len(os.Args) == 2 {
        data, err := os.ReadFile(os.Args[1])
        if err != nil {
            fmt.Println("Can not read input file" + os.Args[1])
            os.Exit(1)
        }
        // Python use newline character '\n' to terminate statement, in some editor the
        // last line might not terminate if writer not hit enter key, so below check for
        // newline character and add it to source if it not there
        if data[len(data)-1] != '\n' {
            data = append(data, '\n')
        }
        return string(data)
    } else {
        fmt.Println("invalid number of command line arguments")
        fmt.Println("usage: ./tokenizer <infile>")
        os.Exit(1)
        return "" // never reach, added to pass compiler's complain of no return statement
    } 
}

func main() {
    // read source from source python file
    source = readSourceFile()
    tokenizer()
    for _, token := range tokenlist {
        fmt.Printf("%d %d %s\n", token.row, token.column, string(token.lexeme))
    }
    fmt.Println("length of tokenlist" , len(tokenlist))
}







