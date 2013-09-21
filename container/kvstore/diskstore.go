package kvstore

import (
	"os"
	"reflect"
)

type kvstore struct {
	path    string
	cache   map[interface{}]interface{}
	coder   Coder
	keyType reflect.Type
	valType reflect.Type
}

func newKVStore(fpath string, key, value interface{}, coder Coder) (KVStore, error) {
	p := &kvstore{
		path:    fpath,
		cache:   make(map[interface{}]interface{}),
		coder:   coder,
		keyType: reflect.TypeOf(key),
		valType: reflect.TypeOf(value),
	}
	p.load()

	return p, nil
}

type kvpair struct {
	key   interface{}
	value interface{}
}

func (p *kvstore) decodeStream(dec Decoder) <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		for {
			key := reflect.New(p.keyType).Elem()
			value := reflect.New(p.valType).Elem()
			if err := dec.DecodeValue(key); err != nil {
				ch <- err
				close(ch)
				break
			}
			if err := dec.DecodeValue(value); err != nil {
				ch <- err
				close(ch)
				break
			}

			ch <- kvpair{
				key.Interface(),
				value.Interface(),
			}
		}
	}()
	return ch
}

func (p *kvstore) load() error {
	f, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer f.Close()

	for ret := range p.decodeStream(p.coder.NewDecoder(f)) {
		switch ret := ret.(type) {
		case kvpair:
			p.cache[ret.key] = ret.value
		case error:
			return err
		default:
			panic("invalid case")
		}
	}

	return nil
}

func (p *kvstore) store() error {
	f, err := os.Create(p.path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := p.coder.NewEncoder(f)
	for k, v := range p.cache {
		if err := enc.EncodeValue(reflect.ValueOf(k)); err != nil {
			return err
		}
		if err := enc.EncodeValue(reflect.ValueOf(v)); err != nil {
			return err
		}
	}

	return nil
}
