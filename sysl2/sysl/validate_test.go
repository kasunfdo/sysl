package main

import (
	"flag"
	"testing"

	"github.com/anz-bank/sysl/src/proto"
	"github.com/stretchr/testify/assert"
)

func TestValidatorGetTypeName(t *testing.T) {
	cases := map[string]struct {
		input    *sysl.Type
		expected string
	}{
		"Primitive string": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_STRING}},
			expected: "STRING"},
		"Primitive bool": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_BOOL}},
			expected: "BOOL"},
		"Primitive int": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_INT}},
			expected: "INT"},
		"Primitive float": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_FLOAT}},
			expected: "FLOAT"},
		"Primitive decimal": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_DECIMAL}},
			expected: "DECIMAL"},
		"Primitive no type": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_EMPTY}},
			expected: "EMPTY"},
		"Sequence of primitives": {
			input: &sysl.Type{Type: &sysl.Type_Sequence{
				Sequence: &sysl.Type{Type: &sysl.Type_Primitive_{Primitive: sysl.Type_INT}}}},
			expected: "INT"},
		"Sequence of ref": {
			input: &sysl.Type{Type: &sysl.Type_Sequence{
				Sequence: &sysl.Type{Type: &sysl.Type_TypeRef{TypeRef: &sysl.ScopedRef{
					Ref: &sysl.Scope{Path: []string{"RefType"}}}}}}},
			expected: "RefType"},
		"Ref": {
			input: &sysl.Type{Type: &sysl.Type_TypeRef{
				TypeRef: &sysl.ScopedRef{Ref: &sysl.Scope{Path: []string{"RefType"}}}}},
			expected: "RefType"},
		"Unknown": {
			input:    &sysl.Type{Type: &sysl.Type_Map_{}},
			expected: "Unknown"},
		"Nil": {
			input:    nil,
			expected: "Unknown"},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			typeName := getTypeName(input)
			assert.Equal(t, expected, typeName, "Unexpected result")
		})
	}
}

func TestValidatorIsCollectionType(t *testing.T) {
	cases := map[string]struct {
		input    *sysl.Type
		expected bool
	}{
		"Sequence": {
			input:    &sysl.Type{Type: &sysl.Type_Sequence{}},
			expected: true},
		"Map": {
			input:    &sysl.Type{Type: &sysl.Type_Map_{}},
			expected: true},
		"Set": {
			input:    &sysl.Type{Type: &sysl.Type_Set{}},
			expected: true},
		"List": {
			input:    &sysl.Type{Type: &sysl.Type_List_{}},
			expected: true},
		"Primitive string": {
			input:    &sysl.Type{Type: &sysl.Type_Primitive_{}},
			expected: false},
		"Ref": {
			input: &sysl.Type{Type: &sysl.Type_TypeRef{
				TypeRef: &sysl.ScopedRef{Ref: &sysl.Scope{Path: []string{"RefType"}}}}},
			expected: false},
		"Unknown": {
			input:    &sysl.Type{Type: &sysl.Type_NoType_{}},
			expected: false},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			typeName := isCollectionType(input)
			assert.Equal(t, expected, typeName, "Unexpected result")
		})
	}
}

func TestValidatorValidateEntryPoint(t *testing.T) {
	start := "EntryPoint"
	p := NewParser()
	transform, _ := loadTransform("tests", "transform1.sysl", p)

	var entryPointView *sysl.View
	var nonEntryPointView *sysl.View
	var invalidEntryPointView *sysl.View

	entryPointView = transform.Views[start]
	nonEntryPointView = transform.Views["TfmDefaultEmpty"]
	invalidEntryPointView = transform.Views["EntryPointInvalid"]

	cases := map[string]struct {
		input    *sysl.Application
		expected []Msg
	}{
		"Exists": {
			input: &sysl.Application{
				Views: map[string]*sysl.View{start: entryPointView, "nonEntryPoint": nonEntryPointView}},
			expected: nil},
		"Not exists": {
			input: &sysl.Application{
				Views: map[string]*sysl.View{"nonEntryPoint": nonEntryPointView}},
			expected: []Msg{
				{MessageID: ErrEntryPointUndefined, MessageData: []string{start}}}},
		"Incorrect output": {
			input: &sysl.Application{
				Views: map[string]*sysl.View{start: invalidEntryPointView, "nonEntryPoint": nonEntryPointView}},
			expected: []Msg{
				{MessageID: ErrInvalidEntryPointReturn,
					MessageData: []string{start, start}}}},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(nil, input, p.InferredAssigns(), p.InferredLets())
			validator.validateEntryPoint(start)
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidateFileName(t *testing.T) {
	viewName := "filename"
	p := NewParser()
	transform, _ := loadTransform("tests", "transform1.sysl", p)

	var fileNameView *sysl.View
	var nonFileNameView *sysl.View
	var invalidFileNameView1 *sysl.View
	var invalidFileNameView2 *sysl.View
	var invalidFileNameView3 *sysl.View

	fileNameView = transform.Views[viewName]
	nonFileNameView = transform.Views["TfmDefaultEmpty"]
	invalidFileNameView1 = transform.Views["TfmFilenameInvalid1"]
	invalidFileNameView2 = transform.Views["TfmFilenameInvalid2"]
	invalidFileNameView3 = transform.Views["TfmFilenameInvalid3"]

	cases := map[string]struct {
		input    *sysl.Application
		expected []Msg
	}{
		"Exists": {
			input: &sysl.Application{
				Views: map[string]*sysl.View{viewName: fileNameView, "nonEntryPoint": nonFileNameView}},
			expected: nil},
		"Not exists": {
			input:    &sysl.Application{Views: map[string]*sysl.View{"tfmDefaultEmpty": nonFileNameView}},
			expected: []Msg{{MessageID: ErrUndefinedView, MessageData: []string{viewName}}}},
		"Incorrect output": {
			input:    &sysl.Application{Views: map[string]*sysl.View{viewName: invalidFileNameView1}},
			expected: []Msg{{MessageID: ErrInvalidReturn, MessageData: []string{viewName, "string"}}}},
		"Incorrect assignment": {
			input:    &sysl.Application{Views: map[string]*sysl.View{viewName: invalidFileNameView2}},
			expected: []Msg{{MessageID: ErrMissingReqField, MessageData: []string{viewName, viewName, "string"}}}},
		"Excess assignment": {
			input: &sysl.Application{Views: map[string]*sysl.View{viewName: invalidFileNameView3}},
			expected: []Msg{
				{MessageID: ErrExcessAttr, MessageData: []string{"foo", viewName, "string"}}}},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(nil, input, p.InferredAssigns(), p.InferredLets())
			validator.validateFileName()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorHasSameType(t *testing.T) {
	type inputData struct {
		type1 *sysl.Type
		type2 *sysl.Type
	}
	cases := map[string]struct {
		input    inputData
		expected bool
	}{
		"Same primitive types": {
			input:    inputData{type1: typeString(), type2: typeString()},
			expected: true},
		"Different primitive types1": {
			input:    inputData{type1: typeString(), type2: typeInt()},
			expected: false},
		"Different primitive types2": {
			input:    inputData{type1: typeInt(), type2: typeString()},
			expected: false},
		"Same transform typerefs1": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"Statement"}},
					},
				},
			}, type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"Statement"}},
					},
				},
			}},
			expected: true},
		"Different transform typerefs1-1": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"Statement"}},
					},
				},
			}, type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"StatementList"}},
					},
				},
			}},
			expected: false},
		"Different transform typerefs1-2": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"StatementList"}},
					},
				},
			}, type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"Statement"}},
					},
				},
			}},
			expected: false},
		"Same transform typerefs2": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"Statement"}}},
					},
				},
			}, type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"Statement"}}},
					},
				},
			}},
			expected: true},
		"Different transform typerefs2-1": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"Statement"}}},
					},
				},
			}, type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"StatementList"}}},
					},
				},
			}},
			expected: false},
		"Different transform typerefs2-2": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"StatementList"}}},
					},
				},
			}, type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"Statement"}}},
					},
				},
			}},
			expected: false},
		"Different types1": {
			input:    inputData{type1: typeNone(), type2: typeString()},
			expected: false},
		"Different types2": {
			input:    inputData{type1: typeString(), type2: typeNone()},
			expected: false},
		"Different types3": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"Statement"}}},
					},
				},
			}, type2: typeString()},
			expected: false},
		"Different types3.5": {
			input: inputData{type2: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Appname: &sysl.AppName{Part: []string{"Statement"}}},
					},
				},
			}, type1: typeString()},
			expected: false},
		"Different types4": {
			input: inputData{type1: &sysl.Type{
				Type: &sysl.Type_TypeRef{
					TypeRef: &sysl.ScopedRef{
						Ref: &sysl.Scope{Path: []string{"StatementList"}},
					},
				},
			}, type2: typeString()},
			expected: false},
		"Nil types": {
			input:    inputData{type1: nil, type2: nil},
			expected: false},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			isSame := hasSameType(input.type1, input.type2)
			assert.True(t, expected == isSame, "Unexpected result")
		})
	}
}

/*
func TestValidatorResolveExprType(t *testing.T) {
	type inputData struct {
		viewName string
		expr     *sysl.Expr
	}
	type expectedData struct {
		syslType *sysl.Type
		messages []Msg
	}

	expressions := map[string]*sysl.Expr{}

	transform, _ := loadAndGetDefaultApp("tests", "transform1.sysl")

	for _, tfm := range transform.GetApps() {
		for _, stmt := range tfm.Views["varTypeResolve"].GetExpr().GetTransform().GetStmt() {
			expressions[stmt.GetAssign().GetName()] = stmt.GetAssign().GetExpr()
		}
	}

	cases := map[string]struct {
		input    inputData
		expected expectedData
	}{
		"String": {
			input:    inputData{viewName: "varTypeResolve", expr: expressions["stringType"]},
			expected: expectedData{syslType: typeString(), messages: nil}},
		"Int": {
			input:    inputData{viewName: "varTypeResolve", expr: expressions["intType"]},
			expected: expectedData{syslType: typeInt(), messages: nil}},
		"Bool": {
			input:    inputData{viewName: "varTypeResolve", expr: expressions["boolType"]},
			expected: expectedData{syslType: typeBool(), messages: nil}},
		"Transform type primitive": {
			input: inputData{viewName: "varTypeResolve", expr: expressions["transformTypePrimitive"]},
			expected: expectedData{
				syslType: &sysl.Type{
					Type: &sysl.Type_Tuple_{},
				}, messages: nil}},
		"Transform type ref": {
			input: inputData{viewName: "varTypeResolve", expr: expressions["transformTypeRef"]},
			expected: expectedData{
				syslType: &sysl.Type{
					Type: &sysl.Type_Tuple_{
						Tuple: &sysl.Type_Tuple{
							AttrDefs: map[string]*sysl.Type{"VarDecl": {
								Type: &sysl.Type_Tuple_{
									Tuple: &sysl.Type_Tuple{},
								}}}}},
				}, messages: nil}},
		"Valid bool unary result": {
			input:    inputData{viewName: "varTypeResolve", expr: expressions["unaryResultValidBool"]},
			expected: expectedData{syslType: typeBool(), messages: nil}},
		"Valid int unary result": {
			input:    inputData{viewName: "varTypeResolve", expr: expressions["unaryResultValidInt"]},
			expected: expectedData{syslType: typeInt(), messages: nil}},
		"Invalid unary result bool": {
			input: inputData{viewName: "varTypeResolve", expr: expressions["unaryResultInvalidBool"]},
			expected: expectedData{
				syslType: typeBool(),
				messages: []Msg{{
					MessageID:   ErrInvalidUnary,
					MessageData: []string{"varTypeResolve", "STRING"}}}}},
		"Invalid unary result int": {
			input: inputData{viewName: "varTypeResolve", expr: expressions["unaryResultInvalidInt"]},
			expected: expectedData{
				syslType: typeInt(),
				messages: []Msg{{
					MessageID:   ErrInvalidUnary,
					MessageData: []string{"varTypeResolve", "STRING"}}}}},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			resolver := NewResolver(nil)
			syslType := resolver.resolveExprType(input.expr, input.viewName, "", nil)
			messages := resolver.GetMessages()
			assert.True(t, hasSameType(expected.syslType, syslType), "Unexpected result")
			assert.Equal(t, expected.messages, messages, "Unexpected result")
		})
	}
}*/

func TestValidatorValidateViews(t *testing.T) {
	p := NewParser()
	transform, _ := loadTransform("tests", "transform1.sysl", p)
	grammar, _ := loadGrammar("tests/grammar.sysl")

	cases := map[string]struct {
		inputAssign map[string]TypeData
		inputLet    map[string]TypeData
		expected    []Msg
	}{
		"Equal": {
			inputAssign: map[string]TypeData{"TfmValid": p.InferredAssigns()["TfmValid"]},
			inputLet:    map[string]TypeData{"TfmValid": p.InferredLets()["TfmValid"]},
			expected:    nil},
		"Not Equal": {
			inputAssign: map[string]TypeData{"TfmInvalid": p.InferredAssigns()["TfmInvalid"]},
			inputLet:    map[string]TypeData{"TfmInvalid": p.InferredLets()["TfmInvalid"]},
			expected: []Msg{
				{MessageID: ErrMissingReqField, MessageData: []string{"FunctionName", "TfmInvalid", "MethodDecl"}}}},
		"Absent optional": {
			inputAssign: map[string]TypeData{"TfmNoOptional": p.InferredAssigns()["TfmNoOptional"]},
			inputLet:    map[string]TypeData{"TfmNoOptional": p.InferredLets()["TfmNoOptional"]},
			expected:    nil},
		"Excess attributes without optionals": {
			inputAssign: map[string]TypeData{"TfmExcessAttrs1": p.InferredAssigns()["TfmExcessAttrs1"]},
			inputLet:    map[string]TypeData{"TfmExcessAttrs1": p.InferredLets()["TfmExcessAttrs1"]},
			expected: []Msg{
				{MessageID: ErrExcessAttr, MessageData: []string{"ExcessAttr1", "TfmExcessAttrs1", "MethodDecl"}}}},
		"Excess attributes with optionals": {
			inputAssign: map[string]TypeData{"TfmExcessAttrs2": p.InferredAssigns()["TfmExcessAttrs2"]},
			inputLet:    map[string]TypeData{"TfmExcessAttrs2": p.InferredLets()["TfmExcessAttrs2"]},
			expected: []Msg{
				{MessageID: ErrExcessAttr, MessageData: []string{"ExcessAttr1", "TfmExcessAttrs2", "MethodDecl"}}}},
		"Valid choice": {
			inputAssign: map[string]TypeData{"ValidChoice": p.InferredAssigns()["ValidChoice"]},
			inputLet:    map[string]TypeData{"ValidChoice": p.InferredLets()["ValidChoice"]},
			expected:    nil},
		"Relational Type": {
			inputAssign: map[string]TypeData{"Relational": p.InferredAssigns()["Relational"]},
			inputLet:    map[string]TypeData{"Relational": p.InferredLets()["Relational"]},
			expected:    nil},
		"Inner relational Type": {
			inputAssign: map[string]TypeData{"InnerRelational": p.InferredAssigns()["InnerRelational"]},
			inputLet:    map[string]TypeData{"InnerRelational": p.InferredLets()["InnerRelational"]},
			expected:    nil},
		"Transform variable valid": {
			inputAssign: map[string]TypeData{"TransformVarValid": p.InferredAssigns()["TransformVarValid"]},
			inputLet:    map[string]TypeData{"TransformVarValid": p.InferredLets()["TransformVarValid"]},
			expected:    nil},
		"Transform variable redefined": {
			inputAssign: map[string]TypeData{"TransformVarRedefined": p.InferredAssigns()["TransformVarRedefined"]},
			inputLet:    map[string]TypeData{"TransformVarRedefined": p.InferredLets()["TransformVarRedefined"]},
			expected:    []Msg{{MessageID: 409, MessageData: []string{"TransformVarRedefined", "varDeclaration"}}}},
		"Transform inner-variable redefined": {
			inputAssign: map[string]TypeData{"TransformInnerVarRedefined": p.InferredAssigns()["TransformInnerVarRedefined"]},
			inputLet:    map[string]TypeData{"TransformInnerVarRedefined": p.InferredLets()["TransformInnerVarRedefined"]},
			expected:    []Msg{{MessageID: 409, MessageData: []string{"TransformInnerVarRedefined:varDeclaration", "foo"}}}},
		"Transform assign redefined": {
			inputAssign: map[string]TypeData{"TransformAssignRedefined": p.InferredAssigns()["TransformAssignRedefined"]},
			inputLet:    map[string]TypeData{"TransformAssignRedefined": p.InferredLets()["TransformAssignRedefined"]},
			expected:    []Msg{{MessageID: 409, MessageData: []string{"TransformAssignRedefined", "VarDecl"}}}},
		"Transform inner-assign redefined": {
			inputAssign: map[string]TypeData{"TransformInnerAssignRedefined": p.InferredAssigns()["TransformInnerAssignRedefined"]},
			inputLet:    map[string]TypeData{"TransformInnerAssignRedefined": p.InferredLets()["TransformInnerAssignRedefined"]},
			expected:    []Msg{{MessageID: 409, MessageData: []string{"TransformInnerAssignRedefined:VarDecl", "TypeName"}}}},
		"Transform variable invalid": {
			inputAssign: map[string]TypeData{"TransformVarInvalid": p.InferredAssigns()["TransformVarInvalid"]},
			inputLet:    map[string]TypeData{"TransformVarInvalid": p.InferredLets()["TransformVarInvalid"]},
			expected: []Msg{
				{MessageID: 405, MessageData: []string{"identifier", "TransformVarInvalid:varDeclaration", "VarDecl"}},
				{MessageID: 406, MessageData: []string{"foo", "TransformVarInvalid:varDeclaration", "VarDecl"}}}},
	}

	for name, test := range cases {
		inputAssign := test.inputAssign
		inputLet := test.inputLet
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(grammar, transform, inputAssign, inputLet)
			validator.validateViews()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidateViewsInnerTypes(t *testing.T) {
	p := NewParser()
	transform, _ := loadTransform("tests", "transform1.sysl", p)
	grammar, _ := loadGrammar("tests/grammar.sysl")

	cases := map[string]struct {
		inputAssign map[string]TypeData
		inputLet    map[string]TypeData
		expected    []Msg
	}{
		"Valid inner type": {
			inputAssign: map[string]TypeData{"ValidInnerAttrs": p.InferredAssigns()["ValidInnerAttrs"]},
			inputLet:    map[string]TypeData{"ValidInnerAttrs": p.InferredLets()["ValidInnerAttrs"]},
			expected:    nil},
		"Invalid inner type": {
			inputAssign: map[string]TypeData{"InvalidInnerAttrs": p.InferredAssigns()["InvalidInnerAttrs"]},
			inputLet:    map[string]TypeData{"InvalidInnerAttrs": p.InferredLets()["InvalidInnerAttrs"]},
			expected: []Msg{
				{MessageID: ErrMissingReqField, MessageData: []string{"PackageName", "InvalidInnerAttrs", "PackageClause"}},
				{MessageID: ErrExcessAttr, MessageData: []string{"Foo", "InvalidInnerAttrs", "PackageClause"}}}},
	}
	for name, test := range cases {
		inputAssign := test.inputAssign
		inputLet := test.inputLet
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(grammar, transform, inputAssign, inputLet)
			validator.validateViews()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidateViewsChoiceTypes(t *testing.T) {
	p := NewParser()
	transform, _ := loadTransform("tests", "transform1.sysl", p)
	grammar, _ := loadGrammar("tests/grammar.sysl")

	cases := map[string]struct {
		inputAssign map[string]TypeData
		inputLet    map[string]TypeData
		expected    []Msg
	}{
		"Valid choice": {
			inputAssign: map[string]TypeData{"ValidChoice": p.InferredAssigns()["ValidChoice"]},
			inputLet:    map[string]TypeData{"ValidChoice": p.InferredLets()["ValidChoice"]},
			expected:    nil},
		"Invalid choice": {
			inputAssign: map[string]TypeData{"InvalidChoice": p.InferredAssigns()["InvalidChoice"]},
			inputLet:    map[string]TypeData{"InvalidChoice": p.InferredLets()["InvalidChoice"]},
			expected: []Msg{
				{MessageID: ErrInvalidOption, MessageData: []string{"InvalidChoice", "Foo", "Statement"}},
				{MessageID: ErrExcessAttr, MessageData: []string{"Foo", "InvalidChoice", "Statement"}}}},
		"Valid choice combination": {
			inputAssign: map[string]TypeData{"ValidChoiceCombination": p.InferredAssigns()["ValidChoiceCombination"]},
			inputLet:    map[string]TypeData{"ValidChoiceCombination": p.InferredLets()["ValidChoiceCombination"]},
			expected:    nil},
		"Valid choice non-combination": {
			inputAssign: map[string]TypeData{"ValidChoiceNonCombination": p.InferredAssigns()["ValidChoiceNonCombination"]},
			inputLet:    map[string]TypeData{"ValidChoiceNonCombination": p.InferredLets()["ValidChoiceNonCombination"]},
			expected:    nil},
		"Invalid choice combination excess": {
			inputAssign: map[string]TypeData{"InvalidChoiceCombinationExcess": p.InferredAssigns()["InvalidChoiceCombinationExcess"]},
			inputLet:    map[string]TypeData{"InvalidChoiceCombinationExcess": p.InferredLets()["InvalidChoiceCombinationExcess"]},
			expected: []Msg{{
				MessageID:   ErrExcessAttr,
				MessageData: []string{"Foo", "InvalidChoiceCombinationExcess", "MethodSpec"}}}},
		"Invalid choice combination missing": {
			inputAssign: map[string]TypeData{"InvalidChoiceCombiMissing": p.InferredAssigns()["InvalidChoiceCombiMissing"]},
			inputLet:    map[string]TypeData{"InvalidChoiceCombiMissing": p.InferredLets()["InvalidChoiceCombiMissing"]},
			expected: []Msg{
				{MessageID: ErrMissingReqField, MessageData: []string{"Signature", "InvalidChoiceCombiMissing", "MethodSpec"}},
				{MessageID: ErrExcessAttr, MessageData: []string{"Foo", "InvalidChoiceCombiMissing", "MethodSpec"}}}},
		"Invalid choice non-combination missing": {
			inputAssign: map[string]TypeData{"InvalidChoiceNonCombination": p.InferredAssigns()["InvalidChoiceNonCombination"]},
			inputLet:    map[string]TypeData{"InvalidChoiceNonCombination": p.InferredLets()["InvalidChoiceNonCombination"]},
			expected: []Msg{
				{
					MessageID:   ErrInvalidOption,
					MessageData: []string{"InvalidChoiceNonCombination", "Interface", "MethodSpec"}},
				{
					MessageID:   ErrExcessAttr,
					MessageData: []string{"Interface", "InvalidChoiceNonCombination", "MethodSpec"}}}},
	}
	for name, test := range cases {
		inputAssign := test.inputAssign
		inputLet := test.inputLet
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(grammar, transform, inputAssign, inputLet)
			validator.validateViews()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidate(t *testing.T) {
	p := NewParser()
	transform, _ := loadTransform("tests", "transform2.sysl", p)
	grammar, _ := loadGrammar("tests/grammar.sysl")

	validator := NewValidator(grammar, transform, p.InferredAssigns(), p.InferredLets())
	validator.Validate("goFile")
	actual := validator.GetMessages()
	assert.Nil(t, actual, "Unexpected result")
}

func TestValidatorLoadTransformSuccess(t *testing.T) {
	p := NewParser()
	tfm, err := loadTransform("tests", "transform2.sysl", p)
	assert.NotNil(t, tfm, "Unexpected result")
	assert.Nil(t, err, "Unexpected result")
}

func TestValidatorLoadTransformError(t *testing.T) {
	p := NewParser()
	tfm, err := loadTransform("foo", "bar.sysl", p)
	assert.Nil(t, tfm, "Unexpected result")
	assert.NotNil(t, err, "Unexpected result")
}

func TestValidatorLoadGrammarSuccess(t *testing.T) {
	grammar, err := loadGrammar("tests/grammar.sysl")
	assert.NotNil(t, grammar, "Unexpected result")
	assert.Nil(t, err, "Unexpected result")
}

func TestValidatorLoadGrammarError(t *testing.T) {
	grammar, err := loadGrammar("foo/bar.g")
	assert.Nil(t, grammar, "Unexpected result")
	assert.NotNil(t, err, "Unexpected result")
}

func TestValidatorDoValidate(t *testing.T) {
	cases := map[string]struct {
		args     []string
		flags    *flag.FlagSet
		isErrNil bool
	}{
		"Success": {
			args: []string{
				"sysl2", "validate", "-root-transform", "tests", "-transform", "transform2.sysl", "-grammar",
				"tests/grammar.sysl", "-start", "goFile"},
			flags: flag.NewFlagSet("validate", flag.PanicOnError), isErrNil: true},
		"Flag parse fail": {
			args: []string{
				"sysl2", "validate", "-root-transforms", "tests", "-transform", "transform2.sysl", "-grammar",
				"tests/grammar.sysl", "-start", "goFile"},
			flags: flag.NewFlagSet("validate", flag.ContinueOnError), isErrNil: false},
		"Grammar loading fail": {
			args: []string{
				"sysl2", "validate", "-root-transform", "tests", "-transform", "transform2.sysl", "-grammar",
				"tests/go.sysl", "-start", "goFile"},
			flags: flag.NewFlagSet("validate", flag.PanicOnError), isErrNil: false},
		"Transform loading fail": {
			args: []string{
				"sysl2", "validate", "-root-transform", "tests", "-transform", "tfm.sysl", "-grammar",
				"tests/grammar.sysl", "-start", "goFile"},
			flags: flag.NewFlagSet("validate", flag.PanicOnError), isErrNil: false},
		"Has validation messages": {
			args: []string{
				"sysl2", "validate", "-root-transform", "tests", "-transform", "transform1.sysl", "-grammar",
				"tests/grammar.sysl", "-start", "goFile"},
			flags: flag.NewFlagSet("validate", flag.PanicOnError), isErrNil: false},
	}

	for name, test := range cases {
		args := test.args
		flags := test.flags
		isErrNil := test.isErrNil
		t.Run(name, func(t *testing.T) {
			err := DoValidate(flags, args)
			if isErrNil {
				assert.Nil(t, err, "Unexpected result")
			} else {
				assert.NotNil(t, err, "Unexpected result")
			}
		})
	}
}
