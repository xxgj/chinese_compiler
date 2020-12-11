package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kevinchen147/chinese_compiler/ast"
	"github.com/kevinchen147/chinese_compiler/parser"
)

var flagIn string
var flagOut string
var flagLine int

func init() {
	flag.StringVar(&flagIn, "in", "code.txt", "指定输入的代码文件路径")
	flag.StringVar(&flagOut, "out", "", "指定输出的语法树文件路径，默认为标准输出")
	flag.IntVar(&flagLine, "line", 1, "指定解析开始的行数，默认从第一行开始解析")
}

func main() {
	flag.Parse()
	code, err := ioutil.ReadFile(flagIn)
	if err != nil {
		fmt.Println("读文件出错", err)
		return
	}
	if flagOut != "" {
		f, err := os.OpenFile(flagOut, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0777)
		if err != nil {
			fmt.Println("写文件出错", err)
			return
		}
		os.Stdout = f
	}

	p := parser.NewParser(strings.NewReader(string(code)), flagIn, flagLine)
	err = p.Parse()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	ast.Dump(p.Tree, os.Stdout)
}
