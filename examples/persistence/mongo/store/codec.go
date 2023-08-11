package store

import (
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var (
	TDateTime = reflect.TypeOf(DateTime(""))
)

type DateTimeCodec struct{}

func (d *DateTimeCodec) EncodeValue(_ bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != TDateTime {
		return bsoncodec.ValueEncoderError{Name: "DateTimeEncodeValue", Types: []reflect.Type{TDateTime}, Received: val}
	}
	td, ok := val.Interface().(DateTime)
	if !ok {
		return errors.New("failed to encode date time")
	}
	tt, err := td.Time()
	if err != nil {
		return bsoncodec.ValueEncoderError{Name: "DateTimeEncodeValue", Types: []reflect.Type{TDateTime}, Received: val}
	}
	return vw.WriteDateTime(tt.UnixMilli())
}

func (d *DateTimeCodec) DecodeValue(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != TDateTime {
		return bsoncodec.ValueDecoderError{Name: "DecimalDecodeValue", Types: []reflect.Type{TDateTime}, Received: val}
	}

	var dateTimeVal DateTime
	//nolint:exhaustive
	switch t := vr.Type(); t {
	case bsontype.DateTime:
		dt, err := vr.ReadDateTime()
		if err != nil {
			return err
		}
		dateTimeVal = NewDateTime(time.UnixMilli(dt))
	case bsontype.String:
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
