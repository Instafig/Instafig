package main

import (
	"github.com/hashicorp/go-version"
	"github.com/zhemao/glisp/interpreter"
	"regexp"
	"unicode/utf8"
)

func defmacroCondValues(env *glisp.Glisp) {
	macro := "(defmac cond-values [ & body] `(cond ~@body))"
	env.EvalString(macro)
}

// version-cmp
func versionCompareFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var vleft, vright string

	switch t := args[0].(type) {
	case glisp.SexpStr:
		vleft = string(t)
	default:
		//return glisp.SexpNull, errors.New("version argument must be string")
		return glisp.SexpNull, nil
	}

	switch t := args[1].(type) {
	case glisp.SexpStr:
		vright = string(t)
	default:
		//return glisp.SexpNull, errors.New("version argument must be string")
		return glisp.SexpNull, nil
	}

	v1, err := version.NewVersion(vleft)
	if err != nil {
		return glisp.SexpNull, nil
		//return glisp.SexpNull, errors.New("version format error")
	}
	v2, err := version.NewVersion(vright)
	if err != nil {
		return glisp.SexpNull, nil
		//return glisp.SexpNull, errors.New("version format error")
	}
	res := v1.Compare(v2)
	return glisp.SexpInt(res), nil
}

func defVersionCompareFunctions(env *glisp.Glisp) {
	env.AddFunction("version-cmp", versionCompareFunction)
	shortcuts := `
         (defn ver= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (= ret 0) false)))
         (defn ver> [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (> ret 0) false)))
         (defn ver< [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (< ret 0) false)))
         (defn ver>= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (>= ret 0) false)))
         (defn ver<= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (<= ret 0) false)))
         (defn ver!= [v1 v2] (let [ret (version-cmp v1 v2)] (cond (int? ret) (not= ret 0) false)))
    `
	env.EvalString(shortcuts)
}

// string functions

func wildcardMatchFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var pattern, target string
	var regexp_pattern string = ""

	switch t := args[0].(type) {
	case glisp.SexpStr:
		pattern = string(t)
	default:
		//return glisp.SexpNull, errors.New("wildcard parttern must be string")
		return glisp.SexpNull, nil
	}

	switch t := args[1].(type) {
	case glisp.SexpStr:
		target = string(t)
	default:
		//return glisp.SexpNull, errors.New("arg1 must be string")
		return glisp.SexpNull, nil
	}

	for i, w, l := 0, 0, 0; i < len(pattern); i += w {
		runeValue, width := utf8.DecodeRuneInString(pattern[i:])
		if runeValue == '*' {
			regexp_pattern += `\Q`
			regexp_pattern += pattern[l:i]
			regexp_pattern += `\E`
			regexp_pattern += `.*`
			l = i + width
		} else if runeValue == '?' {
			regexp_pattern += `\Q`
			regexp_pattern += pattern[l:i]
			regexp_pattern += `\E`
			regexp_pattern += `.`
			l = i + width
		} else if i == len(pattern)-1 {
			regexp_pattern += `\Q`
			regexp_pattern += pattern[l:]
			regexp_pattern += `\E`
		}
		w = width
	}
	match, err := regexp.MatchString("^"+regexp_pattern+"$", target)
	return glisp.SexpBool(match), err
}

func defStringFunctions(env *glisp.Glisp) {
	env.AddFunction("wildcard", wildcardMatchFunction)
	shortcuts := `
         (defn wildcard-not [v1 v2] (not (wildcard v1 v2)))
    `
	env.EvalString(shortcuts)
}

func NewGlisp() *glisp.Glisp {
	env := glisp.NewGlisp()
	defmacroCondValues(env)
	defVersionCompareFunctions(env)
	defStringFunctions(env)
	return env
}
