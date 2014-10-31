package db

import (
	"errors"
	"reflect"

	"appengine"

	"github.com/qedus/nds"

	"appengine/datastore"
)

var (
	keyType = reflect.TypeOf((*datastore.Key)(nil))
)

// An OnGetter has its callback called immediately after Get, GetMulti, or Query.Next
type OnGetter interface {
	OnGet(ctx appengine.Context) error
}

// An OnPutter has its callback called immediately before Put or PutMulti
type OnPutter interface {
	OnPut(ctx appengine.Context) error
}

// Delete deletes the entity associated with key
func Delete(c appengine.Context, key *datastore.Key) error {
	return nds.Delete(c, key)
}

// DeleteMulti deletes the entity associated with keys
func DeleteMulti(c appengine.Context, keylist []*datastore.Key) error {
	return nds.DeleteMulti(c, keylist)
}

// Get fills val in based on its key as returned by Key ID and Parent
func Get(c appengine.Context, val interface{}) error {
	err := nds.Get(c, Key(c, val), val)
	if err != nil {
		return err
	}
	if g, ok := val.(OnGetter); ok {
		return g.OnGet(c)
	}
	return nil
}

// GetMulti fills in the values in vals based on their keys as returned by Keys
func GetMulti(c appengine.Context, vals interface{}) error {
	err := nds.GetMulti(c, Keys(c, vals), vals)
	if err != nil {
		return err
	}
	return rangeInterface(func(i interface{}) error {
		if g, ok := i.(OnGetter); ok {
			err := g.OnGet(c)
			if err != nil {
				return err
			}
		}
		return nil
	}, vals)
}

// Put inserts val into the database under the key returned by Key
func Put(c appengine.Context, val interface{}) (*datastore.Key, error) {
	if g, ok := val.(OnPutter); ok {
		return k, g.OnPut(c)
	}
	k, err := nds.Put(c, Key(c, val), val)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// PutMulti inserts vals into the database under the keys as returned by Keys
func PutMulti(c appengine.Context, vals interface{}) ([]*datastore.Key, error) {
	err := rangeInterface(func(i interface{}) error {
		if g, ok := i.(OnPutter); ok {
			err := g.OnPut(c)
			if err != nil {
				return err
			}
		}
		return nil
	}, vals)
	if err != nil {
		return nil, err
	}
	keys, err := nds.PutMulti(c, Keys(c, vals), vals)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// RunInTransaction runs f within a transaction
func RunInTransaction(c appengine.Context, f func(tc appengine.Context) error, opts *datastore.TransactionOptions) error {
	return nds.RunInTransaction(c, f, opts)
}

// Key returns a key based on fields in src.
// Options are:
// An int64 field named ID
// A datastore.Key field named Parent
func Key(ctx appengine.Context, src interface{}) *datastore.Key {
	val := reflect.ValueOf(src)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		panic("cannot get key of non-struct type")
	}
	idField := val.FieldByName("ID")
	parentField := val.FieldByName("Parent")
	var (
		id     int64
		parent *datastore.Key
	)
	if idField.IsValid() && idField.Kind() == reflect.Int64 {
		id = idField.Int()
	}
	if parentField.IsValid() && parentField.Type().AssignableTo(keyType) {
		parent = parentField.Interface().(*datastore.Key)
	}
	return datastore.NewKey(ctx, val.Type().Name(), "", id, parent)
}

// Keys applies Key to all elements in src. src must be a slice.
func Keys(ctx appengine.Context, src interface{}) []*datastore.Key {
	var keys []*datastore.Key
	err := rangeInterface(func(i interface{}) error {
		keys = append(keys, Key(ctx, i))
		return nil
	}, src)
	if err != nil {
		panic(err.Error())
	}
	return keys
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

// GetN stores n values from query in dst
func GetN(cx appengine.Context, query *datastore.Query, dst interface{}, n int) ([]*datastore.Key, string, error) {
	ctx := cx
	out := reflect.ValueOf(dst)
	if out.Kind() != reflect.Ptr || out.Elem().Kind() != reflect.Slice {
		return nil, "", errors.New("dst must be a *[]T")
	}

	var (
		count    = 0
		iter     = query.Run(ctx)
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
		if getter, ok := x.(OnGetter); ok {
			err = getter.OnGet(cx)
			if err != nil {
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
