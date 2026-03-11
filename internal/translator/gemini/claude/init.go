package claude

import (
	. "github.com/lionpizzicato/CLIProxyAPI/v6/internal/constant"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/interfaces"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		Claude,
		Gemini,
		ConvertClaudeRequestToGemini,
		interfaces.TranslateResponse{
			Stream:     ConvertGeminiResponseToClaude,
			NonStream:  ConvertGeminiResponseToClaudeNonStream,
			TokenCount: ClaudeTokenCount,
		},
	)
}
