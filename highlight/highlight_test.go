package highlight_test

import (
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "bitbucket.org/johnsto/httpoo/highlight"
)

func TestHighlightHTML(t *testing.T) {
	r := strings.NewReader(`<!doctype html>
	<html>
		<head><title>What</title></head>
		<body class="content">
			You <!-- don't --> smell <b style="color: #f00;">nice</b> today!
		</body>
	</html>`)

	tokens, err := HTML.Tokenise(r)
	assert.Nil(t, err)
	for _, token := range tokens {
		log.Printf("%s - %s", string(token.Value), token.Type)
	}
}

func TestHighlightJSON(t *testing.T) {
	r := strings.NewReader("{\n  \"key\": \"value\"\n}")

	tokens, err := JSON.Tokenise(r)
	assert.Nil(t, err)
	for _, token := range tokens {
		log.Printf("%s - %s", string(token.Value), token.Type)
	}
}
