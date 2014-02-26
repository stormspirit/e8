package e8asm

import (
	"fmt"
	"github.com/h8liu/e8/vm/inst"
	"math"
	"strconv"
	"strings"
)

type Line struct {
	in     inst.Inst
	label  string
	isJump bool
}

func newLine(in inst.Inst) *Line {
	ret := new(Line)
	ret.in = in
	return ret
}

func ef(s string, args ...interface{}) (*Line, error) {
	return nil, fmt.Errorf(s, args...)
}

func trim(s string) string  { return strings.TrimSpace(s) }
func lower(s string) string { return strings.ToLower(s) }
func fields(args string) []string {
	ret := strings.Split(args, ",")
	for i, s := range ret {
		ret[i] = trim(s)
	}
	return ret
}

func opSplit(s string) (op, args string) {
	firstSpace := strings.IndexAny(s, " \t")
	if firstSpace < 0 {
		op = s
	} else {
		op = s[:firstSpace]
		args = trim(s[firstSpace:])
	}

	return
}

var r3Insts = map[string]uint8{
	"add": inst.FnAdd,
	"sub": inst.FnSub,
	"and": inst.FnAnd,
	"or":  inst.FnOr,
	"xor": inst.FnXor,
	"nor": inst.FnXor,
	"slt": inst.FnSlt,

	"mul":  inst.FnMul,
	"mulu": inst.FnMulu,
	"div":  inst.FnDiv,
	"divu": inst.FnDivu,
	"mod":  inst.FnMod,
	"modu": inst.FnModu,
}

var r3rInsts = map[string]uint8{
	"sllv": inst.FnSllv,
	"srlv": inst.FnSrlv,
	"srav": inst.FnSrav,
}

var r3sInsts = map[string]uint8{
	"sll": inst.FnSll,
	"srl": inst.FnSrl,
	"sra": inst.FnSra,
}

var i3aInsts = map[string]uint8{
	"lw":  inst.OpLw,
	"lhs": inst.OpLhs,
	"lhu": inst.OpLhu,
	"lbs": inst.OpLbs,
	"lbu": inst.OpLbu,
	"sw":  inst.OpSw,
	"sh":  inst.OpSh,
	"sb":  inst.OpSb,
}

var i3Insts = map[string]uint8{
	"addi": inst.OpAddi,
	"andi": inst.OpAndi,
	"ori":  inst.OpOri,
	"slti": inst.OpSlti,
}

var i2Insts = map[string]uint8{
	"lui": inst.OpLui,
}

var bInsts = map[string]uint8{
	"bne": inst.OpBne,
	"beq": inst.OpBeq,
}

func ParseLine(s string) (*Line, error) {
	s = trim(s)
	op, args := opSplit(s)
	op = lower(op)

	if op == "j" {
		return jLine(inst.OpJ, args)
	} else if code, found := bInsts[op]; found {
		return bLine(code, args)
	} else if code, found := i3Insts[op]; found {
		return i3Line(code, args)
	} else if code, found := i3aInsts[op]; found {
		return i3aLine(code, args)
	} else if code, found := i2Insts[op]; found {
		return i2Line(code, args)
	} else if code, found := r3Insts[op]; found {
		return r3Line(code, args)
	} else if code, found := r3rInsts[op]; found {
		return r3rLine(code, args)
	} else if code, found := r3sInsts[op]; found {
		return r3sLine(code, args)
	}

	return nil, fmt.Errorf("invalid op")
}

func isIdent(s string) bool {
	panic("todo")
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 0, 32)
}

func parseReg(s string) (uint8, bool) {
	if len(s) < 2 {
		return 0, false
	}
	s = lower(s)

	if s == "pc" {
		return inst.RegPC, true
	}

	if s[0] != '$' || s[0] != 'r' {
		return 0, false
	}

	n, e := parseInt(s[1:])
	if e != nil {
		return 0, false
	}
	if n < 0 {
		return 0, false
	}
	if n >= inst.Nreg {
		return 0, false
	}

	return uint8(n), true
}

func parseShamt(s string) (uint8, bool) {
	n, e := parseInt(s)
	if e != nil {
		return 0, false
	}
	if n < 0 {
		return 0, false
	}
	if n >= 32 {
		return 0, false
	}
	return uint8(n), true
}

func parseIms(s string) (uint16, bool) {
	n, e := parseInt(s)
	if e != nil {
		return 0, false
	}
	if n < math.MinInt16 {
		return 0, false
	}
	if n > math.MaxInt16 {
		return 0, false
	}
	return uint16(int16(n)), true
}

func parseImu(s string) (uint16, bool) {
	n, e := parseInt(s)
	if e != nil {
		return 0, false
	}
	if n < 0 {
		return 0, false
	}
	if n > math.MaxUint16 {
		return 0, false
	}
	return uint16(n), true
}

func jLine(code uint8, args string) (*Line, error) {
	if !isIdent(args) {
		return ef("invalid label")
	}

	ret := newLine(inst.Jinst(0))
	ret.label = args
	ret.isJump = true

	return ret, nil
}

func bLine(code uint8, args string) (*Line, error) {
	fs := fields(args)
	if len(fs) != 3 {
		return ef("invalid field count")
	}

	rs, valid := parseReg(fs[0])
	if !valid {
		return ef("first field not register")
	}
	rt, valid := parseReg(fs[1])
	if !valid {
		return ef("second field not register")
	}

	label := fs[2]
	if !isIdent(label) {
		return ef("third field is not a label")
	}

	ret := newLine(inst.Iinst(code, rs, rt, 0))
	ret.label = label

	return ret, nil
}

func i3Line(code uint8, args string) (*Line, error) {
	fs := fields(args)
	if len(fs) != 3 {
		return ef("invalid field count")
	}

	rt, valid := parseReg(fs[0])
	if !valid {
		return ef("first field not register")
	}
	rs, valid := parseReg(fs[1])
	if !valid {
		return ef("second field not register")
	}

	im := uint16(0)
	if code == inst.OpSlti {
		im, valid = parseIms(fs[2])
		if !valid {
			return ef("third field not a signed immediate")
		}
	} else {
		im, valid = parseImu(fs[2])
		if !valid {
			return ef("third field not an unsigned immediate")
		}
	}

	ret := newLine(inst.Iinst(code, rs, rt, im))
	return ret, nil
}

func i3aLine(code uint8, args string) (*Line, error) {
	panic("todo")
}

func i2Line(code uint8, args string) (*Line, error) {
	fs := fields(args)
	if len(fs) != 2 {
		return ef("invalid field count")
	}

	rt, valid := parseReg(fs[0])
	if !valid {
		return ef("first field not register")
	}

	im, valid := parseIms(fs[2])
	if !valid {
		return ef("second field not a signed immediate")
	}

	ret := newLine(inst.Iinst(code, 0, rt, im))
	return ret, nil
}

func r3Line(code uint8, args string) (*Line, error) {
	fs := fields(args)
	if len(fs) != 3 {
		return ef("invalid field count")
	}

	rd, valid := parseReg(fs[0])
	if !valid {
		return ef("first field not register")
	}
	rs, valid := parseReg(fs[1])
	if !valid {
		return ef("second field not register")
	}
	rt, valid := parseReg(fs[2])
	if !valid {
		return ef("third field not register")
	}

	ret := newLine(inst.Rinst(rs, rt, rd, code))
	return ret, nil
}

func r3rLine(code uint8, args string) (*Line, error) {
	fs := fields(args)
	if len(fs) != 3 {
		return ef("invalid field count")
	}

	rd, valid := parseReg(fs[0])
	if !valid {
		return ef("first field not register")
	}
	rt, valid := parseReg(fs[1])
	if !valid {
		return ef("second field not register")
	}
	rs, valid := parseReg(fs[2])
	if !valid {
		return ef("third field not register")
	}

	ret := newLine(inst.Rinst(rs, rt, rd, code))
	return ret, nil
}

func r3sLine(code uint8, args string) (*Line, error) {
	fs := fields(args)
	if len(fs) != 3 {
		return ef("invalid field count")
	}

	rd, valid := parseReg(fs[0])
	if !valid {
		return ef("first field not register")
	}
	rt, valid := parseReg(fs[1])
	if !valid {
		return ef("second field not register")
	}
	shamt, valid := parseShamt(fs[2])
	if !valid {
		return ef("third field not shamt")
	}

	ret := newLine(inst.RinstShamt(0, rt, rd, shamt, code))
	return ret, nil
}