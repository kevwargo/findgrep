package config

import "reflect"

var optFieldCount = reflect.TypeOf(Option{}).NumField()

func (o *Option) merge(src Option) {
	this := reflect.ValueOf(o).Elem()
	other := reflect.ValueOf(src)

	for i := 0; i < optFieldCount; i++ {
		if newVal := other.Field(i); !newVal.IsZero() {
			this.Field(i).Set(newVal)
		}
	}
}
