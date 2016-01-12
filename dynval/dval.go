package dynval

import (
	"github.com/zhemao/glisp/interpreter"
	"strings"
)

type DynVal struct {
	Sexp     glisp.Sexp
	Sexp_str string
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

func (dval *DynVal) Execute(env *glisp.Glisp) (glisp.Sexp, error) {
	env.LoadExpressions([]glisp.Sexp{dval.Sexp})
	sexp, err := env.Run()
	if err != nil {
		return glisp.SexpNull, err
	}
	return sexp, nil
}
