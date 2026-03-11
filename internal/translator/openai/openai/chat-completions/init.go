package chat_completions

import (
	. "github.com/lionpizzicato/CLIProxyAPI/v6/internal/constant"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/interfaces"
	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		OpenAI,
		OpenAI,
		ConvertOpenAIRequestToOpenAI,
		interfaces.TranslateResponse{
			Stream:    ConvertOpenAIResponseToOpenAI,
			NonStream: ConvertOpenAIResponseToOpenAINonStream,
		},
	)
}
