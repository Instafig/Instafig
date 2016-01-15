package main

import (
	"errors"
	"github.com/hashicorp/go-version"
	"github.com/zhemao/glisp/interpreter"
)

func defmacroCondValues(env *glisp.Glisp) {
	macro := "(defmac cond-values [ & body] `(cond ~@body))"
	env.EvalString(macro)
}

// version-cmp
func VersionCompareFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var vleft, vright string

	switch t := args[0].(type) {
	case glisp.SexpStr:
		vleft = string(t)
	default:
		return glisp.SexpNull, errors.New("version argument must be string")
	}

	switch t := args[1].(type) {
	case glisp.SexpStr:
		vright = string(t)
	default:
		return glisp.SexpNull, errors.New("version argument must be string")
	}

	v1, err := version.NewVersion(vleft)
	if err != nil {
		return glisp.SexpNull, errors.New("version format error")
	}
	v2, err := version.NewVersion(vright)
	if err != nil {
		return glisp.SexpNull, errors.New("version format error")
	}
	res := v1.Compare(v2)
	return glisp.SexpInt(res), nil
}

func defVersionCompareFunctions(env *glisp.Glisp) {
	env.AddFunction("version-cmp", VersionCompareFunction)
	shortcuts := `
         (defn ver= [v1 v2] (= (version-cmp v1 v2) 0))
         (defn ver> [v1 v2] (> (version-cmp v1 v2) 0))
         (defn ver< [v1 v2] (< (version-cmp v1 v2) 0))
         (defn ver>= [v1 v2] (>= (version-cmp v1 v2) 0))
         (defn ver<= [v1 v2] (<= (version-cmp v1 v2) 0))
         (defn ver!= [v1 v2] (not (= (version-cmp v1 v2) 0)))
    `
	env.EvalString(shortcuts)
}

func NewGlisp() *glisp.Glisp {
	env := glisp.NewGlisp()
	defmacroCondValues(env)
	defVersionCompareFunctions(env)
	return env
}
