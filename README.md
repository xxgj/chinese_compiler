# chinese_compiler
A toy compiler for chinese language.

## Build

Compilation requires Go 1.13+

`go build -o compiler ./main/parser.go`

## Help

`compiler -h`

## Usage

`compiler --in ./example/for.txt`

You must specify the input file.

Some examples can be found in folder `example`.

## Example

### Code Input

```
字符串型 今天。
今天 为 今天星期几（）。
如果（今天 等于 “星期一”）【
	上程序设计语言原理课（）。
】另外如果（今天 等于 “星期日”）【
	写程序语言设计原理作业（）。
】否则【
	干别的事情（）。
】

```

### AST Output

```
「第1行｜第5列」 -> *ast.Block: 
    「第1行｜第5列」 -> *ast.Decl:
    「第2行｜第12列」 -> *ast.ExprStmt:
        「第2行｜第3列」 -> *ast.assign: 常规赋值
            #左值:
                「第2行｜第0列」 -> *ast.IdentExpr: 今天
            #右值:
                「第2行｜第10列」 -> *ast.Funcall:
                    #函数:
                        「第2行｜第5列」 -> *ast.IdentExpr: 今天星期几
                    #参数:
                        (nil)
    「第3行｜第0列」 -> *ast.If:
        #条件:
            「第3行｜第6列」 -> *ast.Binary: 相等
                #左值:
                    「第3行｜第3列」 -> *ast.IdentExpr: 今天
                #右值:
                    「第3行｜第9列」 -> *ast.StrLiteral: 星期一
        #执行块:
            「第3行｜第15列」 -> *ast.Block:
                「第4行｜第13列」 -> *ast.ExprStmt:
                    「第4行｜第11列」 -> *ast.Funcall:
                        #函数:
                            「第4行｜第1列」 -> *ast.IdentExpr: 上程序设计语言原理课
                        #参数:
                            (nil)
        #后续:
            「第5行｜第1列」 -> *ast.If:
                #条件:
                    「第5行｜第9列」 -> *ast.Binary: 相等
                        #左值:
                            「第5行｜第6列」 -> *ast.IdentExpr: 今天
                        #右值:
                            「第5行｜第12列」 -> *ast.StrLiteral: 星期日
                #执行块:
                    「第5行｜第18列」 -> *ast.Block:
                        「第6行｜第14列」 -> *ast.ExprStmt:
                            「第6行｜第12列」 -> *ast.Funcall:
                                #函数:
                                    「第6行｜第1列」 -> *ast.IdentExpr: 写程序语言设计原理作业
                                #参数:
                                    (nil)
                #后续:
                    「第7行｜第3列」 -> *ast.Block:
                        「第8行｜第8列」 -> *ast.ExprStmt:
                            「第8行｜第6列」 -> *ast.Funcall: 
                                #函数:
                                    「第8行｜第1列」 -> *ast.IdentExpr: 干别的事情
                                #参数:
                                    (nil)

```

