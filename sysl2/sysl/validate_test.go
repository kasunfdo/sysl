package main

import (
	"flag"
	"testing"

	sysl "github.com/anz-bank/sysl/src/proto"
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
	transform, _ := loadAndGetDefaultApp("tests", "transform1.sysl")

	var entryPointView *sysl.View
	var nonEntryPointView *sysl.View
	var invalidEntryPointView *sysl.View

	for _, tfm := range transform.GetApps() {
		entryPointView = tfm.Views[start]
		nonEntryPointView = tfm.Views["TfmDefaultEmpty"]
		invalidEntryPointView = tfm.Views["EntryPointInvalid"]
	}

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
			validator := NewValidator(nil, input)
			validator.validateEntryPoint(start)
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidateFileName(t *testing.T) {
	viewName := "filename"
	transform, _ := loadAndGetDefaultApp("tests", "transform1.sysl")

	var fileNameView *sysl.View
	var nonFileNameView *sysl.View
	var invalidFileNameView1 *sysl.View
	var invalidFileNameView2 *sysl.View
	var invalidFileNameView3 *sysl.View

	for _, tfm := range transform.GetApps() {
		fileNameView = tfm.Views[viewName]
		nonFileNameView = tfm.Views["TfmDefaultEmpty"]
		invalidFileNameView1 = tfm.Views["TfmFilenameInvalid1"]
		invalidFileNameView2 = tfm.Views["TfmFilenameInvalid2"]
		invalidFileNameView3 = tfm.Views["TfmFilenameInvalid3"]
	}

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
			validator := NewValidator(nil, input)
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
			syslType := resolver.resolveExprType(input.expr, input.viewName, "", "")
			messages := resolver.GetMessages()
			assert.True(t, hasSameType(expected.syslType, syslType), "Unexpected result")
			assert.Equal(t, expected.messages, messages, "Unexpected result")
		})
	}
}

func TestValidatorValidateViews(t *testing.T) {
	transformModule, tfmAppName := loadAndGetDefaultApp("tests", "transform1.sysl")
	grammarModule, grammarAppName := loadAndGetDefaultApp("tests", "grammar.sysl")

	grammar := grammarModule.GetApps()[grammarAppName]
	tfmViews := transformModule.GetApps()[tfmAppName].GetViews()

	cases := map[string]struct {
		input    map[string]*sysl.View
		expected []Msg
	}{
		"Equal": {
			input:    map[string]*sysl.View{"TfmValid": tfmViews["TfmValid"]},
			expected: nil},
		"Not Equal": {
			input: map[string]*sysl.View{"TfmInvalid": tfmViews["TfmInvalid"]},
			expected: []Msg{
				{MessageID: ErrMissingReqField, MessageData: []string{"FunctionName", "TfmInvalid", "MethodDecl"}}}},
		"Absent optional": {
			input:    map[string]*sysl.View{"TfmNoOptional": tfmViews["TfmNoOptional"]},
			expected: nil},
		"Excess attributes without optionals": {
			input: map[string]*sysl.View{"TfmExcessAttrs1": tfmViews["TfmExcessAttrs1"]},
			expected: []Msg{
				{MessageID: ErrExcessAttr, MessageData: []string{"ExcessAttr1", "TfmExcessAttrs1", "MethodDecl"}}}},
		"Excess attributes with optionals": {
			input: map[string]*sysl.View{"TfmExcessAttrs2": tfmViews["TfmExcessAttrs2"]},
			expected: []Msg{
				{MessageID: ErrExcessAttr, MessageData: []string{"ExcessAttr1", "TfmExcessAttrs2", "MethodDecl"}}}},
		"Valid choice": {
			input:    map[string]*sysl.View{"ValidChoice": tfmViews["ValidChoice"]},
			expected: nil},
		"Relational Type": {
			input:    map[string]*sysl.View{"Relational": tfmViews["Relational"]},
			expected: nil},
		"Inner relational Type": {
			input:    map[string]*sysl.View{"InnerRelational": tfmViews["InnerRelational"]},
			expected: nil},
		"Transform variable valid": {
			input:    map[string]*sysl.View{"TransformVarValid": tfmViews["TransformVarValid"]},
			expected: nil},
		"Transform variable invalid": {
			input: map[string]*sysl.View{"TransformVarInvalid": tfmViews["TransformVarInvalid"]},
			expected: []Msg{
				{MessageID: 405, MessageData: []string{"identifier", "TransformVarInvalid:varDeclaration", "VarDecl"}},
				{MessageID: 406, MessageData: []string{"foo", "TransformVarInvalid:varDeclaration", "VarDecl"}}}},
	}

	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(grammar, &sysl.Application{Views: input})
			validator.validateViews()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidateViewsInnerTypes(t *testing.T) {
	transformModule, tfmAppName := loadAndGetDefaultApp("tests", "transform1.sysl")
	grammarModule, grammarAppName := loadAndGetDefaultApp("tests", "grammar.sysl")

	grammar := grammarModule.GetApps()[grammarAppName]
	tfmViews := transformModule.GetApps()[tfmAppName].GetViews()

	cases := map[string]struct {
		input    map[string]*sysl.View
		expected []Msg
	}{
		"Valid inner type": {
			input:    map[string]*sysl.View{"ValidInnerAttrs": tfmViews["ValidInnerAttrs"]},
			expected: nil},
		"Invalid inner type": {
			input: map[string]*sysl.View{"InvalidInnerAttrs": tfmViews["InvalidInnerAttrs"]},
			expected: []Msg{
				{MessageID: ErrMissingReqField, MessageData: []string{"PackageName", "InvalidInnerAttrs", "PackageClause"}},
				{MessageID: ErrExcessAttr, MessageData: []string{"Foo", "InvalidInnerAttrs", "PackageClause"}}}},
	}
	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(grammar, &sysl.Application{Views: input})
			validator.validateViews()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidateViewsChoiceTypes(t *testing.T) {
	transformModule, tfmAppName := loadAndGetDefaultApp("tests", "transform1.sysl")
	grammarModule, grammarAppName := loadAndGetDefaultApp("tests", "grammar.sysl")

	grammar := grammarModule.GetApps()[grammarAppName]
	tfmViews := transformModule.GetApps()[tfmAppName].GetViews()

	cases := map[string]struct {
		input    map[string]*sysl.View
		expected []Msg
	}{
		"Valid choice": {
			input:    map[string]*sysl.View{"ValidChoice": tfmViews["ValidChoice"]},
			expected: nil},
		"Invalid choice": {
			input: map[string]*sysl.View{"InvalidChoice": tfmViews["InvalidChoice"]},
			expected: []Msg{
				{MessageID: ErrInvalidOption, MessageData: []string{"InvalidChoice", "Foo", "Statement"}},
				{MessageID: ErrExcessAttr, MessageData: []string{"Foo", "InvalidChoice", "Statement"}}}},
		"Valid choice combination": {
			input:    map[string]*sysl.View{"ValidChoiceCombination": tfmViews["ValidChoiceCombination"]},
			expected: nil},
		"Valid choice non-combination": {
			input:    map[string]*sysl.View{"ValidChoiceNonCombination": tfmViews["ValidChoiceNonCombination"]},
			expected: nil},
		"Invalid choice combination excess": {
			input: map[string]*sysl.View{"InvalidChoiceCombinationExcess": tfmViews["InvalidChoiceCombinationExcess"]},
			expected: []Msg{{
				MessageID:   ErrExcessAttr,
				MessageData: []string{"Foo", "InvalidChoiceCombinationExcess", "MethodSpec"}}}},
		"Invalid choice combination missing": {
			input: map[string]*sysl.View{"InvalidChoiceCombiMissing": tfmViews["InvalidChoiceCombiMissing"]},
			expected: []Msg{
				{MessageID: ErrMissingReqField, MessageData: []string{"Signature", "InvalidChoiceCombiMissing", "MethodSpec"}},
				{MessageID: ErrExcessAttr, MessageData: []string{"Foo", "InvalidChoiceCombiMissing", "MethodSpec"}}}},
		"Invalid choice non-combination missing": {
			input: map[string]*sysl.View{"InvalidChoiceNonCombination": tfmViews["InvalidChoiceNonCombination"]},
			expected: []Msg{
				{
					MessageID:   ErrInvalidOption,
					MessageData: []string{"InvalidChoiceNonCombination", "Interface", "MethodSpec"}},
				{
					MessageID:   ErrExcessAttr,
					MessageData: []string{"Interface", "InvalidChoiceNonCombination", "MethodSpec"}}}},
	}
	for name, test := range cases {
		input := test.input
		expected := test.expected
		t.Run(name, func(t *testing.T) {
			validator := NewValidator(grammar, &sysl.Application{Views: input})
			validator.validateViews()
			actual := validator.GetMessages()
			assert.Equal(t, expected, actual, "Unexpected result")
		})
	}
}

func TestValidatorValidate(t *testing.T) {
	transformModule, tfmAppName := loadAndGetDefaultApp("tests", "transform2.sysl")
	grammarModule, grammarAppName := loadAndGetDefaultApp("tests", "grammar.sysl")

	grammar := grammarModule.GetApps()[grammarAppName]
	transform := transformModule.GetApps()[tfmAppName]
	validator := NewValidator(grammar, transform)
	validator.Validate("goFile")
	actual := validator.GetMessages()
	assert.Nil(t, actual, "Unexpected result")
}

func TestValidatorLoadTransformSuccess(t *testing.T) {
	tfm, err := loadTransform("tests", "transform2.sysl")
	assert.NotNil(t, tfm, "Unexpected result")
	assert.Nil(t, err, "Unexpected result")
}

func TestValidatorLoadTransformError(t *testing.T) {
	tfm, err := loadTransform("foo", "bar.sysl")
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
