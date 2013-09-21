package kvstore

func NewGobKVStore(fpath string, key, value interface{}) (KVStore, error) {
	return NewKVStore(fpath, key, value, gobCoder{})
}

func NewKVStore(fpath string, key, value interface{}, coder Coder) (KVStore, error) {
	p, err := newKVStore(fpath, key, value, coder)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *kvstore) Get(key interface{}) (interface{}, error) {
	return p.cache[key], nil
}

func (p *kvstore) Put(key, value interface{}) error {
	p.cache[key] = value
	return p.store()
}

func (p *kvstore) Del(key interface{}) error {
	delete(p.cache, key)
	return p.store()
}
