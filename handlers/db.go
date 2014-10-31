package handlers

import (
	"net/http"
	"reflect"
	"strconv"

	"code.google.com/p/go.net/context"

	"github.com/Logiraptor/butter/db"

	"github.com/Logiraptor/nap"

	"appengine/datastore"
)

// QueryFunc defines a handler which creates a query for any http.Request
type QueryFunc func(ctx context.Context, req *http.Request) (*datastore.Query, error)

// Paged wraps a queryfunc and provides seamless paging functionality
// TODO: add mapping mechanism for this.
// Maybe it's good enough to implement a seperate handler when in-memory filtering is needed.
func Paged(qf QueryFunc, elemType interface{}) Handler {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) Response {
		query, err := qf(ctx, req)
		if err != nil {
			return nap.JSONError(400, err.Error())
		}

		limit := req.FormValue("limit")
		var lim = 10
		if limit != "" {
			lim, err = strconv.Atoi(limit)
			if err != nil {
				return nap.JSONError(400, err.Error())
			}
		}

		pageToken := req.FormValue("pageToken")
		if pageToken != "" {
			c, err := datastore.DecodeCursor(pageToken)
			if err != nil {
				return nap.JSONError(400, err.Error())
			}
			query = query.Start(c)
		}

		dstType := reflect.SliceOf(reflect.TypeOf(elemType))
		dstVal := reflect.New(dstType)
		dst := dstVal.Interface()

		_, token, err := db.GetN(ctx, query, dst, lim)
		if err != nil {
			return nap.JSONError(500, err.Error())
		}

		return nap.JSON(map[string]interface{}{
			"PageToken": token,
			"Response":  dst,
		})
	})
}

// KeyFunc takes a context and a request and returns an entity to be served.
// The entity mush have it's ID/Parent field set if it is retrieving from db.
type KeyFunc func(ctx context.Context, req *http.Request) (interface{}, error)

// GetInstance creates a simple instance getter handler
func GetInstance(kf KeyFunc, elemType interface{}) Handler {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) Response {
		entity, err := kf(ctx, req)
		if err != nil {
			return nap.JSONError(400, err.Error())
		}

		key := db.Key(ctx, entity)
		if key.Incomplete() {
			return nap.JSON(entity)
		}

		err = db.Get(ctx, entity)
		if err != nil {
			return nap.JSONError(500, err.Error())
		}

		return nap.JSON(entity)
	})
}
