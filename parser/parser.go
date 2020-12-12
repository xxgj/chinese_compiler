package parser

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/kevinchen147/chinese_compiler/ast"
)

type Token struct {
	Symbol int
	Value  interface{}
	line   int
	column int
}

func (t *Token) Line() int   { return t.line }
func (t *Token) Column() int { return t.column }

func (t *Token) String() string {
	return fmt.Sprintf("&%T{%d:%d:%s:%v}", t, t.Line(), t.Column(), tokName(t.Symbol), t.Value)
}

type Parser struct {
	Tree        ast.Node
	src         *bufio.Reader
	ungetcheds  *list.List
	SrcFileName string
	line        int
	column      int
	lastToken   *Token
	lastErr     error
	ErrorCount  int
	OnError     func(*Parser, error)
	blocks      *list.List
}

func NewParser(src io.Reader, srcFileName string, initialLineNumber int) *Parser {
	return &Parser{
		src:         bufio.NewReader(src),
		ungetcheds:  list.New(),
		SrcFileName: srcFileName,
		line:        initialLineNumber,
		column:      -1,
		blocks:      list.New(),
	}
}

func (p *Parser) Parse() error {
	p.lastErr = nil
	yyParse(p)
	return p.lastErr
}

func (p *Parser) getch() (rune, error) {
	if p.ungetcheds.Len() > 0 {
		c := p.ungetcheds.Remove(p.ungetcheds.Front()).(rune)
		return c, nil
	}
	c, _, err := p.src.ReadRune()
	return c, err
}

// c == '\n' ?
func (p *Parser) ungetch(c rune) {
	p.column--
	if c == '\n' {
		p.line--
	}
	p.ungetcheds.PushFront(c)
}

func (p *Parser) PushBlock(b *ast.Block) {
	p.blocks.PushFront(b)
}

func (p *Parser) PopBlock() *ast.Block {
	if p.blocks.Len() <= 0 {
		return nil
	}
	return p.blocks.Remove(p.blocks.Front()).(*ast.Block)
}

func (p *Parser) Line() int {
	if p.lastToken == nil {
		return p.line
	}
	return p.lastToken.Line()
}

func (p *Parser) Column() int {
	if p.lastToken == nil {
		return p.column
	}
	return p.lastToken.Column()
}

func (p *Parser) Error(msg string) {
	p.reportError(fmt.Errorf("%s:%d:%d: %s", p.SrcFileName, p.Line(), p.Column(), msg))
}

func (p *Parser) reportError(e error) {
	p.ErrorCount++
	if p.OnError != nil {
		p.OnError(p, e)
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", e)
	p.lastErr = e
}

func (p *Parser) Lex(lval *yySymType) int {
	t := &Token{
		Symbol: 0,
		line:   p.line,
		column: p.column,
	}
	err := p.lexLoop(t)
	if err != nil && err != io.EOF {
		p.reportError(err)
		return 0
	}
	p.lastToken = t
	lval.token = t
	return t.Symbol
}

func (p *Parser) lexLoop(t *Token) error {
	buf := make([]rune, 0)
	state := kSTATE_INITIAL
	var str *stringLiteral
	var c rune
	var err error

	for {
		c, err = p.getch()
		if err != nil {
			break
		}
		p.column++
		if c == '\n' {
			p.line++
			p.column = -1
		}
		switch state {
		case kSTATE_INITIAL:
			t.line = p.line
			t.column = p.column
			switch {
			case unicode.IsSpace(c):
				// do nothing
			case c == '0':
				state = kSTATE_ZERO
			case unicode.IsDigit(c):
				buf = append(buf, c)
				state = kSTATE_INTEGER
			case c == '“':
				str = &stringLiteral{
					begLine:   p.line,
					begCol:    p.column,
					isCharLit: false,
				}
				state = kSTATE_STRING
			case c == '\'':
				str = &stringLiteral{
					begLine:   p.line,
					begCol:    p.column,
					isCharLit: true,
				}
				state = kSTATE_STRING
			case c == '/':
				state = kSTATE_SLASH
			case unicode.IsLetter(c):
				buf = append(buf, c)
				state = kSTATE_IDENT
			default:
				buf = append(buf, c)
				state = kSTATE_OPERATOR
			}
		case kSTATE_ZERO:
			switch {
			case c == '.':
				buf = append(buf, '0')
				state = kSTATE_FLOAT_DOT
			case c == 'x' || c == 'X':
				state = kSTATE_HEX
			default:
				p.ungetch(c)
				t.Value = 0
				t.Symbol = tINT_LITERAL
				return nil
			}
		case kSTATE_HEX:
			if isXDigit(c) {
				buf = append(buf, c)
				state = kSTATE_HEX_DIGITS
			} else {
				return fmt.Errorf("%s:%d:%d: missing number literal digits", p.SrcFileName, p.Line(), p.Column())
			}
		case kSTATE_HEX_DIGITS:
			if !isXDigit(c) {
				p.ungetch(c)
				t.Value = parseInt(string(buf), 16)
				t.Symbol = tINT_LITERAL
				return nil
			}
			buf = append(buf, c)
		case kSTATE_INTEGER:
			switch {
			case c == '.':
				state = kSTATE_FLOAT_DOT
			case unicode.IsDigit(c):
				buf = append(buf, c)
			default:
				p.ungetch(c)
				t.Value = parseInt(string(buf), 10)
				t.Symbol = tINT_LITERAL
				return nil
			}
		case kSTATE_FLOAT_DOT:
			if !unicode.IsDigit(c) {
				p.ungetch(c)
				p.ungetch('.')
				t.Value = parseInt(string(buf), 10)
				t.Symbol = tINT_LITERAL
				return nil
			}
			buf = append(buf, '.')
			buf = append(buf, c)
			state = kSTATE_FLOAT
		case kSTATE_FLOAT:
			switch {
			case unicode.IsDigit(c):
				buf = append(buf, c)
			case c == 'e' || c == 'E':
				buf = append(buf, c)
				state = kSTATE_FLOAT_E
			default:
				p.ungetch(c)
				t.Value = parseFloat(string(buf))
				t.Symbol = tDOUBLE_LITERAL
				return nil
			}
		case kSTATE_FLOAT_E:
			switch {
			case c == '+' || c == '-':
				buf = append(buf, c)
			case unicode.IsDigit(c):
				buf = append(buf, c)
				state = kSTATE_FLOAT_E_DIGITS
			default:
				return fmt.Errorf("missing number literal digits")
			}
		case kSTATE_FLOAT_E_DIGITS:
			if !unicode.IsDigit(c) {
				p.ungetch(c)
				t.Value = parseFloat(string(buf))
				t.Symbol = tDOUBLE_LITERAL
				return nil
			}
			buf = append(buf, c)
		case kSTATE_STRING:
			switch c {
			case '\\':
				state = kSTATE_STRING_ESC
			case '”':
				if !str.isCharLit {
					t.Value = string(buf)
					t.Symbol = tSTRING_LITERAL
					return nil
				}
				buf = append(buf, c)
			case '\'':
				if str.isCharLit {
					t.Value = string(buf)
					t.Symbol = tINT_LITERAL
					return nil
				}
				buf = append(buf, c)
			default:
				buf = append(buf, c)
			}
		case kSTATE_STRING_ESC:
			ch, ok := escapes[c]
			if ok {
				buf = append(buf, ch)
			} else {
				buf = append(buf, '\\')
				buf = append(buf, c)
			}
			state = kSTATE_STRING
		case kSTATE_IDENT:
			if unicode.IsLetter(c) {
				buf = append(buf, c)
			} else {
				p.ungetch(c)
				if sym, ok := keywords[string(buf)]; ok {
					t.Symbol = sym
				} else if op, ok := operators[string(buf)]; ok {
					t.Symbol = op
				} else {
					result := TakeNumberFromString(string(buf))
					switch result.(type) {
					case int:
						t.Value = result
						t.Symbol = tINT_LITERAL
					case float64:
						t.Value = result
						t.Symbol = tDOUBLE_LITERAL
					default:
						t.Value = string(buf)
						t.Symbol = tIDENT
					}
				}
				return nil
			}
		case kSTATE_OPERATOR:
			buf = append(buf, c)
			if !isPrefixOfOperators(string(buf), operators) {
				p.ungetch(c)
				buf = buf[0 : len(buf)-1] // chop
				switch {
				case len(buf) <= 0:
					panic("must not happen")
				case len(buf) == 1:
					t.Symbol = int(buf[0])
					return nil
				default:
					t.Symbol = operators[string(buf)]
					return nil
				}
			}
		case kSTATE_SLASH:
			switch c {
			case '/':
				state = kSTATE_COMMENT_LINE
			case '*':
				state = kSTATE_COMMENT_BLOCK
			default:
				p.ungetch(c)
				buf = append(buf, '/')
				state = kSTATE_OPERATOR
			}
		case kSTATE_COMMENT_LINE:
			if c == '\n' {
				state = kSTATE_INITIAL
			}
		case kSTATE_COMMENT_BLOCK:
			if c == '*' {
				state = kSTATE_COMMENT_STAR
			}
		case kSTATE_COMMENT_STAR:
			if c == '/' {
				state = kSTATE_INITIAL
			} else {
				p.ungetch(c)
				state = kSTATE_COMMENT_BLOCK
			}
		}
	}

	if err == io.EOF && (state == kSTATE_STRING || state == kSTATE_STRING_ESC) {
		return fmt.Errorf("%s:%d:%d: missing string literal termination", p.SrcFileName, str.begLine, str.begCol)
	}

	return err
}

var xdigitLetters = map[rune]bool{
	'a': true,
	'b': true,
	'c': true,
	'd': true,
	'e': true,
	'f': true,
	'A': true,
	'B': true,
	'C': true,
	'D': true,
	'E': true,
	'F': true,
}

func isXDigit(c rune) bool {
	if unicode.IsDigit(c) {
		return true
	}
	_, ok := xdigitLetters[c]
	return ok
}

func parseInt(src string, base int) int {
	v, err := strconv.ParseInt(src, base, 0)
	if err != nil {
		panic(err)
	}
	return int(v)
}

func parseFloat(src string) float64 {
	v, err := strconv.ParseFloat(src, 0)
	if err != nil {
		panic(err)
	}
	return v
}

func isPrefixOfOperators(str string, table map[string]int) bool {
	for k, _ := range table {
		if strings.HasPrefix(k, str) {
			return true
		}
	}
	return false
}

var keywords = map[string]int{
	"如果":   kIF,
	"否则":   kELSE,
	"另外如果": kELSIF,
	"假设":   kSWITCH,
	"若":    kCASE,
	"默认":   kDEFAULT,
	"每当":   kWHILE,
	"做":    kDO,
	"循环":   kFOR,
	"逐个循环": kFOREACH,
	"返回":   kRETURN,
	"中断":   kBREAK,
	"跳过":   kCONTINUE,
	"空值":   kNULL,
	"真":    kTRUE,
	"假":    kFALSE,
	"尝试":   kTRY,
	"捕捉":   kCATCH,
	"最终执行": kFINALLY,
	"抛出异常": kTHROW,
	"声明异常": kTHROWS,
	"布尔型":   kBOOLEAN,
	"空":    kVOID,
	"整型":   kINT,
	"双精度型":  kDOUBLE,
	"字符串型":  kSTRING,
	"野指针":  kNATIVE_POINTER,
	"新建":   kNEW,
	"需要":   kREQUIRE,
	"重命名":  kRENAME,
	"类":    kCLASS,
	"接口":   kINTERFACE,
	"公开":   kPUBLIC,
	"私有":   kPRIVATE,
	"虚拟":   kVIRTUAL,
	"重写":   kOVERRIDE,
	"抽象":   kABSTRACT,
	"自身":   kTHIS,
	"父类":   kSUPER,
	"构造":   kCONSTRUCTOR,
	"实例":   kINSTANCEOF,
	"委派":   kDELEGATE,
	"枚举":   kENUM,
	"终值":   kFINAL,
	"常数":   kCONST,
}

var operators = map[string]int{
	"为":     tIS,
	"并且":    tLOG_AND,
	"或者":    tLOG_OR,
	"等于":    tEQ,
	"不等于":   tNE,
	"大于":  tG,
	"小于":  tL,
	"大于等于":  tGE,
	"小于等于":  tLE,
	"加后赋值":  tADD_A,
	"减后赋值":  tSUB_A,
	"乘后赋值":  tMUL_A,
	"除后赋值":  tDIV_A,
	"取模后赋值": tMOD_A,
	"自加一":    tINC,
	"自减一":    tDEC,
	"：：":    tDCAST_BEG,
	"：》":    tDCAST_END,
}

var escapes = map[rune]rune{
	'\\': '\\',
	'"':  '"',
	'\'': '\'',
	'n':  '\n',
	't':  '\t',
}

type stringLiteral struct {
	begLine   int
	begCol    int
	isCharLit bool
}

type state_t int

const (
	kSTATE_INITIAL state_t = iota
	kSTATE_SLASH
	kSTATE_COMMENT_LINE
	kSTATE_COMMENT_BLOCK
	kSTATE_COMMENT_STAR
	kSTATE_ZERO
	kSTATE_HEX
	kSTATE_HEX_DIGITS
	kSTATE_INTEGER
	kSTATE_FLOAT_DOT
	kSTATE_FLOAT
	kSTATE_FLOAT_E
	kSTATE_FLOAT_E_DIGITS
	kSTATE_STRING
	kSTATE_STRING_ESC
	kSTATE_IDENT
	kSTATE_OPERATOR
)
