package avro

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type DecodeSubObject1 struct {
	Field1 string
	Field2 int32
}

type DecodeSubObject2 struct {
	Field1    bool
	Field2    []byte
	SubStruct *DecodeSubObject1
}

type DecodeTarget struct {
	UnknownField   string `avro:"unknown"`
	fStrCantSet    string `avro:"fStrCantSet"`
	FStrWithoutTag string
	FStrWithTag    string `avro:"f_str_with_tag"`
	FStrOptional   string
	FStrPointer    *string

	FInt32WithTag  int32 `avro:"f_int32_with_tag"`
	FInt32Optional int32
	FInt32Pointer  *int32

	FInt64WithTag  int64 `avro:"f_int64_with_tag"`
	FInt64Optional int64
	FInt64Pointer  *int64

	FFloatWithTag  float32 `avro:"f_float_with_tag"`
	FFloatOptional float32
	FFloatPointer  *float32

	FDoubleWithTag  float64 `avro:"f_double_with_tag"`
	FDoubleOptional float64
	FDoublePointer  *float64

	FBoolWithTag  bool `avro:"f_bool_with_tag"`
	FBoolOptional bool
	FBoolPointer  *bool

	FBytesWithTag  []byte `avro:"f_bytes_with_tag"`
	FBytesOptional []byte
	FBytesPointer  *[]byte

	FArrStrWithTag  []string `avro:"f_arr_str_with_tag"`
	FArrStrOptional []string
	FArrStrPointer  *[]string

	FArrStruct            []DecodeSubObject1
	FArrStructPointer     *[]DecodeSubObject1
	FArrStructWithPointer *[]*DecodeSubObject1

	FArrInt32WithTag  []int32 `avro:"f_arr_int32_with_tag"`
	FArrInt32Optional []int32
	FArrInt32Pointer  *[]int32

	FWithSubObject1    interface{}
	FWithSubObject2    interface{}
	FWithSubObjectNull *DecodeSubObject1

	FWithSubObject1NotPointer DecodeSubObject1  `avro:"fWithSubObject1"`
	FWithSubObject2Pointer    *DecodeSubObject2 `avro:"fWithSubObject2"`

	FMapWithBytes  map[string][]byte
	FMapWithString map[string]*[]string
	FMapWithInt32  map[string][]int32

	FEmum string
}

func init() {
	Register(&DecodeSubObject1{})
	Register(DecodeSubObject2{})
}

func TestDecodeToNotStruct(t *testing.T) {

	require.EqualError(t,
		DecodeToStruct(1, map[string]interface{}{}),
		"the target is not a struct")
}

func TestDoubleRegister(t *testing.T) {

	type TypeTestDoubleRegister struct{}

	Register(&TypeTestDoubleRegister{})

	defer func() {
		if e := recover(); e != nil {
			require.Equal(t, "object with name already registered: TypeTestDoubleRegister", e)
		}
	}()

	Register(&TypeTestDoubleRegister{})
}

func TestGetRegisteredType(t *testing.T) {

	type TypeTestGetRegisteredType struct{}

	{
		res, ok := getRegisteredType("TypeTestGetRegisteredType")
		require.False(t, ok)
		require.Nil(t, res)
	}

	Register(&TypeTestGetRegisteredType{})

	{
		res, ok := getRegisteredType("TypeTestGetRegisteredType")
		require.True(t, ok)
		require.Equal(t, reflect.TypeOf(TypeTestGetRegisteredType{}), res)
	}
}

func TestUnsupportedSimpleType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fieldName", "type": "string"}
		]}`

	srcVal := map[string]interface{}{
		"fieldName": "value",
	}

	type Dest struct {
		FieldName int
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		"field: fieldName: failed to set int value: unsupported type")
}

func TestUnsupportedSliceType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fieldName", "type": {"type":"array","items":"string"}}
		]}`

	srcVal := map[string]interface{}{
		"fieldName": []string{},
	}

	type Dest struct {
		FieldName []int
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		"field: fieldName: failed to set slice[int] value: unsupported type")
}

func TestInvalidSimpleType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fieldName", "type": "string"}
		]}`

	srcVal := map[string]interface{}{
		"fieldName": "value",
	}

	type Dest struct {
		FieldName int32
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		`field: fieldName: failed to set int32 value: "value"`)
}

func TestInvalidSliceType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fieldName", "type": {"type":"array","items":"string"}}
		]}`

	srcVal := map[string]interface{}{
		"fieldName": []string{"value"},
	}

	type Dest struct {
		FieldName []int32
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		`field: fieldName: failed to set slice value: "value"(item: 0)`)
}

func TestInvalidBytesSliceType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fieldName", "type": {"type":"array","items":"string"}}
		]}`

	srcVal := map[string]interface{}{
		"fieldName": []string{"value"},
	}

	type Dest struct {
		FieldName []byte
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		`field: fieldName: failed to set slice value: []interface {}{"value"}`)
}

func TestUnregisteredStructType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fieldName", "type": {"type":"record","name":"DecodeSubObject",
			"fields":[
				{"name": "field1","type": "string"},
				{"name": "field2","type": "int"}
			]}}
		]}`

	srcVal := map[string]interface{}{
		"fieldName": map[string]interface{}{
			"field1": "str",
			"field2": 1,
		},
	}

	type DecodeSubObject struct{}

	type Dest struct {
		FieldName DecodeSubObject
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		`field: fieldName: failed to set struct value: is not registered type with name 'DecodeSubObject': map[field1:str field2:1]`)
}

func TestUnregisteredSubStructType(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{
			"name": "struct",
			"type": {
				"type":"record",
				"name":"DecodeObject",
				"fields":[
					{
						"name": "subStruct",
						"type": {
							"type":"record",
							"name":"DecodeSubObject",
							"fields":[{"name": "str","type": "string"}]
						}
					}
				]
			}}
		]}`

	srcVal := map[string]interface{}{
		"struct": map[string]interface{}{
			"subStruct": map[string]interface{}{
				"str": "value",
			},
		},
	}

	type DecodeSubObject struct {
	}

	type DecodeObject struct {
		SubStruct DecodeSubObject
	}
	Register(&DecodeObject{})

	type Dest struct {
		Struct DecodeObject
	}

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.EqualError(t,
		DecodeToStruct(dest, decVal),
		`field: struct: field: subStruct: failed to set struct value: is not registered type with name 'DecodeSubObject': map[str:value]`)
}

func TestRegisterName(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"namespace": "im.dlg.project.avro",
		"fields": [
			{"name": "fieldName", "type": ["null", {
			"type":"record",

			"name":"CustomObjectName",
			"fields":[
				{"name": "field1","type": "string"},
				{"name": "field2","type": "int"}
			]}] }
		]}`

	srcVal := map[string]interface{}{
		"fieldName": map[string]interface{}{
			"im.dlg.project.avro.CustomObjectName": map[string]interface{}{
				"field1": "str",
				"field2": 1,
			},
		},
	}

	type DecodeSubObject struct {
		Field1 string
		Field2 int32
	}

	type Dest struct {
		FieldName interface{}
	}

	RegisterName("im.dlg.project.avro.CustomObjectName", DecodeSubObject{})

	decVal := getDecodedValue(t, schema, srcVal)
	dest := &Dest{}
	require.NoError(t, DecodeToStruct(dest, decVal))
	require.Equal(t,
		&Dest{
			FieldName: &DecodeSubObject{
				Field1: "str",
				Field2: 1,
			},
		},
		dest)
}

func TestDecode(t *testing.T) {

	schema, val := getDefaultSchemaAndValue()
	src := getDecodedValue(t, schema, val)

	target := DecodeTarget{}
	require.NoError(t, DecodeToStruct(&target, src))
	require.Equal(t,
		DecodeTarget{
			FStrWithoutTag: "val1",
			FStrWithTag:    "val2",
			FStrOptional:   "val3",
			FStrPointer:    func(v string) *string { return &v }("val4"),

			FInt32WithTag:  1,
			FInt32Optional: 2,
			FInt32Pointer:  func(v int32) *int32 { return &v }(3),

			FInt64WithTag:  4,
			FInt64Optional: 5,
			FInt64Pointer:  func(v int64) *int64 { return &v }(6),

			FFloatWithTag:  7.1,
			FFloatOptional: 8.2,
			FFloatPointer:  func(v float32) *float32 { return &v }(9.3),

			FDoubleWithTag:  10.1,
			FDoubleOptional: 11.2,
			FDoublePointer:  func(v float64) *float64 { return &v }(12.3),

			FBoolWithTag:  true,
			FBoolOptional: true,
			FBoolPointer:  func(v bool) *bool { return &v }(true),

			FBytesWithTag:  []byte{0x01, 0x02},
			FBytesOptional: []byte{0x03, 0x04},
			FBytesPointer:  func(v []byte) *[]byte { return &v }([]byte{0x05, 0x06}),

			FArrStrWithTag:  []string{"0x07", "0x08"},
			FArrStrOptional: []string{"0x09", "0x10"},
			FArrStrPointer:  func(v []string) *[]string { return &v }([]string{"0x11", "0x12"}),

			FArrStruct: []DecodeSubObject1{
				{Field1: "text1", Field2: 1},
				{Field1: "text2", Field2: 2},
			},

			FArrStructPointer: func(v []DecodeSubObject1) *[]DecodeSubObject1 { return &v }([]DecodeSubObject1{
				{Field1: "text3", Field2: 3},
				{Field1: "text4", Field2: 4},
			}),

			FArrStructWithPointer: func(v []*DecodeSubObject1) *[]*DecodeSubObject1 { return &v }([]*DecodeSubObject1{
				{Field1: "text5", Field2: 5},
				{Field1: "text6", Field2: 6},
			}),

			FArrInt32WithTag:  []int32{13, 14},
			FArrInt32Optional: []int32{15, 16},
			FArrInt32Pointer:  func(v []int32) *[]int32 { return &v }([]int32{17, 18}),

			FWithSubObject1: &DecodeSubObject1{
				Field1: "text",
				Field2: 1,
			},
			FWithSubObject2: &DecodeSubObject2{
				Field1: true,
				Field2: []byte{0x01, 0x02, 0x03},
				SubStruct: &DecodeSubObject1{
					Field1: "text1",
					Field2: 2,
				},
			},

			FWithSubObject1NotPointer: DecodeSubObject1{
				Field1: "text",
				Field2: 1,
			},

			FWithSubObject2Pointer: &DecodeSubObject2{
				Field1: true,
				Field2: []byte{0x01, 0x02, 0x03},
				SubStruct: &DecodeSubObject1{
					Field1: "text1",
					Field2: 2,
				},
			},

			FMapWithBytes: map[string][]byte{
				"a1": []byte{0x01, 0x02},
				"a2": []byte{0x03, 0x04},
			},

			FMapWithString: map[string]*[]string{
				"b1": func(v []string) *[]string { return &v }([]string{"0x05", "0x06"}),
				"b2": func(v []string) *[]string { return &v }([]string{"0x07", "0x08"}),
			},

			FMapWithInt32: map[string][]int32{
				"c1": []int32{9},
				"c2": []int32{10},
			},

			FEmum: "Val1",
		},
		target)

	require.Empty(t, target.fStrCantSet)
}

func BenchmarkDecodeToStruct(b *testing.B) {

	ts := time.Now()
	defer func() { fmt.Println("\nbenchmark time:", time.Since(ts), " items:", b.N) }()

	schema, val := getDefaultSchemaAndValue()
	s, err := New(schema)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		enc, err := s.Encode(val)
		if err != nil {
			b.Fatal(err)
		}

		decSrc, err := s.Decode(enc)
		if err != nil {
			b.Fatal(err)
		}
		m := decSrc.(map[string]interface{})

		target := DecodeTarget{}
		if err := DecodeToStruct(&target, m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNative(b *testing.B) {

	ts := time.Now()
	defer func() { fmt.Println("\nbenchmark time:", time.Since(ts), " items:", b.N) }()

	schema, val := getDefaultSchemaAndValue()
	s, err := New(schema)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		enc, err := s.Encode(val)
		if err != nil {
			b.Fatal(err)
		}

		decSrc, err := s.Decode(enc)
		if err != nil {
			b.Fatal(err)
		} else if len(val) != len(decSrc.(map[string]interface{})) {
			b.Fatal("invalid result")
		}
	}
}

func getDefaultSchemaAndValue() (string, map[string]interface{}) {

	subObject1 := `{"type":"record","name":"DecodeSubObject1",
      "fields":[
		  {"name": "field1","type": "string"},
		  {"name": "field2","type": "int"}
	  ]}`
	subObject2 := `{"type":"record","name":"DecodeSubObject2",
	  "fields":[
		  {"name": "field1","type": "boolean"},
		  {"name": "field2","type": "bytes"},
		  {"name": "subStruct","type":"DecodeSubObject1"}
	  ]}`

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "fStrCantSet", "type": "string"},
			{"name": "fStrWithoutTag", "type": "string"},
			{"name": "f_str_with_tag", "type": "string"},
			{"name": "fStrOptional", "type": ["null","string"]},
			{"name": "fStrPointer", "type": "string"},
			{"name": "f_int32_with_tag", "type": "int"},
			{"name": "fInt32Optional", "type": ["null","int"]},
			{"name": "fInt32Pointer", "type": "int"},
			{"name": "f_int64_with_tag", "type": "long"},
			{"name": "fInt64Optional", "type": ["null","long"]},
			{"name": "fInt64Pointer", "type": "long"},
			{"name": "f_float_with_tag", "type": "float"},
			{"name": "fFloatOptional", "type": ["null","float"]},
			{"name": "fFloatPointer", "type": "float"},
			{"name": "f_double_with_tag", "type": "double"},
			{"name": "fDoubleOptional", "type": ["null","double"]},
			{"name": "fDoublePointer", "type": "double"},
			{"name": "f_bool_with_tag", "type": "boolean"},
			{"name": "fBoolOptional", "type": ["null","boolean"]},
			{"name": "fBoolPointer", "type": "boolean"},
			{"name": "f_bytes_with_tag", "type": "bytes"},
			{"name": "fBytesOptional", "type": ["null","bytes"]},
			{"name": "fBytesPointer", "type": "bytes"},

			{"name": "f_arr_str_with_tag", "type": {"type":"array","items":"string"}},
			{"name": "fArrStrOptional", "type": ["null", {"type":"array","items":"string"}]},
			{"name": "fArrStrPointer", "type": {"type":"array","items":"string"}},

			{"name":"fArrStruct","type":{"type":"array","items":` + subObject1 + `}},
			{"name":"fArrStructPointer","type":{"type":"array","items":` + subObject1 + `}},
			{"name":"fArrStructWithPointer","type":{"type":"array","items":` + subObject1 + `}},

			{"name": "f_arr_int32_with_tag", "type": {"type":"array","items":"int"}},
			{"name": "fArrInt32Optional", "type": ["null", {"type":"array","items":"int"}]},
			{"name": "fArrInt32Pointer", "type": {"type":"array","items":"int"}},

			{"name": "fWithSubObject1", "type": [` + subObject1 + `,` + subObject2 + `] },
			{"name": "fWithSubObject2", "type": [` + subObject1 + `,` + subObject2 + `] },
			{"name": "fWithSubObjectNull", "type": [` + subObject1 + `,"null"] },

			{"name": "fMapWithBytes", "type": {"type": "map", "values": "bytes"} },
			{"name": "fMapWithString", "type": {"type": "map", "values": {"type":"array","items":"string"}} },
			{"name": "fMapWithInt32", "type": {"type": "map", "values": {"type":"array","items":"int"}} },

			{"name":"fEmum", "type": {
              "type": "enum",
              "name": "MyEnum",
              "symbols": ["Val1","Val2"]
            }}
		]}`

	val := map[string]interface{}{
		"fStrCantSet":    "cant-set",
		"fStrWithoutTag": "val1",

		"f_str_with_tag": "val2",
		"fStrOptional":   map[string]interface{}{"string": "val3"},
		"fStrPointer":    "val4",

		"f_int32_with_tag": int32(1),
		"fInt32Optional":   map[string]interface{}{"int": int32(2)},
		"fInt32Pointer":    int32(3),

		"f_int64_with_tag": int64(4),
		"fInt64Optional":   map[string]interface{}{"long": int64(5)},
		"fInt64Pointer":    int64(6),

		"f_float_with_tag": float32(7.1),
		"fFloatOptional":   map[string]interface{}{"float": float32(8.2)},
		"fFloatPointer":    float32(9.3),

		"f_double_with_tag": float64(10.1),
		"fDoubleOptional":   map[string]interface{}{"double": float64(11.2)},
		"fDoublePointer":    float64(12.3),

		"f_bool_with_tag": true,
		"fBoolOptional":   map[string]interface{}{"boolean": true},
		"fBoolPointer":    true,

		"f_bytes_with_tag": []byte{0x01, 0x02},
		"fBytesOptional":   map[string]interface{}{"bytes": []byte{0x03, 0x04}},
		"fBytesPointer":    []byte{0x05, 0x06},

		"f_arr_str_with_tag": []string{"0x07", "0x08"},
		"fArrStrOptional":    map[string]interface{}{"array": []string{"0x09", "0x10"}},
		"fArrStrPointer":     []string{"0x11", "0x12"},

		"fArrStruct": []map[string]interface{}{
			{
				"field1": "text1",
				"field2": 1,
			},
			{
				"field1": "text2",
				"field2": 2,
			},
		},

		"fArrStructPointer": []map[string]interface{}{
			{
				"field1": "text3",
				"field2": 3,
			},
			{
				"field1": "text4",
				"field2": 4,
			},
		},

		"fArrStructWithPointer": []map[string]interface{}{
			{
				"field1": "text5",
				"field2": 5,
			},
			{
				"field1": "text6",
				"field2": 6,
			},
		},

		"f_arr_int32_with_tag": []int32{13, 14},
		"fArrInt32Optional":    map[string]interface{}{"array": []int32{15, 16}},
		"fArrInt32Pointer":     []int32{17, 18},

		"fWithSubObject1": map[string]interface{}{"DecodeSubObject1": map[string]interface{}{
			"field1": "text",
			"field2": 1,
		}},
		"fWithSubObject2": map[string]interface{}{"DecodeSubObject2": map[string]interface{}{
			"field1": true,
			"field2": []byte{0x01, 0x02, 0x03},
			"subStruct": map[string]interface{}{
				"field1": "text1",
				"field2": 2,
			},
		}},
		"fWithSubObjectNull": map[string]interface{}{"null": nil},

		"fMapWithBytes": map[string]interface{}{
			"a1": []byte{0x01, 0x02},
			"a2": []byte{0x03, 0x04},
		},

		"fMapWithString": map[string]interface{}{
			"b1": []string{"0x05", "0x06"},
			"b2": []string{"0x07", "0x08"},
		},

		"fMapWithInt32": map[string]interface{}{
			"c1": []int32{9},
			"c2": []int32{10},
		},

		"fEmum": "Val1",
	}

	return schema, val
}

func getDecodedValue(t *testing.T, schema string, val map[string]interface{}) map[string]interface{} {
	t.Helper()

	s, err := New(schema)
	require.NoError(t, err)

	enc, err := s.Encode(val)
	require.NoError(t, err)

	decSrc, err := s.Decode(enc)
	require.NoError(t, err)

	return decSrc.(map[string]interface{})
}
