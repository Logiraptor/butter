package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"appengine"

	"code.google.com/p/go.net/context"
	"github.com/Logiraptor/butter/keys"

	"appengine/datastore"

	"github.com/Logiraptor/butter/db"

	"appengine/aetest"

	"testing"
)

// NOTE: this test fails due to NewContext being called with an invalid request.
func TestPaged(t *testing.T) {
	type E struct {
		ID     int64
		Parent *datastore.Key
		Name   string
	}
	aectx, _ := aetest.NewContext(nil)
	ctx := keys.WithAEContext(context.Background(), aectx)
	k := datastore.NewKey(aectx, "Parent_TYPE", "", 1, nil)
	db.RunInTransaction(ctx, func(c appengine.Context) error {
		_, err := db.PutMulti(ctx, []interface{}{
			&E{0, k, "Name 1"},
			&E{0, k, "Name 2"},
			&E{0, k, "Name 3"},
			&E{0, k, "Name 4"},
			&E{0, k, "Name 5"},
			&E{0, k, "Name 6"},
			&E{0, k, "Name 7"},
			&E{0, k, "Name 8"},
			&E{0, k, "Name 9"},
			&E{0, k, "Name 10"},
			&E{0, k, "Name 11"},
			&E{0, k, "Name 12"},
			&E{0, k, "Name 12"},
			&E{0, k, "Name 13"},
		})
		if err != nil {
			t.Error(err.Error())
		}
		return nil
	}, nil)

	req, _ := http.NewRequest("GET", "http://www.example.com/page?limit=3", nil)
	p := Paged(func(ctx context.Context, req *http.Request) (*datastore.Query, error) {
		return datastore.NewQuery("E").Ancestor(k).Order("Name"), nil
	}, E{})

	recorder := httptest.NewRecorder()
	db.RunInTransaction(ctx, func(c appengine.Context) error {
		p.ServeHTTP(keys.WithAEContext(ctx, c), recorder, req).Encode(recorder)
		return nil
	}, nil)
	var resp struct {
		PageToken string
		Response  []E
	}
	json.NewDecoder(recorder.Body).Decode(&resp)
	if len(resp.Response) != 3 || resp.PageToken == "" {
		t.Error("Paged failed.", resp)
	}

	aectx.Close()
}
