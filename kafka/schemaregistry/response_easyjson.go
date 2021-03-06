// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package schemaregistry

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry(in *jlexer.Lexer, out *ResSchema) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "schema":
			out.Schema = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry(out *jwriter.Writer, in ResSchema) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"schema\":"
		out.RawString(prefix[1:])
		out.String(string(in.Schema))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResSchema) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResSchema) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResSchema) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResSchema) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry(l, v)
}
func easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry1(in *jlexer.Lexer, out *ResRegisterNewSchema) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry1(out *jwriter.Writer, in ResRegisterNewSchema) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int(int(in.ID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResRegisterNewSchema) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResRegisterNewSchema) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResRegisterNewSchema) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResRegisterNewSchema) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry1(l, v)
}
func easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry2(in *jlexer.Lexer, out *ResGetSubjectVersion) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "version":
			out.Version = int(in.Int())
		case "schema":
			out.Schema = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry2(out *jwriter.Writer, in ResGetSubjectVersion) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix[1:])
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"version\":"
		out.RawString(prefix)
		out.Int(int(in.Version))
	}
	{
		const prefix string = ",\"schema\":"
		out.RawString(prefix)
		out.String(string(in.Schema))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResGetSubjectVersion) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResGetSubjectVersion) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResGetSubjectVersion) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResGetSubjectVersion) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry2(l, v)
}
func easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry3(in *jlexer.Lexer, out *ResConfig) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "compatibilityLevel":
			out.Compatibility = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry3(out *jwriter.Writer, in ResConfig) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"compatibilityLevel\":"
		out.RawString(prefix[1:])
		out.String(string(in.Compatibility))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResConfig) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResConfig) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResConfig) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResConfig) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry3(l, v)
}
func easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry4(in *jlexer.Lexer, out *ResCheckSubject) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "subject":
			out.Subject = string(in.String())
		case "id":
			out.ID = int(in.Int())
		case "version":
			out.Version = int(in.Int())
		case "schema":
			out.Schema = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry4(out *jwriter.Writer, in ResCheckSubject) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"subject\":"
		out.RawString(prefix[1:])
		out.String(string(in.Subject))
	}
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix)
		out.Int(int(in.ID))
	}
	{
		const prefix string = ",\"version\":"
		out.RawString(prefix)
		out.Int(int(in.Version))
	}
	{
		const prefix string = ",\"schema\":"
		out.RawString(prefix)
		out.String(string(in.Schema))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResCheckSubject) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResCheckSubject) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6ff3ac1dEncodeGithubComDialogsDialogGoLibKafkaSchemaregistry4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResCheckSubject) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResCheckSubject) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6ff3ac1dDecodeGithubComDialogsDialogGoLibKafkaSchemaregistry4(l, v)
}
