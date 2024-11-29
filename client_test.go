// Copyright (c) 2020, 2023, Geert JM Vanderkelen

package gogql

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/golistic/xgo/xt"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// MustMD5SumIt returns the hexadecimal string representation of the MD5sum of s.
func MustMD5SumIt(s string) string {
	h := md5.New()
	if _, err := io.WriteString(h, s); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

var responses = map[string]*Result{
	"9714361de333a06aa7ecfac29993eb8e": {
		Data: json.RawMessage(`
{
	"allFilms": {
		"films": [
			{
			  "title": "A New Hope"
			},
			{
			  "title": "The Empire Strikes Back"
			},
			{
			  "title": "Return of the Jedi"
			},
			{
			  "title": "The Phantom Menace"
			},
			{
			  "title": "Attack of the Clones"
			},
			{
			  "title": "Revenge of the Sith"
			},
			{
			  "title": "The Force Awakens"
			},
			{
			  "title": "The Last Jedi"
			},
			{
			  "title": "The Rise of Skywalker"
			}
		]
	}
}
`),
	},
}

func TestNewClient(t *testing.T) {
	t.Run("content type is by default application/json", func(t *testing.T) {
		c := NewClient("http://example.com/graphql") // URI does not matter for test
		xt.Eq(t, "application/json", c.ContentType)
	})
}

func TestClient_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handlerGraphQLSWAPI))
	defer func() { server.Close() }()

	t.Run("allFilms", func(t *testing.T) {
		q := `query {
  allFilms {
    films {
      title
    }
  }
}`
		c := NewClient(server.URL)

		var data struct {
			AllFilms struct {
				Films []struct {
					Title string `json:"title"`
				} `json:"films"`
			} `json:"allFilms"`
		}

		err := c.Execute(q, &data)
		xt.Assert(t, err == nil)

		var titles []string
		for _, f := range data.AllFilms.Films {
			titles = append(titles, f.Title)
		}
		sort.Strings(titles)
		exp := []string{
			"A New Hope", "Attack of the Clones", "Return of the Jedi",
			"Revenge of the Sith", "The Empire Strikes Back", "The Force Awakens",
			"The Last Jedi", "The Phantom Menace", "The Rise of Skywalker"}
		xt.Eq(t, exp, titles)
	})

	t.Run("error returned", func(t *testing.T) {
		c := NewClient(server.URL)

		var data map[string]any

		err := c.Execute(`{ notExistingQuery { id } }`, &data)
		xt.Assert(t, err != nil)
		xt.Eq(t, "input: test response not available", err[0].Error())
	})
}

func handlerGraphQLSWAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.WriteHeader(http.StatusOK)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		err.Error()
		_, _ = w.Write([]byte("failed reading body"))
	}

	var payload Payload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		_, _ = w.Write([]byte("failed writing body"))
	}

	res, ok := responses[MustMD5SumIt(payload.Query)]
	if !ok {
		res = &Result{}
		res.Errors = []gqlerror.Error{
			{
				Message: "test response not available",
			},
		}
	}

	buf, err := json.Marshal(res)
	if err != nil {
		_, _ = w.Write([]byte("failed writing body"))
	}

	_, err = w.Write(buf)
	if err != nil {
		_, _ = w.Write([]byte("failed writing body"))
	}
}
