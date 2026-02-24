package store

import (
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	TDateTime = reflect.TypeFor[DateTime]()
)

type DateTimeCodec struct{}

func (d *DateTimeCodec) EncodeValue(_ bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != TDateTime {
		return bson.ValueEncoderError{Name: "DateTimeEncodeValue", Types: []reflect.Type{TDateTime}, Received: val}
	}
	td, ok := val.Interface().(DateTime)
	if !ok {
		return errors.New("failed to encode date time")
	}
	tt, err := td.Time()
	if err != nil {
		return bson.ValueEncoderError{Name: "DateTimeEncodeValue", Types: []reflect.Type{TDateTime}, Received: val}
	}
	return vw.WriteDateTime(tt.UnixMilli())
}

func (d *DateTimeCodec) DecodeValue(_ bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != TDateTime {
		return bson.ValueDecoderError{Name: "DecimalDecodeValue", Types: []reflect.Type{TDateTime}, Received: val}
	}

	var dateTimeVal DateTime
	//nolint:exhaustive
	switch t := vr.Type(); t {
	case bson.TypeDateTime:
		dt, err := vr.ReadDateTime()
		if err != nil {
			return err
		}
		dateTimeVal = NewDateTime(time.UnixMilli(dt))
	case bson.TypeString:
		decimalStr, err := vr.ReadString()
		if err != nil {
			return err
		}
		dateTimeVal = DateTime(decimalStr)
	default:
		return fmt.Errorf("cannot decode %v into a DateTime", t) //nolint:goerr113
	}

	val.Set(reflect.ValueOf(dateTimeVal))
	return nil
}
