%{
package parser

import (
    "fmt"
    "github.com/kevinchen147/chinese_compiler/ast"
)

func tokName(c int) string {
    switch c {
    case 0:
        return "EOF"
    case '\n':
        return "\\n"
    }
    i := c - kIF
    if i < 0 || i > len(yyToknames)-1 {
        return fmt.Sprintf("%c", c)
    }
    return yyToknames[i]
}

func addTopLevelStmt(l yyLexer, stmt ast.Node) {
    p := l.(*Parser)
    if p.Tree == nil {
        p.Tree = ast.NewBlock(stmt)
    }
    p.Tree.(*ast.Block).Add(stmt)
}

func defun(l yyLexer, pos *Token, typeSpec baseType, name *Token, params []ast.Node, body ast.Node) {
}

func openBlock(l yyLexer, pos *Token) ast.Node {
    p := l.(*Parser)
    b := ast.NewBlock(pos)
    p.PushBlock(b)    
    return b
}

func closeBlock(l yyLexer, block ast.Node, stmts []ast.Node) ast.Node {
    p := l.(*Parser)
    b := p.PopBlock()
    if b != block {
        panic("block stack might be broken")
    }
    b.Append(stmts)
    return b
}


%}

%union {
	token    *Token
	ident     string
	node      ast.Node
	nodes     []ast.Node
	typeSpec  baseType
}

%token<token> kIF kELSE kELSIF kSWITCH kCASE kDEFAULT kWHILE kDO kFOR kFOREACH
              kRETURN kBREAK kCONTINUE kNULL kTRUE kFALSE kTRY kCATCH kFINALLY
              kTHROW kTHROWS kBOOLEAN kVOID kINT kDOUBLE kSTRING kNATIVE_POINTER
              kNEW kREQUIRE kRENAME kCLASS kINTERFACE kPUBLIC kPRIVATE kVIRTUAL
              kOVERRIDE kABSTRACT kTHIS kSUPER kCONSTRUCTOR kINSTANCEOF
              kDELEGATE kENUM kFINAL kCONST tIS tEQ tNE tGE tLE tADD_A tSUB_A
              tMUL_A tDIV_A tMOD_A tINC tDEC tDCAST_BEG tDCAST_END tLOG_AND tLOG_OR
              tINT_LITERAL tDOUBLE_LITERAL tSTRING_LITERAL tREGEXP_LITERAL tIDENT

%type<node>  stmt block_beg block if_tail opt_expr expr assign log_or log_and
             equality relational additive multive unary postfix primary
%type<nodes> stmts params args
%type<token> '【' '。' '，' '《' '》' '+' '-' '*' '/' '%' '！' '（'
%type<ident> opt_label opt_ident
%type<typeSpec> type_spec




%%

trans_unit : def_or_stmt
           | trans_unit def_or_stmt

def_or_stmt : def_func
            | stmt
    {
        addTopLevelStmt(yylex, $1)
    }

type_spec : kBOOLEAN { $$ = ttBOOLEAN }
          | kINT     { $$ = ttINT }
          | kDOUBLE  { $$ = ttDOUBLE }
          | kSTRING  { $$ = ttSTRING }

def_func : type_spec tIDENT '（' params '）' block
    {
        defun(yylex, $2, $1, $2, $4, $6)
    }
    	 | type_spec tIDENT '（' '）' block
    {
        defun(yylex, $2, $1, $2, nil, $5)
    }
         | type_spec tIDENT '（' params '）' '。'
    {
        defun(yylex, $2, $1, $2, $4, nil)
    }
         | type_spec tIDENT '（' '）' '。'
 	{
        defun(yylex, $2, $1, $2, nil, nil)
    }

params : type_spec tIDENT
    {
        $$ = append(make([]ast.Node, 0), ast.NewParam($2, newTypeSpec($1), $2.Value.(string)))
    }
       | params '，' type_spec tIDENT
    {
        $$ = append($1, ast.NewParam($4, newTypeSpec($3), $4.Value.(string)))
    }

block_beg : '【'
    {
	$$ = openBlock(yylex, $1)
    }

block : block_beg '】'
    {
        $$ = closeBlock(yylex, $1, nil)
    }
      | block_beg stmts '】'
    {
        $$ = closeBlock(yylex, $1, $2)
    }

stmts : stmt
    {
        $$ = append(make([]ast.Node, 0), $1)
    }
      | stmts stmt
    {
        $$ = append($1, $2)
    }

stmt : expr '。'
    {
        $$ = ast.NewExprStmt($2, $1)
    }
     | type_spec tIDENT '。'
    {
        $$ = ast.NewDecl($2, newTypeSpec($1), $2.Value.(string), nil)
    }
     | type_spec tIDENT tIS expr '。'
    {
        $$ = ast.NewDecl($2, newTypeSpec($1), $2.Value.(string), $4)
    }
     | kIF '（' expr '）' block if_tail
    {
        $$ = ast.NewIf($1, $3, $5, $6)
    }
     | opt_label kWHILE '（' expr '）' block
    {
        $$ = ast.NewWhile($2, $1, $4, $6)
    }
     | opt_label kFOR '（' opt_expr '；' opt_expr '；' opt_expr '）' block
    {
        $$ = ast.NewFor($2, $1, $4, $6, $8, $10)
    }
     | opt_label kFOREACH '（' tIDENT '：' expr '）' block
    {
        $$ = ast.NewForeach($2, $1, $4.Value.(string), $6, $8)
    }
     | kRETURN opt_expr '。'
    {
        $$ = ast.NewReturn($1, $2)
    }
     | kBREAK opt_ident '。'
    {
        $$ = ast.NewBreak($1, $2)
    }
     | kCONTINUE opt_ident '。'
    {
        $$ = ast.NewContinue($1, $2)
    }
     | kTRY block kCATCH '（' tIDENT '）' block
    {
        $$ = ast.NewTry($1, $2, $5.Value.(string), $7, nil)
    }
     | kTRY block kFINALLY block
    {
        $$ = ast.NewTry($1, $2, "", nil, $4)
    }
     | kTRY block kCATCH '（' tIDENT '）' block kFINALLY block
    {
        $$ = ast.NewTry($1, $2, $5.Value.(string), $7, $9)
    }
     | kTHROW expr '。'
    {
        $$ = ast.NewThrow($1, $2)
    }

if_tail :
    {
        $$ = nil
    }
        | kELSE block
    {
        $$ = $2
    }
        | kELSIF '（' expr '）' block if_tail
    {
        $$ = ast.NewIf($1, $3, $5, $6)
    }

opt_label :
    {
        $$ = ""
    }
          | tIDENT '：'
    {
	$$ = $1.Value.(string)
    }

opt_expr :
    {
        $$ = nil
    }
         | expr

opt_ident :
    {
        $$ = ""
    }
          | tIDENT
	{
		$$ = $1.Value.(string)
	}

expr : assign
     | expr '，' assign
    {
        $$ = ast.NewCommaExpr($2, $1, $3)
    }

assign : log_or
       | postfix tIS assign
    {
        $$ = ast.NewAssign($2, ast.NORMAL_ASSIGN, $1, $3)
    }
       | postfix tADD_A assign
    {
        $$ = ast.NewAssign($2, ast.ADD_ASSIGN, $1, $3)
    }
       | postfix tSUB_A assign
    {
        $$ = ast.NewAssign($2, ast.SUB_ASSIGN, $1, $3)
    }
       | postfix tMUL_A assign
    {
        $$ = ast.NewAssign($2, ast.MUL_ASSIGN, $1, $3)
    }
       | postfix tDIV_A assign
    {
        $$ = ast.NewAssign($2, ast.DIV_ASSIGN, $1, $3)
    }
       | postfix tMOD_A assign
    {
        $$ = ast.NewAssign($2, ast.MOD_ASSIGN, $1, $3)
    }

log_or : log_and
       | log_or tLOG_OR log_and
    {
        $$ = ast.NewBinary($2, ast.LOG_OR, $1, $3)
    }

log_and : equality
        | log_and tLOG_AND equality
    {
        $$ = ast.NewBinary($2, ast.LOG_AND, $1, $3)
    }

equality : relational
         | equality tEQ relational
    {
        $$ = ast.NewBinary($2, ast.EQ, $1, $3)
    }
         | equality tNE relational
    {
        $$ = ast.NewBinary($2, ast.NE, $1, $3)
    }

relational : additive
           | relational '《' additive
    {
        $$ = ast.NewBinary($2, ast.GT, $1, $3)
    }
           | relational '》' additive
    {
        $$ = ast.NewBinary($2, ast.LT, $1, $3)
    }
           | relational tGE additive
    {
        $$ = ast.NewBinary($2, ast.GE, $1, $3)
    }
           | relational tLE additive
    {
        $$ = ast.NewBinary($2, ast.LE, $1, $3)
    }

additive : multive
         | additive '+' multive
    {
        $$ = ast.NewBinary($2, ast.ADD, $1, $3)
    }
         | additive '-' multive
    {
        $$ = ast.NewBinary($2, ast.SUB, $1, $3)
    }

multive : unary
        | multive '*' unary
    {
        $$ = ast.NewBinary($2, ast.MUL, $1, $3)
    }
        | multive '/' unary
    {
        $$ = ast.NewBinary($2, ast.DIV, $1, $3)
    }
        | multive '%' unary
    {
        $$ = ast.NewBinary($2, ast.MOD, $1, $3)
    }

unary : postfix
      | '-' unary
    {
        $$ = ast.NewMinusExpr($1, $2)
    }
      | '！' unary
    {
        $$ = ast.NewLogNot($1, $2)
    }

postfix : primary
        | postfix '（' '）'
    {
        $$ = ast.NewFuncall($2, $1, nil)
    }
        | postfix '（' args '）'
    {
        $$ = ast.NewFuncall($2, $1, $3)
    }
        | postfix tINC
    {
        $$ = ast.NewIncDec($2, $1, ast.INC)
    }
        | postfix tDEC
    {
        $$ = ast.NewIncDec($2, $1, ast.DEC)
    }

primary : '（' expr '）'
    {
        $$ = $2
    }
        | tIDENT
    {
        $$ = ast.NewIdentExpr($1, $1.Value.(string))
    }
        | tINT_LITERAL
    {
        $$ = ast.NewIntLiteral($1, $1.Value.(int))
    }
        | tDOUBLE_LITERAL
    {
        $$ = ast.NewDoubleLiteral($1, $1.Value.(float64))
    }
        | tSTRING_LITERAL
    {
        $$ = ast.NewStrLiteral($1, $1.Value.(string))
    }
        | kTRUE
    {
        $$ = ast.NewBooleanLiteral($1, true)
    }
        | kFALSE
    {
        $$ = ast.NewBooleanLiteral($1, false)
    }

args : assign
    {
        $$ = append(make([]ast.Node, 0), $1)
    }
     | args '，' assign
    {
        $$ = append($1, $3)
    }


%%
