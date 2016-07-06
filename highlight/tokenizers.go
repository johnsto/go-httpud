package highlight

var tokenizers map[string]Tokenizer

func Register(name string, t Tokenizer) {
	if tokenizers == nil {
		tokenizers = map[string]Tokenizer{}
	}
	tokenizers[name] = t
}

func GetTokenizer(name string) Tokenizer {
	return tokenizers[name]
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

// GetTokenizerForFilename returns a Tokenizer for the given filename
// (e.g. "index.html" or "jasons.json"), or nil if one is not found.
func GetTokenizerForFilename(name string) (Tokenizer, error) {
	for _, tokenizer := range tokenizers {
		if matched, err := tokenizer.AcceptsFilename(name); err != nil {
			return nil, err
		} else if matched {
			return tokenizer, nil
		}
	}
	return nil, nil
}
