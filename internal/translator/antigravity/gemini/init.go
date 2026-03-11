package gemini

import (
	. "github.com/lionpizzicato/CLIProxyAPI/v6/internal/constant"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/interfaces"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		Gemini,
		Antigravity,
		ConvertGeminiRequestToAntigravity,
		interfaces.TranslateResponse{
			Stream:     ConvertAntigravityResponseToGemini,
			NonStream:  ConvertAntigravityResponseToGeminiNonStream,
			TokenCount: GeminiTokenCount,
		},
	)
}
