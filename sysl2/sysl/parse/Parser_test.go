package parse

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	sysl "github.com/anz-bank/sysl/src/proto"
	"github.com/anz-bank/sysl/sysl2/sysl/pbutil"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readSyslModule(filename string) (*sysl.Module, error) {
	var buf bytes.Buffer

	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Open file %#v", filename)
	}
	if _, err := io.Copy(&buf, f); err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	module := &sysl.Module{}
	if err := proto.UnmarshalText(buf.String(), module); err != nil {
		return nil, errors.Wrapf(err, "Unmarshal proto: %s", filename)
	}
	return module, nil
}

func retainOrRemove(err error, file *os.File, retainOnError bool) error {
	if err != nil {
		if retainOnError {
			fmt.Printf("%#v retained for checking\n", file.Name())
		} else {
			os.Remove(file.Name())
		}
	}
	return err
}

func parseComparable(
	filename, root string,
	stripSourceContext bool,
) (*sysl.Module, error) {
	module, err := Parse(filename, root)
	if err != nil {
		return nil, err
	}

	if stripSourceContext {
		// remove stuff that does not match legacy.
		for _, app := range module.Apps {
			app.SourceContext = nil
			for _, ep := range app.Endpoints {
				ep.SourceContext = nil
			}
			for _, t := range app.Types {
				t.SourceContext = nil
			}
			for _, t := range app.Views {
				t.SourceContext = nil
			}
		}
	}

	return module, nil
}

func parseAndCompare(
	filename, root, golden string,
	goldenProto proto.Message,
	retainOnError bool,
	stripSourceContext bool,
) (bool, error) {
	module, err := parseComparable(filename, root, stripSourceContext)
	if err != nil {
		return false, err
	}

	if proto.Equal(goldenProto, module) {
		return true, nil
	}

	if err = pbutil.TextPB(goldenProto, golden); err != nil {
		return false, err
	}

	pattern := "sysl-test-*.textpb"
	generated, err := ioutil.TempFile("", pattern)
	if err != nil {
		return false, errors.Wrapf(err, "Create tempfile: %s", pattern)
	}
	generatedClosed := false
	defer func() {
		if !generatedClosed {
			generated.Close()
		}
	}()

	if err = retainOrRemove(pbutil.FTextPB(generated, module), generated, retainOnError); err != nil {
		return false, errors.Wrapf(err, "Generate %#v", generated)
	}
	if err := generated.Close(); err != nil {
		return false, errors.Wrapf(err, "Close %#v", generated)
	}
	generatedClosed = true

	cmd := exec.Command("diff", "-y", golden, generated.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if retainOnError {
			fmt.Printf("%#v retained for checking\n", generated.Name())
		} else {
			os.Remove(generated.Name())
		}
		return false, errors.Wrapf(err, "diff -y %#v %#v", golden, generated.Name())
	}

	os.Remove(generated.Name())
	return false, nil
}

func parseAndCompareWithGolden(filename, root string, stripSourceContext bool) (bool, error) {
	goldenFilename := filename
	if !strings.HasSuffix(goldenFilename, syslExt) {
		goldenFilename += syslExt
	}
	goldenFilename += ".golden.textpb"
	golden := path.Join(root, goldenFilename)

	goldenModule, err := readSyslModule(golden)
	if err != nil {
		return false, err
	}
	return parseAndCompare(filename, root, golden, goldenModule, true, stripSourceContext)
}

func testParseAgainstGolden(t *testing.T, filename, root string) {
	equal, err := parseAndCompareWithGolden(filename, root, true)
	if assert.NoError(t, err) {
		assert.True(t, equal, "%#v %#v", root, filename)
	}
}

func testParseAgainstGoldenWithSourceContext(t *testing.T, filename string) {
	equal, err := parseAndCompareWithGolden(filename, "", false)
	if assert.NoError(t, err) {
		assert.True(t, equal, "%#v", filename)
	}
}

func TestParseBadRoot(t *testing.T) {
	_, err := parseComparable("dontcare.sysl", "NON-EXISTENT-ROOT", false)
	assert.Error(t, err)
}

func TestParseMissingFile(t *testing.T) {
	_, err := parseComparable("doesn't.exist.sysl", "tests", false)
	assert.Error(t, err)
}

func TestParseDirectoryAsFile(t *testing.T) {
	dirname := "not-a-file.sysl"
	tmproot := os.TempDir()
	tmpdir := path.Join(tmproot, dirname)
	require.NoError(t, os.Mkdir(tmpdir, 0755))
	defer os.Remove(tmpdir)
	_, err := parseComparable(dirname, tmproot, false)
	assert.Error(t, err)
}

func TestParseBadFile(t *testing.T) {
	_, err := parseAndCompareWithGolden("sysl.go", "", false)
	assert.Error(t, err)
}

func TestSimpleEP(t *testing.T) {
	testParseAgainstGolden(t, "tests/test1.sysl", "")
}

func TestSimpleEPNoSuffix(t *testing.T) {
	testParseAgainstGolden(t, "tests/test1", "")
}

func TestAttribs(t *testing.T) {
	testParseAgainstGolden(t, "tests/attribs.sysl", "")
}

func TestIfElse(t *testing.T) {
	testParseAgainstGolden(t, "tests/if_else.sysl", "")
}

func TestArgs(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/args.sysl")
}

func TestSimpleEPWithSpaces(t *testing.T) {
	testParseAgainstGolden(t, "tests/with_spaces.sysl", "")
}

func TestSimpleEP2(t *testing.T) {
	testParseAgainstGolden(t, "tests/test4.sysl", "")
}

func TestUnion(t *testing.T) {
	testParseAgainstGolden(t, "tests/union.sysl", "")
}

func TestSimpleEndpointParams(t *testing.T) {
	testParseAgainstGolden(t, "tests/ep_params.sysl", "")
}

func TestOneOfStatements(t *testing.T) {
	testParseAgainstGolden(t, "tests/oneof.sysl", "")
}

func TestDuplicateEndpoints(t *testing.T) {
	testParseAgainstGolden(t, "tests/duplicate.sysl", "")
}

func TestEventing(t *testing.T) {
	testParseAgainstGolden(t, "tests/eventing.sysl", "")
}

func TestCollector(t *testing.T) {
	testParseAgainstGolden(t, "tests/collector.sysl", "")
}

func TestPubSubCollector(t *testing.T) {
	testParseAgainstGolden(t, "tests/pubsub_collector.sysl", "")
}

func TestDocstrings(t *testing.T) {
	testParseAgainstGolden(t, "tests/docstrings.sysl", "")
}

func TestMixins(t *testing.T) {
	testParseAgainstGolden(t, "tests/mixin.sysl", "")
}
func TestForLoops(t *testing.T) {
	testParseAgainstGolden(t, "tests/for_loop.sysl", "")
}

func TestGroupStmt(t *testing.T) {
	testParseAgainstGolden(t, "tests/group_stmt.sysl", "")
}

func TestUntilLoop(t *testing.T) {
	testParseAgainstGolden(t, "tests/until_loop.sysl", "")
}

func TestTuple(t *testing.T) {
	testParseAgainstGolden(t, "tests/test2.sysl", "")
}

func TestInplaceTuple(t *testing.T) {
	testParseAgainstGolden(t, "tests/inplace_tuple.sysl", "")
}

func TestRelational(t *testing.T) {
	testParseAgainstGolden(t, "tests/school.sysl", "")
}

func TestImports(t *testing.T) {
	testParseAgainstGolden(t, "tests/library.sysl", "")
}

func TestRootArg(t *testing.T) {
	testParseAgainstGolden(t, "school.sysl", "tests")
}

func TestSequenceType(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/sequence_type.sysl")
}

func TestRestApi(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/test_rest_api.sysl")
}

func TestRestApiQueryParams(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/rest_api_query_params.sysl")
}

func TestSimpleProject(t *testing.T) {
	testParseAgainstGolden(t, "tests/project.sysl", "")
}

func TestUrlParamOrder(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/rest_url_params.sysl")
}

func TestRestApi_WrongOrder(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/bad_order.sysl")
}

func TestTransform(t *testing.T) {
	testParseAgainstGolden(t, "tests/transform.sysl", "")
}

func TestImpliedDot(t *testing.T) {
	testParseAgainstGolden(t, "tests/implied.sysl", "")
}

func TestStmts(t *testing.T) {
	testParseAgainstGolden(t, "tests/stmts.sysl", "")
}

func TestMath(t *testing.T) {
	testParseAgainstGolden(t, "tests/math.sysl", "")
}

func TestTableof(t *testing.T) {
	testParseAgainstGolden(t, "tests/tableof.sysl", "")
}

func TestRank(t *testing.T) {
	testParseAgainstGolden(t, "tests/rank.sysl", "")
}

func TestMatching(t *testing.T) {
	testParseAgainstGolden(t, "tests/matching.sysl", "")
}

func TestNavigate(t *testing.T) {
	testParseAgainstGolden(t, "tests/navigate.sysl", "")
}

func TestFuncs(t *testing.T) {
	testParseAgainstGolden(t, "tests/funcs.sysl", "")
}

func TestPetshop(t *testing.T) {
	testParseAgainstGolden(t, "tests/petshop.sysl", "")
}

func TestCrash(t *testing.T) {
	testParseAgainstGolden(t, "tests/crash.sysl", "")
}

func TestStrings(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/strings_expr.sysl")
}

func TestTypeAlias(t *testing.T) {
	testParseAgainstGoldenWithSourceContext(t, "tests/alias.sysl")
}
