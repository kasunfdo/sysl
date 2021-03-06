package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/anz-bank/sysl/sysl2/sysl/parse"
	"github.com/anz-bank/sysl/sysl2/sysl/testutil"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMain2(t *testing.T, args []string, golden string) {
	if output := testutil.TempFilename(t, "", "sysl-TestTextPBNilModule-*.textpb"); output != "" {
		rc := main2(append([]string{"sysl", "pb", "-o", output}, args...), main3)
		assert.Zero(t, rc)

		actual, err := ioutil.ReadFile(output)
		assert.NoError(t, err)

		expected, err := ioutil.ReadFile(golden)
		assert.NoError(t, err)

		assert.Equal(t, string(expected), string(actual))
	}
}

func TestMain2TextPB(t *testing.T) {
	testMain2(t, []string{"tests/args.sysl"}, "tests/args.sysl.golden.textpb")
}

func TestMain2JSON(t *testing.T) {
	testMain2(t, []string{"--mode", "json", "tests/args.sysl"}, "tests/args.sysl.golden.json")
}

func testMain2Stdout(t *testing.T, args []string, golden string) {
	rc := main2(append([]string{"sysl", "pb", "-o", " - "}, args...), main3)
	assert.Zero(t, rc)

	_, err := ioutil.ReadFile(golden)
	require.NoError(t, err)

	_, err = os.Stat("-")
	assert.True(t, os.IsNotExist(err))
}

func TestMain2TextPBStdout(t *testing.T) {
	testMain2Stdout(t, []string{"tests/args.sysl"}, "tests/args.sysl.golden.textpb")
}

func TestMain2JSONStdout(t *testing.T) {
	testMain2Stdout(t, []string{"--mode", "json", "tests/args.sysl"}, "tests/args.sysl.golden.json")
}

func TestMain2BadMode(t *testing.T) {
	rc := main2([]string{"sysl", "pb", "-o", " - ", "--mode", "BAD", "tests/args.sysl"}, main3)
	assert.NotZero(t, rc)

	_, err := os.Stat("-")
	assert.True(t, os.IsNotExist(err))
}

func TestMain2BadLog(t *testing.T) {
	rc := main2([]string{"sysl", "pb", "-o", "-", "--log", "BAD", "tests/args.sysl"}, main3)
	assert.NotZero(t, rc)

	_, err := os.Stat("-")
	assert.True(t, os.IsNotExist(err))
}

func TestMain2SeqdiagWithMissingFile(t *testing.T) {
	testHook := test.NewGlobal()
	defer testHook.Reset()
	rc := main2([]string{"sd", "-o", "%(epname).png", "tests/MISSING.sysl", "-a", "Project :: Sequences"}, main3)
	assert.NotEqual(t, 0, rc)
}

func TestMain2SeqdiagWithImpossibleOutput(t *testing.T) {
	testHook := test.NewGlobal()
	defer testHook.Reset()
	rc := main2([]string{"sd", "-s", "MobileApp <- Login", "-o", "tests/call.zzz", "-b", "Server <- DB=call to database",
		"-b", "Server <- Login=call to database", "tests/call.sysl"}, main3)
	assert.NotEqual(t, 0, rc)
}

func TestMain2WithBlackboxParams(t *testing.T) {
	testHook := test.NewGlobal()
	main2([]string{"sysl", "sd", "-s", "MobileApp <- Login", "-o", "tests/call.png", "-b", "Server <- DB=call to database",
		"-b", "Server <- Login=call to database", "tests/call.sysl"}, main3)
	assert.Equal(t, logrus.WarnLevel, testHook.LastEntry().Level)
	assert.Equal(t, "blackbox 'Server <- DB' passed on commandline not hit\n", testHook.LastEntry().Message)
	testHook.Reset()
}

func TestMain2WithBlackboxParamsFaultyArguments(t *testing.T) {
	testHook := test.NewGlobal()
	main2([]string{"sysl", "sd", "-s", "MobileApp <- Login", "-o", "tests/call.png", "-b", "Server <- DB",
		"-b", "Server <- Login", "tests/call.sysl"}, main3)
	assert.Equal(t, logrus.ErrorLevel, testHook.LastEntry().Level)
	assert.Equal(t, "expected KEY=VALUE got 'Server <- DB'", testHook.LastEntry().Message)
	testHook.Reset()
}

func TestMain2WithBlackboxSysl(t *testing.T) {
	testHook := test.NewGlobal()
	main2([]string{"sysl", "sd", "-o", "%(epname).png", "tests/blackbox.sysl", "-a", "Project :: Sequences"}, main3)
	assert.Equal(t, logrus.WarnLevel, testHook.LastEntry().Level)
	assert.Equal(t, "blackbox 'SomeApp <- AppEndpoint' not hit in app 'Project :: Sequences'\n",
		testHook.Entries[len(testHook.Entries)-1].Message)
	assert.Equal(t, "blackbox 'SomeApp <- BarEndpoint1' not hit in app 'Project :: Sequences :: SEQ-Two'\n",
		testHook.Entries[len(testHook.Entries)-2].Message)
	assert.Equal(t, "blackbox 'SomeApp <- BarEndpoint' not hit in app 'Project :: Sequences :: SEQ-One'\n",
		testHook.Entries[len(testHook.Entries)-3].Message)
	testHook.Reset()
}

func TestMain2WithBlackboxSyslEmptyEndpoints(t *testing.T) {
	testHook := test.NewGlobal()
	main2([]string{"sysl", "sd", "-o", "%(epname).png", "tests/blackbox.sysl", "-a", "Project :: Integrations"}, main3)
	assert.Equal(t, logrus.ErrorLevel, testHook.LastEntry().Level)
	assert.Equal(t, "No call statements to build sequence diagram for endpoint PROJECT-E2E", testHook.LastEntry().Message)
	testHook.Reset()
}

func TestMain2Fatal(t *testing.T) {
	testHook := test.NewGlobal()
	assert.Equal(t, 42, main2(nil, func(_ []string) error {
		return parse.Exitf(42, "Exit error")
	}))
	assert.Equal(t, logrus.ErrorLevel, testHook.LastEntry().Level)
	testHook.Reset()
}

func TestMain2WithGroupingParamsGroupParamAbsent(t *testing.T) {
	testHook := test.NewGlobal()
	main2([]string{"sysl", "sd", "-s", "MobileApp <- Login", "-g", "-o", "tests/call.png", "tests/call.sysl"}, main3)
	assert.Equal(t, logrus.ErrorLevel, testHook.LastEntry().Level)
	assert.Equal(t, "expected argument for flag '-g'", testHook.LastEntry().Message)
	testHook.Reset()
}

func TestMain2WithGroupingParamsCommandline(t *testing.T) {
	main2([]string{"sysl", "sd", "-s", "MobileApp <- Login", "-g", "owner", "-o", "tests/call.png",
		"tests/call.sysl"}, main3)
	_, err := os.Stat("tests/call.png")
	assert.NoError(t, err)
	os.Remove("tests/call.png")
}

func TestMain2WithGroupingParamsSysl(t *testing.T) {
	testHook := test.NewGlobal()
	main2([]string{"sysl", "sd", "-g", "location", "-o", "%(epname).png", "tests/groupby.sysl", "-a",
		"Project :: Sequences"}, main3)
	for _, filename := range []string{"SEQ-One.png", "SEQ-Two.png"} {
		_, err := os.Stat(filename)
		assert.NoError(t, err)
		os.Remove(filename)
	}
	assert.Equal(t, logrus.WarnLevel, testHook.LastEntry().Level)
	assert.Equal(t, "Ignoring groupby passed from command line", testHook.LastEntry().Message)
}

func TestMain2WithGenerateIntegrations(t *testing.T) {
	out := "indirect_1.png"
	main2([]string{"sysl", "ints", "--root", "./tests/", "-o", out, "-j", "Project", "indirect_1.sysl"}, main3)
	_, err2 := os.Stat(out)
	assert.True(t, err2 == nil)
	os.Remove(out)
}

func TestMain2WithGenerateCode(t *testing.T) {
	ret := main2([]string{"sysl", "gen", "--root-model", ".", "--model", "tests/model.sysl",
		"--root-transform", ".", "--transform", "tests/test.gen_multiple_annotations.sysl",
		"--grammar", "tests/test.gen.g", "--start", "javaFile"}, main3)
	assert.True(t, ret == 0)
	os.Remove("Model.java")
}

func TestMain2WithTestPbJsonMode(t *testing.T) {
	out := "tests/callout"
	ret := main2([]string{"sysl", "pb", "--mode", "textpb", "-o", out, "tests/call.sysl"}, main3)
	assert.True(t, ret == 0)
	os.Remove(out)
}

func TestMain2WithTestPbMode(t *testing.T) {
	out := "tests/callout"
	ret := main2([]string{"sysl", "pb", "--mode", "json", "-o", out, "tests/call.sysl"}, main3)
	assert.True(t, ret == 0)
	os.Remove(out)
}

func TestMain2WithTestPbJsonConsole(t *testing.T) {
	ret := main2([]string{"sysl", "pb", "--mode", "textpb", "-o", " - ", "tests/call.sysl", "-v"}, main3)
	assert.True(t, ret == 0)
}

func TestMain2WithTestPbConsole(t *testing.T) {
	ret := main2([]string{"sysl", "pb", "--mode", "json", "-o", " - ", "tests/call.sysl", "-v"}, main3)
	assert.True(t, ret == 0)
}
