//go:generate rice embed-go

package templates

import rice "github.com/GeertJohan/go.rice"

func NewBox() *rice.Box {
	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},
	}
	return conf.MustFindBox(".")
}
