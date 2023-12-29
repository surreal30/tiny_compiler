package main

import (
	"fmt"
	"log"
	"strings"
)

type token struct {
	kind string
	value string
}

// Input string of code
func tokenizer(input string) []token {
	// A new line is appended 
	input += "\n"

	// A 'current' variable for tracing our position in the code like a cursor
	current := 0

	// Add a slice of our 'token' type for appending tokens to
	tokens := []token{}

	// Create a for loop to increment 'current' variable as much as we want
	// We can increment as much as we want because token can be of variable length
	for current < len([]rune(input)) {
		// Store 'current' character in 'input'
		char := string([]rune(input)[current])

		// Check for open parenthesis
		if char == "(" {
			// If it is parenthesis then append a new token to our slice with kind 'paren'
			// and set value to open parenthesis
			tokens = append(tokens, token{
				kind: "paren",
				value: "(",
			})

			// Increment 'current'
			current++

			// Continue to the next cycle of loop
			continue
		}

		// Check for closing parenthesis

		if char == ")" {
			// Append new token to tokens
			tokens = append(tokens, token{
				kind: "paren",
				value: ")",
			})

			// Increment current
			current++

			// Continue loop
			continue
		}

		// Skip whitespace
		if char == " " {
			current++
			continue
		}

		// numbers can be of any length. so we check if the first character is number
		if isNumber(char) {
			// Create a string value to append the characters
			value := ""

			// loop to append char to value if it is number
			// increment 'current' after each appending
			for isNumber(char) {
				value += char
				current++
				char = string([]rune(input)[current])
			}

			// Append the number as new token in tokens
			tokens = append(tokens, token{
				kind:"number",
				value: value,
			})

			// Continue the loop
			continue
		}

		// Last type of token is 'name'
		// This is sequence of characters instead of number
		if isLetter(char) {
			value := ""

			// Loop to append char to value
			for isLetter(char) {
				value += char
				current++
				char = string([]rune(input)[current])     
			}

			// Appen the name as new token to token
			tokens = append(tokens, token{
				kind: "name",
				value: value,
			})

			continue
		}

		break
	}

	// Return the tokens array
	return tokens
}

// Check whether character is number or not
// Number range is 0-9
func isNumber(char string) bool {
	if char == "" {
		return false
	}
	n := []rune(char)[0]
	if n >= '0' && n <= '9' {
		return true
	}
	return false
}

// Check whether charact is letter or not
// Letter range is a-z
func isLetter(char string) bool {
	if char == "" {
		return false
	}
	n := []rune(char)[0]
	if n >= 'a' && n <= 'z' {
		return true
	}
	return false
}

// Define type node and pointer types
type node struct {
	kind string
	value string
	name string
	callee *node
	expression *node
	body []node
	params []node
	arguments *[]node
	context *[]node
}

type ast node

// counter variable for parsing 
var pc int

// variable to store slice of token inside it
var pt []token 

// parser function to slice of 'token'
func parser(tokens []token) ast {
	// assign parser counter and parser token a value
	pc = 0
	pt = tokens

	// Create root node of AST with 'program' node
	ast := ast{
		kind: "Program",
		body: []node{},
	}

	// Push nodes to ast.body
	for pc < len(pt) {
		ast.body = append(ast.body, walk())
	}

	// return ast
	return ast
}

func walk() node{
	// grab the current token
	token := pt[pc]

	// Each token is split into different code path
	// First, check if the token is number
	if token.kind == "number" {
		// increment current
		pc++

		// return a new AST node called 'numberLiteral' and set it's value to token value
		return node{
			kind: "NumberLiteral",
			value: token.value,
		}
	}

	// Check for open parenthesis 
	if token.kind == "paren" && token.value == "(" {
		// increment current to skip parenthesis as it is not neccessary in ast
		pc++
		token = pt[pc]

		// Create base node "CallExpression"
		n := node{
			kind: "CallExpression",
			name: token.value,
			params: []node{},
		}

		pc++
		token = pt[pc]

		for token.kind != "paren" || (token.kind == "paren" && token.value != ")") {
			// Call the walk() and return node
			n.params = append(n.params, walk())
			token = pt[pc]
		}

		// Increment 'current'
		pc++

		// Return node
		return n
	}

	// If we haven't recognised the token type then throw error
	log.Fatal(token.kind)
	return node{}
}

// define visitor
type visitor map[string]func(n *node, p node)

// define traverser
func traverser(a ast, v visitor) {
	// Call 'traversorNode' 
	traverseNode(node(a), node{}, v)
}

// Traversor array to itierate over a slice and call traverseNode
func traverseArray(a []node, p node, v visitor) {
	for _, child := range a {
		traverseNode(child, p, v)
	}
}

func traverseNode (n, p node, v visitor) {
	for k, va := range v {
		if k == n.kind {
			va(&n, p)
		}
	}

	switch n.kind {
	
	case "Program":
		traverseArray(n.body, n, v)
		break

	case "CallExpression":
		traverseArray(n.params, n, v)
		break

	case "NumberLiteral":
		break

	// if node type is not recognised then throw error
	default:
		log.Fatal(n.kind)
	}
}

// Transformer function to take in lisp ast
func transformer(a ast) ast {
	// Create AST with 'program' root
	nast := ast{
		kind: "Program",
		body: []node{},
	}

	a.context = &nast.body

	// Call traverser with our ast and a visitor
	traverser(a, map[string]func(n *node, p node){
		// First visitor accepts NumberLiterals
		"NumberLiteral": func(n *node, p node) {
			// Create a new node named NumberLiteral that we will push in the parent context
			*p.context = append(*p.context, node{
				kind: "NumberLiteral",
				value: n.value,
			})
		},

		// Create CallExpression
		"CallExpression": func(n *node, p node) {
			// Create CallExpression node with nested identifier
			e := node{
				kind: "CallExpression",
				callee: &node{
					kind: "Identifier",
					name: n.name,
				},
				arguments: new([]node),
			}

			// define new context
			n.context = e.arguments

			// Check if the parent node is CallExpression or not. If not
			if p.kind != "CallExpression" {
				// Wrap CallExpression with ExpressionStatement
				es := node{
					kind: "ExpressionStatement",
					expression: &e,
				}

				// Push 'CallExpression' to parent's context
				*p.context = append(*p.context, es)
			} else {
				*p.context = append(*p.context, e)
			}
		},
	})

	// Return our new ast
	return nast
}

func codeGenerator(n node) string {
	// breakdown things by the type of the node
	switch n.kind {
		// If we have 'program' node, we will map through each node in the 'body'
		// and run through the code generator and join them with a newline
	case "Program":
		var r []string
		for _, no := range n.body {
			r = append(r, codeGenerator(no))
		}

		return strings.Join(r, "\n")

	// For ExpressionStatement, we will call code generator on nested expression and add semicolon in the end
	case "ExpressionStatement":
		return codeGenerator(*n.expression) + ";"

	// For CallExpression, we will print the 'callee', add an open parenthesis,
	// we wil map through each node in 'argument' array and run them through code generator
	// joining them with comma and then adding closing parenthesis
	case "CallExpression":
		var ra []string
		c := codeGenerator(*n.callee)

		for _, no := range *n.arguments {
			ra = append(ra, codeGenerator(no))
		}

		r := strings.Join(ra, ", ")
		return c + "(" + r + ")"

	// For Identifier just return the node name
	case "Identifier":
		return n.name

	// For NumberLiteral return the value 
	case "NumberLiteral":
		return n.value

	// If we don't recognise the node then throw error
	default:
		log.Fatal("error")
		return ""
	}
}

// Compiler
func compiler(input string)string {
	tokens := tokenizer(input)
	ast := parser(tokens)
	nast := transformer(ast)
	out := codeGenerator(node(nast))

	return out
}

func main() {
	program := "(add 10 (subtract 10 6 ))"
	out := compiler(program)
	fmt.Println(out)
}