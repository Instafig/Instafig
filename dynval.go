package main

import (
	"strings"
	"encoding/json"

	"github.com/zhemao/glisp/interpreter"
)

type DynVal struct {
	Sexp     glisp.Sexp
	Sexp_str string
}

type ExternalSymbol struct {
	Symbol string
}

func NewDynValFromString(str string, env *glisp.Glisp) *DynVal {
	sexp, err := env.ParseStream(strings.NewReader(str))
	if err != nil {
		return nil
	}
	return &DynVal{sexp[0], sexp[0].SexpString()}
}

func NewDynValFromSexp(sexp glisp.Sexp) *DynVal {
	return &DynVal{sexp, sexp.SexpString()}
}

func SetClientData(env *glisp.Glisp, cdata *ClientData) error {
	env.AddGlobal("APP_KEY", glisp.SexpStr(cdata.AppKey))
	env.AddGlobal("OS_TYPE", glisp.SexpStr(cdata.OSType))
	env.AddGlobal("OS_VERSION", glisp.SexpStr(cdata.OSVersion))
	env.AddGlobal("APP_VERSION", glisp.SexpStr(cdata.AppVersion))
	env.AddGlobal("IP", glisp.SexpStr(cdata.Ip))
	env.AddGlobal("LANG", glisp.SexpStr(cdata.Lang))
	env.AddGlobal("DEVICE_ID", glisp.SexpStr(cdata.DeviceId))
	return nil
}

func ClearClientData(env *glisp.Glisp) error {
	env.AddGlobal("APP_KEY", glisp.SexpNull)
	env.AddGlobal("OS_TYPE", glisp.SexpNull)
	env.AddGlobal("OS_VERSION", glisp.SexpNull)
	env.AddGlobal("APP_VERSION", glisp.SexpNull)
	env.AddGlobal("IP", glisp.SexpNull)
	env.AddGlobal("LANG", glisp.SexpNull)
	env.AddGlobal("DEVICE_ID", glisp.SexpNull)
	return nil
}


func (dval *DynVal) Execute(env *glisp.Glisp) (glisp.Sexp, error) {
	env.LoadExpressions([]glisp.Sexp{dval.Sexp})
	sexp, err := env.Run()
	if err != nil {
		return glisp.SexpNull, err
	}
	return sexp, nil
}

func EvalDynValToSexp(code string, cdata *ClientData) (glisp.Sexp, error) {
	env := glisp.NewGlisp()
	SetClientData(env, cdata)
	dval := NewDynValFromString(code, env)
	return dval.Execute(env)
}

func EvalDynVal(code string, cdata *ClientData) interface{} {
	env := glisp.NewGlisp()
	SetClientData(env, cdata)
	dval := NewDynValFromString(code, env)
	data, err := dval.Execute(env)
	if err != nil {
		return nil
	}
	switch data.(type) {
	case glisp.SexpBool:
		return bool(data.(glisp.SexpBool))
	case glisp.SexpInt:
		return int(data.(glisp.SexpInt))
	case glisp.SexpFloat:
		return float64(data.(glisp.SexpFloat))
	case glisp.SexpStr:
		return string(data.(glisp.SexpStr))
	default:
		return nil
	}
}

func sexpToSlice(sexp glisp.Sexp) interface{} {
	if sexp == glisp.SexpNull {
		return nil
	}
	switch sexp.(type){
	case glisp.SexpPair:
		return sexpPairToSlice(sexp.(glisp.SexpPair))
	case glisp.SexpSymbol:
		return ExternalSymbol{(sexp.(glisp.SexpSymbol)).Name()}
	case glisp.SexpBool:
		return bool(sexp.(glisp.SexpBool))
	case glisp.SexpInt:
		return int(sexp.(glisp.SexpInt))
	case glisp.SexpFloat:
		return float64(sexp.(glisp.SexpFloat))
	case glisp.SexpStr:
		return string(sexp.(glisp.SexpStr))
	default:
		return sexp.SexpString()
	}
}

func sexpPairToSlice(pair glisp.SexpPair) []interface{} {
	retv := []interface{}{}

	for {
		switch pair.Tail().(type) {
		case glisp.SexpPair:
			retv = append(retv, sexpToSlice(pair.Head()))
			pair = pair.Tail().(glisp.SexpPair)
			continue
		}
		break
	}

	retv = append(retv, sexpToSlice(pair.Head()))
	// TODO fake list when pair.tail is not SexpNull
	return retv
}

func (dval *DynVal) ToSlice() []interface{} {
	return sexpPairToSlice(dval.Sexp.(glisp.SexpPair))
}

func (dval *DynVal) ToJson() (string, error) {
	data, err := json.Marshal(dval.ToSlice())
	if err != nil {
		return "", err
	}
	return string(data), nil
}
