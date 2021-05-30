package protocol

import (
	"regexp"
	"strconv"
	"strings"
)

type LineType int

const (
	LINE_UNDEFINED LineType = iota
	LINE_COMMENT
	LINE_COMMENT_START
	LINE_COMMENT_END
	LINE_ANY
	LINE_EMPTY
	LINE_GO_PACKAGE
	LINE_GO_IMPORT_HEADER
	LINE_GO_IMPORT_BODY
	LINE_GO_IMPORT_SINGLELINE
	LINE_GO_INTERFACE_HEADER
	LINE_GO_PUBLIC_STRUCT_HEADER
	LINE_GO_PRIVATE_STRUCT_HEADER
	LINE_GO_STRUCT_FIELD_INHERIT
	LINE_GO_STRUCT_FIELD_PUBLIC
	LINE_GO_STRUCT_FIELD_PRIVATE
	LINE_GO_FUNC_HEADER
	LINE_GO_INIT_FUNC_HEADER
	LINE_GO_STRUCT_FUNC_HEADER
	LINE_GO_DEFINER
	LINE_GO_ENUM_DEFINER
	LINE_GO_CONST
	LINE_GO_VARIABLE
	LINE_GO_TAG_DEFINER
	LINE_GO_TAG_REGISTRY
	LINE_GO_CONST_CLOSURE_START
	LINE_GO_VAR_CLOSURE_START
	LINE_GO_ENUM_VARIABLE_IOTA
	LINE_GO_ENUM_VARIABLE
	LINE_GO_ENUM_AUTO

	LINE_CONF_TAG
	LINE_CONF_PACKAGE
	LINE_CONF_OFFSET
	LINE_CONF_RECUR
	LINE_BRACKET_END
	LINE_CLOSURE_END

	LINE_TS_IMPORT_SINGLELINE
	LINE_TS_IMPORT_CLOSURE_START
	LINE_TS_IMPORT_CLOSURE_END
	LINE_TS_IMPORT_OBJ
	LINE_TS_ENUM_CLOSURE_START
	LINE_TS_ENUM_OBJ
	LINE_TS_CLASS_DECORATOR
	LINE_TS_CLASS_HEADER
	LINE_TS_CLASS_CONSTRUCTOR_HEADER
	LINE_TS_CLASS_FIELD_PUBLIC
	LINE_TS_CLASS_FIELD_PRIVATE
	LINE_TS_CLASS_GETTER_HEADER
	LINE_TS_CLASS_SETTER_HEADER
	LINE_TS_CLASS_FUNC_HEADER
	LINE_TS_CLASS_FUNC_END

	LINE_TS_FUNC_HEADER

	LINE_TS_VAR_SINGLELINE
	LINE_TS_VAR_CLOSURE_START
	LINE_TS_VAR_ARRAY_START
	LINE_TS_ARRAY_END

	LINE_TS_DEFINE_SINGLELINE
	LINE_TS_DEFINE_START
	LINE_TS_DEFINE_OBJ
	LINE_TS_DEFINE_END

	LINE_TS_INIT_FUNC_HEADER
	LINE_TS_INIT_FUNC_END
	LINE_TS_ID_REG
	LINE_TS_PROTO_ID_REG

	LINE_MODEL_PACKAGE
	LINE_MODEL_GOPACKAGE
	LINE_MODEL_CSPACKAGE
	LINE_MODEL_TSPACKAGE
	LINE_MODEL_IMPORTS
	LINE_MODEL_IDS_HEADER
	LINE_MODEL_ID
	LINE_MODEL_CLASS_HEADER
	LINE_MODEL_CLASS_FIELD
	LINE_MODEL_ENUM_HEADER
	LINE_MODEL_ENUM_FIELD


	LINE_PROTO_PACKAGE
	LINE_PROTO_HEADER
	LINE_PROTO_FIELD
	LINE_PROTO_ENUM_HEADER
	LINE_PROTO_ENUM_FIELD
	LINE_PROTO_ENUM_ALIAS
	LINE_PROTO_SYNTAX
	LINE_PROTO_ID
)

var line_string = make(map[LineType]string)
var line_regexp_map = make(map[LineType]*regexp.Regexp)
var line_parse_map = make(map[LineType]string)
var line_regexp_replace_name = make(map[LineType]string)
var line_regexp_replace_pkg = make(map[LineType]string)
var line_regexp_replace_type = make(map[LineType]string)
var line_regexp_replace_value = make(map[LineType]string)
var line_regexp_replace_struct_name = make(map[LineType]string)
var line_regexp_replace_tag_name = make(map[LineType]string)

var COMMENT_REGEXP = regexp.MustCompile(`(.*)((//).*)`)
var ICOMPONENT_REGEXP = regexp.MustCompile(`Component`)
var ISERIALIZABLE_REGEXP = regexp.MustCompile(`Serializable`)

func GetLineRegExps()map[string]*regexp.Regexp{
	ret:=make(map[string]*regexp.Regexp)
	for k,v:=range line_regexp_map {
		name:=line_string[k]
		ret[name] = v
	}
	return ret
}

func init() {
	line_string[LINE_UNDEFINED] = "LINE_UNDEFINED"
	line_string[LINE_COMMENT] = "LINE_COMMENT"
	line_string[LINE_COMMENT_START] = "LINE_COMMENT_START"
	line_string[LINE_COMMENT_END] = "LINE_COMMENT_END"
	line_string[LINE_ANY] = "LINE_ANY"
	line_string[LINE_EMPTY] = "LINE_EMPTY"
	line_string[LINE_GO_PACKAGE] = "LINE_GO_PACKAGE"
	line_string[LINE_GO_IMPORT_HEADER] = "LINE_GO_IMPORT_HEADER"
	line_string[LINE_GO_IMPORT_BODY] = "LINE_GO_IMPORT_BODY"
	line_string[LINE_GO_IMPORT_SINGLELINE] = "LINE_GO_IMPORT_SINGLELINE"
	line_string[LINE_GO_PUBLIC_STRUCT_HEADER] = "LINE_GO_PUBLIC_STRUCT_HEADER"
	line_string[LINE_GO_PRIVATE_STRUCT_HEADER] = "LINE_GO_PRIVATE_STRUCT_HEADER"
	line_string[LINE_GO_STRUCT_FIELD_INHERIT] = "LINE_GO_STRUCT_FIELD_INHERIT"
	line_string[LINE_GO_STRUCT_FIELD_PUBLIC] = "LINE_GO_STRUCT_FIELD_PUBLIC"
	line_string[LINE_GO_STRUCT_FIELD_PRIVATE] = "LINE_GO_STRUCT_FIELD_PRIVATE"
	line_string[LINE_GO_FUNC_HEADER] = "LINE_GO_FUNC_HEADER"
	line_string[LINE_GO_INIT_FUNC_HEADER] = "LINE_GO_INIT_FUNC_HEADER"
	line_string[LINE_GO_STRUCT_FUNC_HEADER] = "LINE_GO_STRUCT_FUNC_HEADER"
	line_string[LINE_GO_DEFINER] = "LINE_GO_DEFINER"
	line_string[LINE_GO_ENUM_DEFINER] = "LINE_GO_ENUM_DEFINER"
	line_string[LINE_GO_CONST] = "LINE_GO_CONST"
	line_string[LINE_GO_VARIABLE] = "LINE_GO_VARIABLE"
	line_string[LINE_GO_TAG_DEFINER] = "LINE_GO_TAG_DEFINER"
	line_string[LINE_GO_TAG_REGISTRY] = "LINE_GO_TAG_REGISTRY"
	line_string[LINE_GO_CONST_CLOSURE_START] = "LINE_GO_CONST_CLOSURE_START"
	line_string[LINE_GO_VAR_CLOSURE_START] = "LINE_GO_VAR_CLOSURE_START"
	line_string[LINE_GO_ENUM_VARIABLE_IOTA] = "LINE_GO_ENUM_VARIABLE_IOTA"
	line_string[LINE_GO_ENUM_VARIABLE] = "LINE_GO_ENUM_VARIABLE"
	line_string[LINE_GO_ENUM_AUTO] = "LINE_GO_ENUM_AUTO"

	line_string[LINE_CONF_TAG] = "LINE_CONF_TAG"
	line_string[LINE_CONF_PACKAGE] = "LINE_CONF_PACKAGE"
	line_string[LINE_CONF_OFFSET] = "LINE_CONF_OFFSET"
	line_string[LINE_CONF_RECUR] = "LINE_CONF_RECUR"
	line_string[LINE_BRACKET_END] = "LINE_BRACKET_END"
	line_string[LINE_CLOSURE_END] = "LINE_CLOSURE_END"

	line_string[LINE_TS_IMPORT_SINGLELINE] = "LINE_TS_IMPORT_SINGLELINE"
	line_string[LINE_TS_IMPORT_CLOSURE_START] = "LINE_TS_IMPORT_CLOSURE_START"
	line_string[LINE_TS_IMPORT_CLOSURE_END] = "LINE_TS_IMPORT_CLOSURE_END"
	line_string[LINE_TS_IMPORT_OBJ] = "LINE_TS_IMPORT_OBJ"
	line_string[LINE_TS_ENUM_CLOSURE_START] = "LINE_TS_ENUM_CLOSURE_START"
	line_string[LINE_TS_ENUM_OBJ] = "LINE_TS_ENUM_OBJ"
	line_string[LINE_TS_CLASS_DECORATOR] = "LINE_TS_CLASS_DECORATOR"
	line_string[LINE_TS_CLASS_HEADER] = "LINE_TS_CLASS_HEADER"
	line_string[LINE_TS_CLASS_CONSTRUCTOR_HEADER] = "LINE_TS_CLASS_CONSTRUCTOR_HEADER"
	line_string[LINE_TS_CLASS_FIELD_PUBLIC] = "LINE_TS_CLASS_FIELD_PUBLIC"
	line_string[LINE_TS_CLASS_FIELD_PRIVATE] = "LINE_TS_CLASS_FIELD_PRIVATE"
	line_string[LINE_TS_CLASS_GETTER_HEADER] = "LINE_TS_CLASS_GETTER_HEADER"
	line_string[LINE_TS_CLASS_SETTER_HEADER] = "LINE_TS_CLASS_SETTER_HEADER"
	line_string[LINE_TS_CLASS_FUNC_HEADER] = "LINE_TS_CLASS_FUNC_HEADER"
	line_string[LINE_TS_CLASS_FUNC_END] = "LINE_TS_CLASS_FUNC_END"

	line_string[LINE_TS_FUNC_HEADER] = "LINE_TS_FUNC_HEADER"

	line_string[LINE_TS_VAR_SINGLELINE] = "LINE_TS_VAR_SINGLELINE"
	line_string[LINE_TS_VAR_CLOSURE_START] = "LINE_TS_VAR_CLOSURE_START"
	line_string[LINE_TS_VAR_ARRAY_START] = "LINE_TS_VAR_ARRAY_START"
	line_string[LINE_TS_ARRAY_END] = "LINE_TS_ARRAY_END"

	line_string[LINE_TS_DEFINE_SINGLELINE] = "LINE_TS_DEFINE_SINGLELINE"
	line_string[LINE_TS_DEFINE_START] = "LINE_TS_DEFINE_START"
	line_string[LINE_TS_DEFINE_OBJ] = "LINE_TS_DEFINE_OBJ"
	line_string[LINE_TS_DEFINE_END] = "LINE_TS_DEFINE_END"
	line_string[LINE_TS_INIT_FUNC_HEADER] = "LINE_TS_INIT_FUNC_HEADER"
	line_string[LINE_TS_INIT_FUNC_END] = "LINE_TS_INIT_FUNC_END"
	line_string[LINE_TS_ID_REG] = "LINE_TS_ID_REG"
	line_string[LINE_TS_PROTO_ID_REG] = "LINE_TS_PROTO_ID_REG"

	line_string[LINE_MODEL_PACKAGE] = "LINE_MODEL_PACKAGE"
	line_string[LINE_MODEL_CSPACKAGE] = "LINE_MODEL_CSPACKAGE"
	line_string[LINE_MODEL_GOPACKAGE] = "LINE_MODEL_GOPACKAGE"
	line_string[LINE_MODEL_TSPACKAGE] = "LINE_MODEL_TSPACKAGE"
	line_string[LINE_MODEL_IMPORTS] = "LINE_MODEL_IMPORTS"
	line_string[LINE_MODEL_IDS_HEADER] = "LINE_MODEL_IDS_HEADER"
	line_string[LINE_MODEL_ID] = "LINE_MODEL_ID"
	line_string[LINE_MODEL_CLASS_HEADER] = "LINE_MODEL_CLASS_HEADER"
	line_string[LINE_MODEL_CLASS_FIELD] = "LINE_MODEL_CLASS_FIELD"
	line_string[LINE_MODEL_ENUM_HEADER] = "LINE_MODEL_ENUM_HEADER"
	line_string[LINE_MODEL_ENUM_FIELD] = "LINE_MODEL_ENUM_FIELD"

	line_string[LINE_PROTO_PACKAGE] = "LINE_PROTO_PACKAGE"
	line_string[LINE_PROTO_HEADER] = "LINE_PROTO_HEADER"
	line_string[LINE_PROTO_FIELD] = "LINE_PROTO_FIELD"
	line_string[LINE_PROTO_ENUM_HEADER] = "LINE_PROTO_ENUM_HEADER"
	line_string[LINE_PROTO_ENUM_FIELD] = "LINE_PROTO_ENUM_FIELD"
	line_string[LINE_PROTO_ENUM_ALIAS] = "LINE_PROTO_ENUM_ALIAS"
	line_string[LINE_PROTO_SYNTAX] = "LINE_PROTO_SYNTAX"
	line_string[LINE_PROTO_ID] = "LINE_PROTO_ID"
	/*

	 */
	line_regexp_map[LINE_EMPTY] = regexp.MustCompile(`\s*`)
	line_regexp_map[LINE_COMMENT] = regexp.MustCompile(`\/\/.*`)
	line_regexp_map[LINE_COMMENT_START] = regexp.MustCompile(`\/[*].*`)
	line_regexp_map[LINE_COMMENT_END] = regexp.MustCompile(`.*[*]\/\s*`)
	line_regexp_map[LINE_ANY] = regexp.MustCompile(`.*`)
	line_regexp_map[LINE_GO_PACKAGE] = regexp.MustCompile(`(package)\s+(\w+)`)
	line_parse_map[LINE_GO_PACKAGE] = `package {$pkg}`
	line_regexp_replace_pkg[LINE_GO_PACKAGE] = "$2"
	line_regexp_map[LINE_GO_IMPORT_HEADER] = regexp.MustCompile(`(import)\s*[(]\s*`)
	line_regexp_map[LINE_GO_IMPORT_BODY] = regexp.MustCompile(`\s*(\w*)\s*(["].+["])`)
	line_regexp_map[LINE_GO_IMPORT_SINGLELINE] = regexp.MustCompile(`(import)\s*(\w*)\s*(["].+["])`)
	line_regexp_map[LINE_GO_INTERFACE_HEADER] = regexp.MustCompile(`(type)\s+\w+\s+(interface)\s+[{]\s*`)
	line_regexp_map[LINE_GO_PUBLIC_STRUCT_HEADER] = regexp.MustCompile(`(type)\s+([A-Z]\w*)\s+(struct)\s+[{]\s*`)
	line_parse_map[LINE_GO_PUBLIC_STRUCT_HEADER] = `type {$struct} struct {`
	line_regexp_replace_struct_name[LINE_GO_PUBLIC_STRUCT_HEADER] = "$2"
	line_regexp_map[LINE_GO_PRIVATE_STRUCT_HEADER] = regexp.MustCompile(`(type)\s+([a-z]\w*)\s+(struct)\s+[{]\s*`)
	line_parse_map[LINE_GO_PRIVATE_STRUCT_HEADER] = `type {$struct} struct {`
	line_regexp_replace_struct_name[LINE_GO_PRIVATE_STRUCT_HEADER] = "$2"
	line_regexp_map[LINE_GO_STRUCT_FIELD_INHERIT] = regexp.MustCompile(`\s?(chan )?[*]?(\w|[_]|[.]|\[|\]|[*])+\s*`+"(`.*`)?"+`\s*`)
	line_regexp_map[LINE_GO_STRUCT_FIELD_PUBLIC] = regexp.MustCompile(`\s?([A-Z]\w*)\s+((chan )?[*]?(\w|[_]|[.]|\[|\]|[*])+)\s*`+"(`.*`)?"+`\s*`)
	line_parse_map[LINE_GO_STRUCT_FIELD_PUBLIC] = `\t{$name} {$type}`
	line_regexp_replace_name[LINE_GO_STRUCT_FIELD_PUBLIC] = "$1"
	line_regexp_replace_type[LINE_GO_STRUCT_FIELD_PUBLIC] = "$2"
	line_regexp_map[LINE_GO_STRUCT_FIELD_PRIVATE] = regexp.MustCompile(`\s?([a-z]\w*)\s+((chan )?[*]?(\w|[_]|[.]|\[|\]|[*])+)\s*`+"(`.*`)?"+`\s*`)
	line_parse_map[LINE_GO_STRUCT_FIELD_PRIVATE] = `\t{$name} {$type}`
	line_regexp_replace_name[LINE_GO_STRUCT_FIELD_PRIVATE] = "$1"
	line_regexp_replace_type[LINE_GO_STRUCT_FIELD_PRIVATE] = "$2"
	line_regexp_map[LINE_GO_FUNC_HEADER] = regexp.MustCompile(`(func)\s+(\w+)\s*([(].*[)]).*[{]\s*`)
	line_regexp_replace_name[LINE_GO_FUNC_HEADER] = "$1"
	line_regexp_map[LINE_GO_INIT_FUNC_HEADER] = regexp.MustCompile(`(func)\s+(init)\s*(\(\s*\)).*[{]\s*`)
	line_parse_map[LINE_GO_INIT_FUNC_HEADER] = `func init() {`
	line_regexp_map[LINE_GO_STRUCT_FUNC_HEADER] = regexp.MustCompile(`(func)\s+([(]\w+\s+[*]?\w+[)])\s*(\w+)\s*([(].*[)]).*[{]\s*`)
	line_regexp_replace_struct_name[LINE_GO_STRUCT_FUNC_HEADER] = "$1"
	line_regexp_replace_name[LINE_GO_STRUCT_FUNC_HEADER] = "$2"
	line_regexp_map[LINE_GO_DEFINER] = regexp.MustCompile(`(type)\s+(\w+)\s+((\s|[)]|[(]|\w|[_]|[.]|\[|\]|[*])+)\s*`)
	line_regexp_map[LINE_GO_ENUM_DEFINER] = regexp.MustCompile(`(type)\s+([A-Z|_][A-Z|_|0-9]*)\s+(protocol.Enum)\s*`)
	line_regexp_replace_type[LINE_GO_ENUM_DEFINER] = "$2"
	line_regexp_map[LINE_GO_CONST] = regexp.MustCompile(`(const)\s+(\w+)\s*[=].+\s*`)
	line_regexp_map[LINE_GO_VARIABLE] = regexp.MustCompile(`((var)\s+(\w+).*[=].+\s*)|((var)\s+(\w+))(\s|[)]|[(]|\w|[_]|[.]|\[|\]|[*])+\s*`)
	line_regexp_map[LINE_GO_TAG_DEFINER] = regexp.MustCompile(`\s*([A-Z]+)[_]([A-Z]+(\w)*)\s+(protocol[.]BINARY[_]TAG)\s*[=]\s*([0-9]+)\s*`)
	line_regexp_replace_tag_name[LINE_GO_TAG_DEFINER] = "$1"
	line_regexp_replace_struct_name[LINE_GO_TAG_DEFINER] = "$2"
	line_regexp_replace_value[LINE_GO_TAG_DEFINER] = "$5"
	line_regexp_map[LINE_GO_TAG_REGISTRY] = regexp.MustCompile(`\s*(protocol.GetTypeRegistry\(\).RegistryType\()([A-Z]+)[_]([A-Z]+\w*)(,reflect.TypeOf\(\([*])((\w+)[.])?([A-Z]+\w*)(\)\(nil\)\).Elem\(\)\)\s*)`)
	line_parse_map[LINE_GO_TAG_REGISTRY] = "\tprotocol.GetTypeRegistry().RegistryType({$name},reflect.TypeOf((*{$type})(nil)).Elem())"
	line_regexp_replace_tag_name[LINE_GO_TAG_REGISTRY] = "$2"
	line_regexp_replace_struct_name[LINE_GO_TAG_REGISTRY] = "$3"
	line_regexp_replace_pkg[LINE_GO_TAG_REGISTRY] = "$6"
	line_regexp_replace_type[LINE_GO_TAG_REGISTRY] = "$7"
	line_regexp_map[LINE_GO_CONST_CLOSURE_START] = regexp.MustCompile(`(const)\s+[(]\s*`)
	line_regexp_map[LINE_GO_VAR_CLOSURE_START] = regexp.MustCompile(`(var)\s+[(]\s*`)
	line_regexp_map[LINE_GO_ENUM_VARIABLE_IOTA] = regexp.MustCompile(`\s*([A-Z|_][A-Z|_|0-9]*)\s+([A-Z|_][A-Z|_|0-9]*)\s+[=]\s+(iota)\s*([+]\s*([0-9]+))?\s*`)
	line_regexp_replace_name[LINE_GO_ENUM_VARIABLE_IOTA] = "$1"
	line_regexp_replace_type[LINE_GO_ENUM_VARIABLE_IOTA] = "$2"
	line_regexp_replace_value[LINE_GO_ENUM_VARIABLE_IOTA] = "$5"
	line_regexp_map[LINE_GO_ENUM_VARIABLE] = regexp.MustCompile(`\s*([A-Z|_][A-Z|_|0-9]*)\s+([A-Z|_][A-Z|_|0-9]*)\s*[=]\s*([0-9]+)\s*`)
	line_regexp_replace_name[LINE_GO_ENUM_VARIABLE] = "$1"
	line_regexp_replace_type[LINE_GO_ENUM_VARIABLE] = "$2"
	line_regexp_replace_value[LINE_GO_ENUM_VARIABLE] = "$3"
	line_regexp_map[LINE_GO_ENUM_AUTO] = regexp.MustCompile(`\s*([A-Z|_][A-Z|_|0-9]*)\s*`)
	line_regexp_replace_name[LINE_GO_ENUM_AUTO] = "$1"
	line_regexp_map[LINE_BRACKET_END] = regexp.MustCompile(`[)][\s|;]*`)
	line_regexp_map[LINE_CLOSURE_END] = regexp.MustCompile(`[}][\s|;]*`)
	line_parse_map[LINE_CLOSURE_END] = `}\n`

	line_regexp_map[LINE_CONF_TAG] = regexp.MustCompile(`(tag)\s*[=]\s*([A-Z]+)\s*`)
	line_regexp_replace_value[LINE_CONF_TAG] = "$2"
	line_regexp_map[LINE_CONF_PACKAGE] = regexp.MustCompile(`(package)\s*[=]\s*(\w+)\s*`)
	line_regexp_replace_value[LINE_CONF_PACKAGE] = "$2"
	line_regexp_map[LINE_CONF_OFFSET] = regexp.MustCompile(`(offset)\s*[=]\s*([0-9]+)\s*`)
	line_regexp_replace_value[LINE_CONF_OFFSET] = "$2"
	line_regexp_map[LINE_CONF_RECUR] = regexp.MustCompile(`(recursive)\s*[=]\s*(\w+)\s*`)
	line_regexp_replace_value[LINE_CONF_RECUR] = "$2"

	line_regexp_map[LINE_TS_IMPORT_SINGLELINE] = regexp.MustCompile(`\s*import\s+([*]\s*as\s*)?([{]?[,|\w|\s]+[}]?\s*from\s*)?"[\w+|.|/]+"[;|\s]*`)
	line_regexp_map[LINE_TS_IMPORT_CLOSURE_START] = regexp.MustCompile(`\s*import\s+[{]\s*\s*`)
	line_regexp_map[LINE_TS_IMPORT_OBJ] = regexp.MustCompile(`\s*[\w|\s]+[,]?\s*`)
	line_regexp_map[LINE_TS_IMPORT_CLOSURE_END] = regexp.MustCompile(`[}]\s*from\s*"[\w+|.|/]+"[;|\s]*`)

	line_regexp_map[LINE_TS_ENUM_CLOSURE_START] = regexp.MustCompile(`(export)?\s*enum\s*[\w+|_]+\s*[{]\s*`)
	line_regexp_map[LINE_TS_ENUM_OBJ] = regexp.MustCompile(`\s*[\w+|_]+\s*([=]\s*[0-9]+)?\s*[,]?\s*`)

	line_regexp_map[LINE_TS_CLASS_DECORATOR] = regexp.MustCompile(`\s*@\w+\s*`)
	line_regexp_map[LINE_TS_CLASS_HEADER] = regexp.MustCompile(`(export)?\s*(default)?\s*(class)\s+([A-Z]\w*)\s+(extends\s+([A-Z]\w*)){0,1}\s*[{]\s*`)
	line_regexp_replace_struct_name[LINE_TS_CLASS_HEADER] = "$4"
	line_regexp_map[LINE_TS_CLASS_CONSTRUCTOR_HEADER] = regexp.MustCompile(`\s*constructor\s*\(.*\).*\{\s*`)
	line_regexp_map[LINE_TS_CLASS_FIELD_PUBLIC] = regexp.MustCompile(`\s+(public)\s+(\w+)\s*[:]\s*((\w|[.]|\[|\]|<|>|[,])+)(\s*[=]\s*.+)?\s*`)
	line_regexp_replace_name[LINE_TS_CLASS_FIELD_PUBLIC] = "$2"
	line_regexp_replace_type[LINE_TS_CLASS_FIELD_PUBLIC] = "$3"
	line_regexp_map[LINE_TS_CLASS_FIELD_PRIVATE] = regexp.MustCompile(`\s+(private)\s+(\w+)\s*[:]\s*((\w|[.]|\[|\]|<|>|[,])+)(\s*[=]\s*.+)?\s*`)
	line_regexp_replace_name[LINE_TS_CLASS_FIELD_PRIVATE] = "$2"
	line_regexp_replace_type[LINE_TS_CLASS_FIELD_PRIVATE] = "$3"
	line_regexp_map[LINE_TS_CLASS_GETTER_HEADER] = regexp.MustCompile(`\s+get\s(\w+)\s*\(.*\).*\{\s*`)
	line_regexp_replace_name[LINE_TS_CLASS_GETTER_HEADER] = "$1"
	line_regexp_map[LINE_TS_CLASS_SETTER_HEADER] = regexp.MustCompile(`\s+set\s(\w+)\s*\(.*\).*\{\s*`)
	line_regexp_replace_name[LINE_TS_CLASS_SETTER_HEADER] = "$1"
	line_regexp_map[LINE_TS_CLASS_FUNC_HEADER] = regexp.MustCompile(`\s+(async\s+)?\w+\s*[(]\s*(\w+\s*([:]\s*\w+)?[,]?)*\s*[)]\s*[{]\s*`)
	line_regexp_map[LINE_TS_CLASS_FUNC_END] = regexp.MustCompile(`\s+[}]\s*`)

	line_regexp_map[LINE_TS_FUNC_HEADER] = regexp.MustCompile(`(export)?\s*function\s+\w+\s*\(.*\).*\{\s*`)

	line_regexp_map[LINE_TS_VAR_SINGLELINE] = regexp.MustCompile(`(var|let|const)\s+\w+\s*[=].*[^[{|\[]]`)
	line_regexp_map[LINE_TS_VAR_CLOSURE_START] = regexp.MustCompile(`(var|let|const)\s+\w+\s*[=].*\{\s*`)
	line_regexp_map[LINE_TS_VAR_ARRAY_START] = regexp.MustCompile(`(var|let|const)\s+\w+\s*[=].*\[\s*`)
	line_regexp_map[LINE_TS_ARRAY_END] = regexp.MustCompile(`\][;\s]*`)

	line_regexp_map[LINE_TS_DEFINE_SINGLELINE] = regexp.MustCompile(`[@]define\((\s*["|']\w+["|'])([,]\s*\[\]([,]\s*["|']\s*(\w+)\s*["|']\s*)*)?\)\s*`)
	line_regexp_map[LINE_TS_DEFINE_START] = regexp.MustCompile(`[@]define\((\s*["|']\w+["|'])([,]\s*\[)?\s*`)
	line_regexp_map[LINE_TS_DEFINE_OBJ] = regexp.MustCompile(`\s*\[\s*"\s*(\w+)\s*"\s*(,(\w|[.])+)+\]\s*[,]\s*`)
	line_regexp_replace_name[LINE_TS_DEFINE_OBJ] = "$1"
	line_regexp_map[LINE_TS_DEFINE_END] = regexp.MustCompile(`\]\s*([,]["]\s*\w+\s*["])*\s*\)\s*`)
	line_regexp_map[LINE_TS_INIT_FUNC_HEADER] = regexp.MustCompile(`\(\s*function\s*\(\s*\)\s*\{\s*`)
	line_regexp_map[LINE_TS_INIT_FUNC_END] = regexp.MustCompile(`\}\)\(\)\s*`)
	line_regexp_map[LINE_TS_ID_REG] = regexp.MustCompile(`\s*TypeRegistry[.]getInstance\(\)[.]RegisterCustomTag\(\s*"\w+"\s*[,]\s*[0-9]+\s*\)[;|\s]*`)
	line_regexp_map[LINE_TS_PROTO_ID_REG] = regexp.MustCompile(`\s*TypeRegistry[.]getInstance\(\)[.]RegisterProtoTag\(\s*([.]|\w)+\s*[,]\s*[0-9]+\s*\)[;|\s]*`)

	line_regexp_map[LINE_MODEL_PACKAGE] = regexp.MustCompile(`\s*package\s+(\w+)\s*`)
	line_regexp_replace_pkg[LINE_MODEL_PACKAGE] = "$1"
	line_regexp_map[LINE_MODEL_GOPACKAGE] = regexp.MustCompile(`\s*go[-]package\s+(\w+)\s*`)
	line_regexp_replace_pkg[LINE_MODEL_GOPACKAGE] = "$1"
	line_regexp_map[LINE_MODEL_CSPACKAGE] = regexp.MustCompile(`\s*cs[-]package\s+((\w|[.])+)\s*`)
	line_regexp_replace_pkg[LINE_MODEL_CSPACKAGE] = "$1"
	line_regexp_map[LINE_MODEL_TSPACKAGE] = regexp.MustCompile(`\s*ts[-]package\s+(\w+)\s*`)
	line_regexp_replace_pkg[LINE_MODEL_TSPACKAGE] = "$1"
	line_regexp_map[LINE_MODEL_IMPORTS] = regexp.MustCompile(`\s*import\s+(\w+)\s*`)
	line_regexp_replace_pkg[LINE_MODEL_IMPORTS] = "$1"
	line_regexp_map[LINE_MODEL_IDS_HEADER] = regexp.MustCompile(`\s*[\[]\s*(ids)\s*[\]]\s*`)
	line_regexp_map[LINE_MODEL_ID] = regexp.MustCompile(`\s*(\w+)\s+([0-9]+)((\s*((REQ)|(NTF)|(EVT)))*)(\s*(\w+))*\s*`)
	line_regexp_replace_name[LINE_MODEL_ID] = "$1"
	line_regexp_replace_value[LINE_MODEL_ID] = "$2"
	line_regexp_replace_type[LINE_MODEL_ID] = "$3"
	line_regexp_replace_tag_name[LINE_MODEL_ID] = "$9"
	line_regexp_map[LINE_MODEL_CLASS_HEADER] = regexp.MustCompile(`\s*[\[]\s*class\s+(\w+)\s*[\]]\s*`)
	line_regexp_replace_struct_name[LINE_MODEL_CLASS_HEADER] = "$1"
	line_regexp_map[LINE_MODEL_ENUM_HEADER] = regexp.MustCompile(`\s*[\[]\s*enum\s+(\w+)\s*[\]]\s*`)
	line_regexp_replace_struct_name[LINE_MODEL_ENUM_HEADER] = "$1"
	line_regexp_replace_name[LINE_MODEL_ENUM_HEADER] = "$1"
	line_regexp_replace_name[LINE_MODEL_CLASS_HEADER] = "$1"
	line_regexp_map[LINE_MODEL_CLASS_FIELD] = regexp.MustCompile(`\s*(\w+)\s+((\w|[{]|[]]|[}]|[[]|[:])+)\s*`)
	line_regexp_replace_name[LINE_MODEL_CLASS_FIELD] = "$1"
	line_regexp_replace_type[LINE_MODEL_CLASS_FIELD] = "$2"
	line_regexp_map[LINE_MODEL_ENUM_FIELD] = regexp.MustCompile(`\s*(\w+)\s+([0-9]+)\s*`)
	line_regexp_replace_name[LINE_MODEL_ENUM_FIELD] = "$1"
	line_regexp_replace_value[LINE_MODEL_ENUM_FIELD] = "$2"

	line_regexp_map[LINE_PROTO_FIELD] = regexp.MustCompile(`\s+((map[<]\s*(\w+)\s*[,]\s*(\w+)\s*[>])|(((repeated)(\s+)))?(\w+))\s+(\w+)\s*[=]\s*([0-9]+)[;]?\s*`)
	line_regexp_map[LINE_PROTO_ENUM_FIELD] = regexp.MustCompile(`\s+(\w+)\s*[=]\s*([0-9]+)[;]?\s*`)
	line_regexp_replace_name[LINE_PROTO_ENUM_FIELD] = "$1"
	line_regexp_replace_value[LINE_TS_DEFINE_OBJ] = "$2"
	line_regexp_map[LINE_PROTO_HEADER] = regexp.MustCompile(`\s*message\s+([A-Z]\w+)\s*{\s*`)
	line_regexp_replace_name[LINE_PROTO_HEADER] = "$1"
	line_regexp_map[LINE_PROTO_ENUM_HEADER] = regexp.MustCompile(`\s*enum\s+([A-Z]\w+)\s*{\s*`)
	line_regexp_map[LINE_PROTO_PACKAGE] = regexp.MustCompile(`\s*package\s*(\w+)[;]?\s*`)
	line_regexp_map[LINE_PROTO_SYNTAX] = regexp.MustCompile(`\s*syntax\s*=\s*["]proto3["][;]?\s*`)
	line_regexp_map[LINE_PROTO_ENUM_ALIAS] = regexp.MustCompile(`\s+option\s+allow_alias\s*=\s*true[;]\s*`)
	line_regexp_map[LINE_PROTO_ID] = regexp.MustCompile(`\s*([0-9]+)\s+(\w+)\s*`)
	line_regexp_replace_name[LINE_PROTO_ENUM_FIELD] = "$2"
	line_regexp_replace_value[LINE_TS_DEFINE_OBJ] = "$1"
}

type LineText struct {
	Obj         GeneratorObject
	LineNum     int
	Text        string
	LineType    LineType
	PackageName string
	StructName  string
	Name        string
	TagName     string
	Value       int
	Type        string
}

type Fields map[string]string

func (this *LineText) ObjName() string {
	if this.Obj != nil {
		return this.Obj.ObjectType().String()
	}
	return "Unknown"
}

func (this *LineText) GetStructName() string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	this.StructName = this.LineType.RegReplaceStructName(removeComment)
	return this.StructName
}

func (this *LineText) GetPkgName() string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	this.PackageName = this.LineType.RegReplacePkg(removeComment)
	return this.PackageName
}

func (this *LineText) GetValue() int {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	this.Value, _ = strconv.Atoi(this.LineType.RegReplaceValue(removeComment))
	return this.Value
}

func (this *LineText) GetTypeName() string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	this.Type = this.LineType.RegReplaceType(removeComment)
	return this.Type
}

func (this *LineText) GetName() string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	this.Name = this.LineType.RegReplaceName(removeComment)
	this.Name = strings.TrimLeft(this.Name," ")
	this.Name = strings.TrimRight(this.Name," ")
	return this.Name
}

func (this *LineText) GetTagName() string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	this.TagName = this.LineType.RegReplaceTagName(removeComment)
	return this.TagName
}

func (this *LineText) IsLongStringTag() bool {
	removeComment := COMMENT_REGEXP.ReplaceAllString(this.Text, "$1")
	return regexp.MustCompile(`Tag[.]LongString`).MatchString(removeComment)
}

func (this *LineText) Parse() string {
	this.Text = this.Text
	switch this.LineType {
	case LINE_GO_PACKAGE:
		this.Text = this.LineType.Parse(this.Text, Fields{"pkg": this.PackageName})
	case LINE_GO_PUBLIC_STRUCT_HEADER:
		this.Text = this.LineType.Parse(this.Text, Fields{"struct": this.StructName})
	case LINE_GO_PRIVATE_STRUCT_HEADER:
		this.Text = this.LineType.Parse(this.Text, Fields{"struct": this.StructName})
	case LINE_GO_STRUCT_FIELD_PUBLIC:
		this.Text = this.LineType.Parse(this.Text, Fields{"name": this.Name, "type": this.PackageName + "." + this.Type})
	case LINE_GO_STRUCT_FIELD_PRIVATE:
		this.Text = this.LineType.Parse(this.Text, Fields{"name": this.Name, "type": this.PackageName + "." + this.Type})
	case LINE_GO_INIT_FUNC_HEADER:
	case LINE_GO_TAG_REGISTRY:
		this.Text = this.LineType.Parse(this.Text, Fields{"name": this.Name, "type": this.PackageName + "." + this.Type})
	}
	return this.Text
}

func (this LineType) String() string {
	return line_string[this]
}

func (this LineType) RegExp() *regexp.Regexp {
	return line_regexp_map[this]
}

func (this LineType) RegMatch(str string) bool {
	if this == LINE_EMPTY {
		return line_regexp_map[this].FindString(str) == str
	}
	return line_regexp_map[this].FindString(str) == str && str != ""
}

func (this LineType) Parse(origin string, args map[string]string) string {
	replacer, ok := line_parse_map[this]
	if !ok {
		return ""
	}
	for k, v := range args {
		replacer = strings.Replace(replacer, "{$"+k+"}", v, -1)
	}
	return replacer
}

func (this LineType) RegReplaceName(str string) string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(str, "$1")
	return this.RegExp().ReplaceAllString(removeComment, line_regexp_replace_name[this])
}

func (this LineType) RegReplacePkg(str string) string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(str, "$1")
	return this.RegExp().ReplaceAllString(removeComment, line_regexp_replace_pkg[this])
}

func (this LineType) RegReplaceType(str string) string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(str, "$1")
	return this.RegExp().ReplaceAllString(removeComment, line_regexp_replace_type[this])
}

func (this LineType) RegReplaceValue(str string) string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(str, "$1")
	return this.RegExp().ReplaceAllString(removeComment, line_regexp_replace_value[this])
}

func (this LineType) RegReplaceStructName(str string) string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(str, "$1")
	return this.RegExp().ReplaceAllString(removeComment, line_regexp_replace_struct_name[this])
}

func (this LineType) RegReplaceTagName(str string) string {
	removeComment := COMMENT_REGEXP.ReplaceAllString(str, "$1")
	return this.RegExp().ReplaceAllString(removeComment, line_regexp_replace_tag_name[this])
}