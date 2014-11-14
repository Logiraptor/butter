package db

import (
	"errors"
	"reflect"

	"github.com/qedus/nds"

	"appengine"
	"appengine/datastore"
)

var getterType = reflect.TypeOf(Getter(nil))
var putterType = reflect.TypeOf(Putter(nil))

// Getter specifies a type which can be retrieved from
// the datastore.
type Getter interface {
	Get(appengine.Context, *datastore.Key) error
}

// Putter specifies a type which can be saved to
// the datastore.
type Putter interface {
	Put(appengine.Context) (*datastore.Key, error)
}

// invokeGet implements the post-get functionality
// which is a primary feature of the butter/db package.
func invokeGet(i reflect.Value, ctx appengine.Context, key *datastore.Key) error {
	if i.Type().Implements(getterType) {
		return i.Interface().(Getter).Get(ctx, key)
	}
	// recursive call
	for i.Type().Kind() == reflect.Ptr {
		i = i.Elem()
	}
	if i.Type().Kind() != reflect.Struct {
		return nil
	}
	n := i.NumField()
	for j := 0; j < n; j++ {
		err := invokeGet(i.Field(j), ctx, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// invokePut implements the post-Put functionality
// which is a primary feature of the butter/db package.
func invokePut(i reflect.Value, ctx appengine.Context) (*datastore.Key, error) {
	if i.Type().Implements(putterType) {
		return i.Interface().(Putter).Put(ctx)
	}
	// recursive call
	for i.Type().Kind() == reflect.Ptr {
		i = i.Elem()
	}
	if i.Type().Kind() != reflect.Struct {
		return datastore.NewKey(ctx, i.Type().Name(), "", 0, nil), nil
	}
	n := i.NumField()
	for j := 0; j < n; j++ {
		key, err := invokePut(i.Field(j), ctx)
		if err != nil {
			return nil, err
		}
		if key != nil {
			return key, nil
		}
	}
	return datastore.NewKey(ctx, i.Type().Name(), "", 0, nil), nil
}

func rangeInterface(f func(interface{}) error, list interface{}) error {
	val := reflect.ValueOf(list)
	if val.Kind() != reflect.Slice {
		return errors.New("argument must be a slice")
	}
	length := val.Len()
	for i := 0; i < length; i++ {
		elem := val.Index(i)
		err := f(elem.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func invokeGetMulti(ctx appengine.Context, src interface{}, keys []*datastore.Key) error {
	i := 0
	return rangeInterface(func(x interface{}) error {
		err := invokeGet(reflect.ValueOf(x), ctx, keys[i])
		if err != nil {
			return err
		}
		i++
		return nil
	}, src)
}

func invokePutMulti(ctx appengine.Context, src interface{}) ([]*datastore.Key, error) {
	var keys []*datastore.Key
	err := rangeInterface(func(x interface{}) error {
		key, err := invokePut(reflect.ValueOf(x), ctx)
		if err != nil {
			return err
		}
		keys = append(keys, key)
		return nil
	}, src)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Delete See https://cloud.google.com/appengine/docs/go/datastore/reference#Delete
func Delete(c appengine.Context, key *datastore.Key) error {
	return nds.Delete(c, key)
}

// DeleteMulti See https://cloud.google.com/appengine/docs/go/datastore/reference#DeleteMulti
func DeleteMulti(c appengine.Context, key []*datastore.Key) error {
	return nds.DeleteMulti(c, key)
}

// Get See https://cloud.google.com/appengine/docs/go/datastore/reference#Get
func Get(c appengine.Context, key *datastore.Key, dst interface{}) error {
	err := nds.Get(c, key, dst)
	if err != nil {
		return err
	}
	return invokeGet(reflect.ValueOf(dst), c, key)
}

// GetMulti See https://cloud.google.com/appengine/docs/go/datastore/reference#GetMulti
func GetMulti(c appengine.Context, keys []*datastore.Key, dst interface{}) error {
	err := nds.GetMulti(c, keys, dst)
	if err != nil {
		return err
	}
	return invokeGetMulti(c, dst, keys)
}

// Put See https://cloud.google.com/appengine/docs/go/datastore/reference#Put
func Put(c appengine.Context, src interface{}) (*datastore.Key, error) {
	key, err := invokePut(reflect.ValueOf(src), c)
	if err != nil {
		return nil, err
	}
	return nds.Put(c, key, src)
}

// PutMulti See https://cloud.google.com/appengine/docs/go/datastore/reference#PutMulti
func PutMulti(c appengine.Context, src interface{}) ([]*datastore.Key, error) {
	keys, err := invokePutMulti(c, src)
	if err != nil {
		return nil, err
	}
	return nds.PutMulti(c, keys, src)
}

// RunInTransaction See https://cloud.google.com/appengine/docs/go/datastore/reference#RunInTransaction
func RunInTransaction(c appengine.Context, f func(tc appengine.Context) error, opts *datastore.TransactionOptions) error {
	return nds.RunInTransaction(c, f, opts)
}

// Run returns an iterator for a given query
func Run(ctx appengine.Context, q *datastore.Query) Iterator {
	return iterWrapper{ctx, q.Run(ctx)}
}

// GetAll returns all results for a query
func GetAll(ctx appengine.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	keys, err := q.GetAll(ctx, dst)
	if err != nil {
		return nil, err
	}
	err = invokeGetMulti(ctx, dst, keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// GetN fetches N results from the query
// Returns the keys, a pagetoken, and an error
func GetN(ctx appengine.Context, q *datastore.Query, dst interface{}, n int) ([]*datastore.Key, string, error) {
	out := reflect.ValueOf(dst)
	if out.Kind() != reflect.Ptr || out.Elem().Kind() != reflect.Slice {
		return nil, "", errors.New("dst must be a *[]*T")
	}

	var (
		count    = 0
		iter     = Run(ctx, q)
		slice    = out.Elem()
		elemType = out.Type().Elem().Elem()
		keys     = make([]*datastore.Key, n)
	)

	for count < n {
		e := reflect.New(elemType)
		x := e.Interface()
		k, err := iter.Next(x)
		if err != nil {
			if err == datastore.Done {
				break
			} else {
				return nil, "", err
			}
		}

		keys[count] = k
		slice = reflect.Append(slice, e.Elem())
		count++
	}

	out.Elem().Set(slice)

	c, err := iter.Cursor()
	if err != nil {
		return nil, "", err
	}

	return keys, c.String(), nil
}

// Iterator abstracts the datastore.Iterator
type Iterator interface {
	Cursor() (datastore.Cursor, error)
	Next(dst interface{}) (*datastore.Key, error)
}

type iterWrapper struct {
	ctx appengine.Context
	Iterator
}

func (i iterWrapper) Next(dst interface{}) (*datastore.Key, error) {
	k, err := i.Iterator.Next(dst)
	if err != nil {
		return nil, err
	}
	err = invokeGet(reflect.ValueOf(dst), i.ctx, k)
	if err != nil {
		return nil, err
	}
	return k, nil
}
