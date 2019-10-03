package validate

import (
	"fmt"
	"testing"

	"gotest.tools/assert"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	parser "github.com/anz-bank/sysl/sysl2/naive"
	"github.com/anz-bank/sysl/sysl2/sysl/parse"
	"github.com/anz-bank/sysl/sysl2/sysl/syslutil"
)

func TestGenerateSyslGrammar(t *testing.T) {
	fs := afero.NewOsFs()
	dat, err := afero.ReadFile(fs, "tests/simpleGrammar.g")
	require.NoError(t, err)

	generated := buildSyslGrammar(parser.ParseEBNF(string(dat), "gen", "startRule"))
	syslgrammar, name, err := parse.LoadAndGetDefaultApp(
		"simpleGrammar.sysl",
		syslutil.NewChrootFs(fs, "tests"),
		parse.NewParser())
	syslgrammar.GetApps()[name].Types["startRule"].SourceContext = nil
	fmt.Println(generated.Types["startRule"])
	fmt.Println(syslgrammar.GetApps()[name].Types["startRule"])

	assert.Equal(t, generated.Types["startRule"], syslgrammar.GetApps()[name].Types["startRule"])
}
