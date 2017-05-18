package scale

func FormatManifest(output []*OutputData, parsed []*ParseResult) *ResultsManifest {
	if parsed != nil {
		return &ResultsManifest{
			Version:      "1.1",
			OutputData:   output,
			ParseResults: parsed,
		}
	} else {
		return &ResultsManifest{
			Version:    "1.1",
			OutputData: output,
		}
	}
}
