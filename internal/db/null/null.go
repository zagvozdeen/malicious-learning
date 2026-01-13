package null

import (
	"database/sql"
	"encoding/json/jsontext"
	"fmt"
	"strconv"
	"strings"
)

type String struct {
	sql.Null[string]
}

func (v String) MarshalJSONTo(enc *jsontext.Encoder) error {
	if !v.Valid {
		return enc.WriteToken(jsontext.Null)
	}
	return enc.WriteToken(jsontext.String(v.V))
}

func (v *String) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		v.V, v.Valid = "", false
		return nil
	case '"':
		v.V, v.Valid = tok.String(), true
		return nil
	default:
		return fmt.Errorf("NullString: expected string or null, got %s", tok.Kind().String())
	}
}

func NewString(s string, valid bool) String {
	return String{
		Null: sql.Null[string]{
			V:     s,
			Valid: valid,
		},
	}
}

func WrapString(s string) String {
	return NewString(s, s != "")
}

type Int struct {
	sql.Null[int]
}

func (v Int) MarshalJSONTo(enc *jsontext.Encoder) error {
	if !v.Valid {
		return enc.WriteToken(jsontext.Null)
	}
	return enc.WriteToken(jsontext.Int(int64(v.V)))
}

func (v *Int) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		v.V, v.Valid = 0, false
		return nil
	case '0':
		raw := tok.String()
		if strings.IndexAny(raw, ".eE") >= 0 {
			return fmt.Errorf("NullInt64: non-integer number %q", raw)
		}
		v.V, v.Valid = int(tok.Int()), true
		return nil
	case '"':
		n, err := strconv.ParseInt(tok.String(), 10, 64)
		if err != nil {
			return fmt.Errorf("NullInt64: bad integer string %q: %w", tok.String(), err)
		}
		v.V, v.Valid = int(n), true
		return nil
	default:
		return fmt.Errorf("NullInt64: expected number, string(number) or null, got %s", tok.Kind().String())
	}
}

func NewInt(v int, valid bool) Int {
	return Int{
		Null: sql.Null[int]{
			V:     v,
			Valid: valid,
		},
	}
}

func WrapInt(v int) Int {
	return NewInt(v, v != 0)
}
