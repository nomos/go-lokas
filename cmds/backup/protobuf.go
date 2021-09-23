package backup

import (
	"regexp"
)

func init(){
	RegisterCmd("proto2js",proto2js)
	RegisterCmd("proto2go",proto2go)
}

var (
	proto2js = &WrappedCmd{
		CmdString:  `
set timeout 5
spawn sudo -s
expect "*assword:"
send "$4\n"
expect "*#"
send "rm $3/$1.js\n"
expect "*#"
send "rm $3/$1_define.js\n"
expect "*#"
send "$3/$1.d.ts\n"
expect "*#"
send "pbjs -t json-module -w commonjs -o $3/$1.js $2/*.proto\n"
expect "*#"
send "pbjs -t static-module $2/*.proto  | pbts  --no-delimited -o $3/$1.d.ts --keep-case --no-comments  -\n"
expect "*#"
send "chmod a+r+w  $3/$1.js\n"
expect "*#"
send "chmod a+r+w  $3/$1.d.ts\n"
expect eof
exit
`,
		Tips:       "proto2js [package] [protoPath] [tsPath] [pass]",
		ParamsNum: 4,
		ParamsMap:  nil,
		CmdType:    Cmd_Expect,
		CmdHandler: func(output CmdOutput) *CmdResult {
			success:=regexp.MustCompile(`.+[#]\s*`).FindString(output.LastOutput()) == output.LastOutput()
			return &CmdResult{
				Outputs: output,
				Success: success,
				Results: nil,
			}
		},
	}
	proto2go = &WrappedCmd{
		CmdString:  `
protoc -I=$2/ --gofast_out=$3/ $2/*.proto
`,
		Tips:       "proto2go [package] [protoPath] [goPath]",
		ParamsNum:  3,
		ParamsMap:  nil,
		CmdType:    "",
		CmdHandler: func(output CmdOutput) *CmdResult {
			success:=output.LastOutput()==""
			return &CmdResult{
				Outputs: output,
				Success: success,
				Results: nil,
			}
		},
	}
)