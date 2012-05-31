package main

type list []string

func (l *list) Append(val string) {
	*l = append(*l, val)
}

func (l list) Dup() (r list) {
	r = make(list, 0, len(l))
	for _, v := range l {
		r = append(r, v)
	}
	return
}

type Meta struct {
	CSS       list
	JS        list
	BaseTitle string
	Title     string
}

func (m *Meta) Dup() *Meta {
	return &Meta{
		CSS:       m.CSS.Dup(),
		JS:        m.JS.Dup(),
		Title:     m.Title,
		BaseTitle: m.BaseTitle,
	}
}
