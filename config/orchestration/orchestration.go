package orchestrationConfig

import "reflect"

var SourceRefParsing = struct {
	Obj  string
	Http string
}{"${{ibridge.object}}", "${{ibridge.httprequest}}"}

var HttpRequest = struct {
	Auth      string
	Body      string
	Params    string
	FormFile  string
	FormValue string
	Query     string
	Headers   string
	Header    string
}{"Authorization", "Body", "Params", "FormFile", "FormValue", "QueryParams", "Headers", "Header"}

var BodySourceRef = struct {
	File  string
	Value string
	Json  string
}{"${{ibridge.formFile}}", "${{ibridge.formValue}}", "${{ibridge.json}}"}

var ConditionalCheck = struct {
	Nil    interface{}
	NotNil interface{}
}{"${{ibridgereflect.Value.IsNil}}", "${{ibridge!reflect.Value.IsNil}}"}

var Operator = struct {
	Nil           interface{}
	NotNil        interface{}
	Like          interface{}
	NotLike       interface{}
	In            interface{}
	NotIn         interface{}
	MoreThan      interface{}
	LessThan      interface{}
	Equal         interface{}
	NotEqual      interface{}
	MoreThanEqual interface{}
	LessThanEqual interface{}
	Regex         interface{}
	NotRegex      interface{}
	LenMore       interface{}
	LenLess       interface{}
	LenEqual      interface{}
	Contains      interface{}
	ObjectType    interface{}
	Any           interface{}
}{"${{ibridge.Operator.Null}}", "${{ibridge.Operator.Notnull}}", "${{ibridge.Operator.Like}}", "${{ibridge.Operator.Notlike}}", "${{ibridge.Operator.In}}", "${{ibridge.Operator.Notin}}", "${{ibridge.Operator.Morethan}}", "${{ibridge.Operator.Lessthan}}", "${{ibridge.Operator.Equal}}", "${{ibridge.Operator.Notequal}}", "${{ibridge.Operator.Morethanequal}}", "${{ibridge.Operator.Lessthanequal}}", "${{ibridge.Operator.Regex}}", "${{ibridge.Operator.Notregex}}", "${{ibridge.Operator.LenMore}}", "${{ibridge.Operator.LenLess}}", "${{ibridge.Operator.LenEqual}}", "${{ibridge.Operator.Contains}}", "${{ibridge.Operator.ObjectType}}", "${{ibridge.Operator.Any}}"}

var Helpers = struct {
	GenerateUUIDV4         interface{}
	GenerateUUIDV6         interface{}
	UTCTimestamp           interface{}
	LocalTimestamp         interface{}
	UTCTimestampFormat     interface{}
	LocalTimestampFormat   interface{}
	SubtractDateTime       interface{}
	GenerateString         interface{}
	GenerateReferenceId    any
	GenerateNumber         interface{}
	AnyToString            interface{}
	URLToBase64            interface{}
	StringConcat           interface{}
	UnixTimestamp          interface{}
	UnixTimestampMilli     interface{}
	UnixTimestampMicro     interface{}
	UnixTimestampNano      interface{}
	StringTrimLength       interface{}
	StringPaddingZero      interface{}
	StringPaddingSpace     interface{}
	SetDefaultString       interface{}
	Sum                    interface{}
	DecodeStringTag        interface{}
	AppendArray            interface{}
	SortArray              interface{}
	TrimSpace              interface{}
	JoinAndTrimLines       interface{}
	AddLuhnAlgorithm       interface{}
	RemoveStringElement    interface{}
	StringTrimLeft         interface{}
	StringTrimRight        interface{}
	GetImageBase64FromURL  interface{}
	StringReplaceAll       interface{}
	GetLength              interface{}
	GenerateHexadecimal    interface{}
	Multiply               interface{}
	GetEnv                 interface{}
	Encrypt                interface{}
	ConvertTimestampFormat interface{}
	ConvertToUnixTimestamp interface{}
	MathDivision           interface{}
	MathIntDiv             interface{}
	MathIntMod             interface{}
	FloatRounding          interface{}
	AddTimestamp           interface{}
	CurrencyFormatID       interface{}
	ConvertGenderTm        interface{}
	CensorAccountNumber    interface{}
	ComputeCRC32           interface{}
	ToLower                interface{}
	ToTitle                interface{}
	TrimPrefix             interface{}
	CalculateAge           interface{}
	CensorString           interface{}
	SumArray               interface{}
}{"${{ibridge.Helpers.GenerateUUIDV4}}", "${{ibridge.Helpers.GenerateUUIDV6}}", "${{ibridge.Helpers.UTCTimestamp}}", "${{ibridge.Helpers.LocalTimestamp}}", "${{ibridge.Helpers.UTCTimestampFormat}}", "${{ibridge.Helpers.LocalTimestampFormat}}", "${{ibridge.Helpers.SubtractDateTime}}", "${{ibridge.Helpers.GenerateString}}", "${{ibridge.Helpers.GenerateReferenceId}}", "${{ibridge.Helpers.GenerateNumber}}", "${{ibridge.Helpers.AnyToString}}", "${{ibridge.Helpers.URLToBase64}}", "${{ibridge.Helpers.StringConcat}}", "${{ibridge.Helpers.UnixTimestamp}}", "${{ibridge.Helpers.UnixTimestampMilli}}", "${{ibridge.Helpers.UnixTimestampMicro}}", "${{ibridge.Helpers.UnixTimestampNano}}", "${{ibridge.Helpers.StringTrimLength}}", "${{ibridge.Helpers.StringPaddingZero}}", "${{ibridge.Helpers.StringPaddingSpace}}", "${{ibridge.Helpers.SetDefaultString}}", "${{ibridge.Helpers.Sum}}", "${{ibridge.Helpers.DecodeStringTag}}", "${{ibridge.Helpers.AppendArray}}", "${{ibridge.Helpers.SortArray}}", "${{ibridge.Helpers.TrimSpace}}", "${{ibridge.Helpers.JoinAndTrimLines}}", "${{ibridge.Helpers.AddLuhnAlgorithm}}", "${{ibridge.Helpers.RemoveStringElement}}", "${{ibridge.Helpers.StringTrimLeft}}", "${{ibridge.Helpers.StringTrimRight}}", "${{ibridge.Helpers.GetImageBase64FromURL}}", "${{ibridge.Helpers.StringReplaceAll}}", "${{ibridge.Helpers.GetLength}}", "${{ibridge.Helpers.GenerateHexadecimal}}", "${{ibridge.Helpers.Multiply}}", "${{ibridge.Helpers.GetEnv}}", "${{ibridge.Helpers.Encrypt}}", "${{ibridge.Helpers.ConvertTimestampFormat}}", "${{ibridge.Helpers.ConvertToUnixTimestamp}}", "${{ibridge.Helpers.MathDivision}}", "${{ibridge.Helpers.MathIntDiv}}", "${{ibridge.Helpers.MathIntMod}}", "${{ibridge.Helpers.FloatRounding}}", "${{ibridge.Helpers.AddTimestamp}}", "${{ibridge.Helpers.CurrencyFormatID}}", "${{ibridge.Helpers.ConvertGenderTm}}", "${{ibridge.Helpers.CensorAccountNumber}}", "${{ibridge.Helpers.ComputeCRC32}}", "${{ibridge.Helpers.ToLower}}", "${{ibridge.Helpers.ToTitle}}", "${{ibridge.Helpers.TrimPrefix}}", "${{ibridge.Helpers.CalculateAge}}", "${{ibridge.Helpers.CensorString}}", "${{ibridge.Helpers.SumArray}}"}

var EndParsing = struct {
	Merge     string
	Separated string
	EndEvent  string
}{"Merged", "Separated", "${{ibridge.DataCenter}}"}

var TimeoutHandle = struct {
	End      string
	Continue string
}{"End", "Continue"}

var objectType = struct {
	Map    reflect.Kind
	Slice  reflect.Kind
	String reflect.Kind
}{reflect.Map, reflect.Slice, reflect.String}

var ObjectOperator = map[string]reflect.Kind{
	"Map":    reflect.Map,
	"Slice":  reflect.Slice,
	"String": reflect.String,
}
