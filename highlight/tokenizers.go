package highlight

var tokenizers []Tokenizer

func Register(t Tokenizer) {
	tokenizers = append(tokenizers, t)
}

// GetTokenizerForContentType returns a Tokenizer for the given content type
// (e.g. "text/html" or "application/json"), or nil if one is not found.
func GetTokenizerForContentType(contentType string) (Tokenizer, error) {
	for _, tokenizer := range tokenizers {
		if matched, err := tokenizer.AcceptsMediaType(contentType); err != nil {
			return nil, err
		} else if matched {
			return tokenizer, nil
		}
	}
	return nil, nil
}