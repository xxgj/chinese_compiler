package ast

type Operator int

const (
	ADD Operator = iota
	SUB
	MUL
	DIV
	MOD

	EQ
	NE
	GT
	LT
	GE
	LE
	LOG_OR
	LOG_AND

	INC
	DEC

	NORMAL_ASSIGN
	ADD_ASSIGN
	SUB_ASSIGN
	MUL_ASSIGN
	DIV_ASSIGN
	MOD_ASSIGN
)

var operatorStrings = []string{
	"加",
	"减",
	"乘",
	"除",
	"取模",
	"相等",
	"不相等",
	"大于",
	"小于",
	"大于等于",
	"小于等于",
	"逻辑或",
	"逻辑与",
	"加一",
	"减一",
	"常规赋值",
	"加后赋值",
	"减后赋值",
	"乘后赋值",
	"除后赋值",
	"取模后赋值",
}

func (op Operator) String() string {
	return operatorStrings[op-ADD]
}
