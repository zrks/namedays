package namedays

import "errors"

var (
	NotFoundErr = errors.New("not found")
)

type Nameday struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

type MemStore struct {
	list map[string]Nameday
}

func NewMemStore() *MemStore {
	return &MemStore{
		list: make(map[string]Nameday),
	}
}

func (m *MemStore) Add(name string, nameday Nameday) error {
	m.list[name] = nameday
	return nil
}

func (m *MemStore) Get(name string) (Nameday, error) {
	if val, ok := m.list[name]; ok {
		return val, nil
	}
	return Nameday{}, NotFoundErr
}

func (m *MemStore) List() (map[string]Nameday, error) {
	return m.list, nil
}

func (m *MemStore) Update(name string, nameday Nameday) error {
	if _, ok := m.list[name]; ok {
		m.list[name] = nameday
		return nil
	}
	return NotFoundErr
}

func (m *MemStore) Remove(name string) error {
	if _, ok := m.list[name]; ok {
		delete(m.list, name)
		return nil
	}
	return NotFoundErr
}
