package validate

import (
	"fmt"
	"strings"

	sysl "github.com/anz-bank/sysl/src/proto"
	ebnf "github.com/anz-bank/sysl/sysl2/proto"
)

func buildSyslGrammar(ebnfGrammar *ebnf.Grammar) *sysl.Application {
	app := &sysl.Application{
		Name:  &sysl.AppName{Part: []string{ebnfGrammar.Name}},
		Types: map[string]*sysl.Type{},
	}

	for ruleName, rule := range ebnfGrammar.Rules {
		generateTypes(ebnfGrammar, app, ruleName, rule.GetChoices())
	}

	return app
}

func generateTypes(grammar *ebnf.Grammar, app *sysl.Application, ruleName string, choices *ebnf.Choice) {
	if c, _ := getCount(choices); c == 1 {
		attrDefs := map[string]*sysl.Type{}
		for _, term := range choices.GetSequence()[0].GetTerm() {
			switch t := term.Atom.Union.(type) {
			case *ebnf.Atom_Rulename:
				name, syslType := generateField(term, grammar)
				attrDefs[name] = syslType
			case *ebnf.Atom_Choices:
				generateTypes(grammar, app, ruleName, t.Choices)
			default:
				continue
			}
		}

		if t, ok := app.Types[ruleName]; ok {
			for k, v := range t.Type.(*sysl.Type_Tuple_).Tuple.AttrDefs {
				attrDefs[k] = v
			}
		}

		app.Types[ruleName] = &sysl.Type{
			Type: &sysl.Type_Tuple_{
				Tuple: &sysl.Type_Tuple{
					AttrDefs: attrDefs,
				},
			},
		}
	} else if c > 1 {
		genChoices(grammar, app, choices.GetSequence(), ruleName)
	}
}

func generateField(term *ebnf.Term, grammar *ebnf.Grammar) (string, *sysl.Type) {
	ruleName := term.Atom.Union.(*ebnf.Atom_Rulename)
	name := ruleName.Rulename.Name
	if _, ok := grammar.Rules[ruleName.Rulename.Name]; ok {
		syslType := &sysl.Type{
			Type: &sysl.Type_TypeRef{
				TypeRef: &sysl.ScopedRef{
					Ref: &sysl.Scope{
						Path: []string{ruleName.Rulename.Name},
					},
				},
			},
		}
		return name, genType(term.Quantifier, syslType)
	} else {
		syslType := &sysl.Type{
			Type: &sysl.Type_Primitive_{
				Primitive: sysl.Type_STRING,
			},
		}
		return name, genType(term.Quantifier, syslType)
	}
}

func genType(quantifier *ebnf.Quantifier, t *sysl.Type) *sysl.Type {
	if quantifier != nil {
		switch quantifier.Union.(type) {
		case *ebnf.Quantifier_Optional:
			t.Opt = true
			return t
		case *ebnf.Quantifier_OnePlus:
			return &sysl.Type{
				Type: &sysl.Type_List_{
					List: &sysl.Type_List{
						Type: t,
					},
				},
				Opt: false,
			}
		case *ebnf.Quantifier_ZeroPlus:
			return &sysl.Type{
				Type: &sysl.Type_Sequence{
					Sequence: t,
				},
				Opt: true,
			}
		default:
			t.Opt = false
		}
	}
	return t
}

func genChoices(grammar *ebnf.Grammar, app *sysl.Application, seq []*ebnf.Sequence, ruleName string) {
	choiceName := "__Choice_" + ruleName
	app.Types[ruleName] = &sysl.Type{
		Type: &sysl.Type_Tuple_{
			Tuple: &sysl.Type_Tuple{
				AttrDefs: map[string]*sysl.Type{
					choiceName: {
						Type: &sysl.Type_TypeRef{
							TypeRef: &sysl.ScopedRef{
								Ref: &sysl.Scope{Path: []string{choiceName}},
							},
						},
					},
				},
			},
		},
	}

	var types []*sysl.Type

	for _, terms := range seq {
		if ruleNameCount(terms.Term) == 1 {
			atomRuleName := terms.Term[0].Atom.Union.(*ebnf.Atom_Rulename).Rulename.Name
			types = append(types, &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{
							Path: []string{atomRuleName},
						},
					},
				},
			})

			if _, ok := grammar.Rules[atomRuleName]; !ok {
				app.Types[atomRuleName] = &sysl.Type{
					Type: &sysl.Type_Primitive_{
						Primitive: sysl.Type_STRING,
					},
				}
			}
		} else {
			var ruleNames []string
			for _, term := range terms.Term {
				if atomRuleName, ok := term.Atom.Union.(*ebnf.Atom_Rulename); ok {
					ruleNames = append(ruleNames, atomRuleName.Rulename.Name)
				} else {
					fmt.Println("skipped:", term.Atom.Union)
				}
			}
			combinationName := "__Choice_Combination_" + ruleName + "_" + strings.Join(ruleNames, "_")
			types = append(types, &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{
							Path: []string{combinationName},
						},
					},
				},
			})
			genCombination(grammar, app, terms.GetTerm(), combinationName)
		}
	}

	app.Types[choiceName] = &sysl.Type{
		Type: &sysl.Type_OneOf_{
			OneOf: &sysl.Type_OneOf{
				Type: types,
			},
		},
	}
}

func genCombination(grammar *ebnf.Grammar, app *sysl.Application, terms []*ebnf.Term, combinationName string) {
	attrDefs := map[string]*sysl.Type{}

	for _, term := range terms {
		if _, ok := term.Atom.Union.(*ebnf.Atom_Rulename); ok {
			name, syslType := generateField(term, grammar)
			attrDefs[name] = syslType
		}
	}

	app.Types[combinationName] = &sysl.Type{
		Type: &sysl.Type_Tuple_{
			Tuple: &sysl.Type_Tuple{
				AttrDefs: attrDefs,
			},
		},
	}
}

func ruleSequenceCount(seq []*ebnf.Sequence) int {
	count := 0
	for _, terms := range seq {
		for _, term := range terms.Term {
			hasRule := false
			switch t := term.Atom.Union.(type) {
			case *ebnf.Atom_Rulename:
				hasRule = true
			case *ebnf.Atom_Choices:
				hasRule = ruleSequenceCount(t.Choices.Sequence) > 0
			default:
				continue
			}

			if hasRule {
				count++
				break
			}
		}
	}
	return count
}

func ruleNameCount(terms []*ebnf.Term) int {
	count := 0
	for _, term := range terms {
		if _, ok := term.Atom.Union.(*ebnf.Atom_Rulename); ok {
			count++
		}
	}

	return count
}

func getCount(choices *ebnf.Choice) (ruleSeqCount, ruleNameCount int) {
	ruleSeqCount, ruleNameCount = 0, 0
	for _, terms := range choices.Sequence {
		for _, term := range terms.Term {
			switch t := term.Atom.Union.(type) {
			case *ebnf.Atom_Rulename:
				ruleNameCount++
			case *ebnf.Atom_Choices:
				_, rnc := getCount(t.Choices)
				//ruleSeqCount += rsc
				ruleNameCount += rnc
			default:
				continue
			}
		}
		if ruleNameCount > 0 {
			ruleSeqCount++
		}
	}
	return
}
