package main

import (
	"fmt"
	//"os"
	"test1/tokenizer"
)

func main() {
	tokenizer.BuildTokenList()
	tokenlist := tokenizer.TokenList
	
	for _, token := range tokenlist {
        fmt.Printf("%d %d %s\n", token.Row, token.Column, token.Lexeme)
    }
    
    fmt.Println("length of tokenlist" , len(tokenlist))
}
