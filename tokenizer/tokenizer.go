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
    EOF = 0 // end of file
    PRINT = 1 // 'print' keyword
    UNSIGNEDINT = 2 // integer
    NAME = 3 // identifier that is not a keyword
    ASSIGNOP = 4 // '='
    LEFTPARENT = 5 // '('
    RIGHTPARENT = 6 // ')' 
    PLUS = 7 // '+'
    MINUS = 8 // '-'
    TIMES = 9 // '*'
    NEWLINE = 10 // newline character
    ERROR = 11 // if not any of the above error
)

var catnames = map[int]string {
    0 : "EOF",
    1 : "PRINT",
    2 : "UNSIGNEDINT",
    3 : "NAME",
    4 : "ASSIGNOP",
    5 : "LEFTPARENT",
    6 : "RIGHTPARENT",
    7 : "PLUS",
    8 : "MINUS",
    9 : "TIMES",
    10 : "NEWLINE",
    11 : "ERROR",
}

// keywords and their category
var keyWords = map[string]int {
    "print" : PRINT,
}

// one-character tokens and their category
var smallTokens = map[string]int {
    "=" : ASSIGNOP, "(" : LEFTPARENT, ")" : RIGHTPARENT, "+" : PLUS,
    "-" : MINUS, "*" : TIMES, "\n" : NEWLINE, "" : EOF,
}

// package level var,
var tokenlist []token // store tokens
var source string     // string of source file
var sourceIndex int   // keep track of current character's index in source string
var prevChar byte     // save previous character
var line int          // line number of current token
var column int        // column number of current token
var isBlankLine bool    // check if current line is a bankline 

// return the current character pointed by sourceIndex and adjust row and column.
func getChar() byte {
    if prevChar == '\n' {
        column = 0
        line += 1
        isBlankLine = true
    }

    // if at the end of source file, change isEOF flag
    if sourceIndex >= len(source) {
        column = 1
        line += 1
        return 0
    }

    c := source[sourceIndex]
    sourceIndex += 1
    column += 1
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
    for {
        // skip white space but not newline
        for currChar != '\n' && unicode.IsSpace(rune(currChar)) {
            currChar = getChar()
        }

        var currToken = token{row: line, column: column, category: -1, lexeme: ""} 
        if unicode.IsDigit(rune(currChar)) {
            currToken.category = UNSIGNEDINT
            for {
                currToken.lexeme += string(currChar)
                currChar = getChar()
                if !unicode.IsDigit(rune(currChar)) {
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
            
        } else if v, ok := smallTokens[string(currChar)]; ok {
            currToken.category = v
            currToken.lexeme = string(currChar)
            currChar = getChar()
        } else if currChar == 0 {
            currToken.category = EOF
            currToken.lexeme = ""
        } else {
            currToken.category = ERROR
            currToken.lexeme = string(currChar)
            fmt.Println("error occur with token " + string(currChar) + " on line", line)
            os.Exit(1)
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
    fmt.Println(source)
    tokenizer()
    for _, token := range tokenlist {
        fmt.Println(token.lexeme, catnames[token.category])
    }
    fmt.Println("length of tokenlist" , len(tokenlist))
}







