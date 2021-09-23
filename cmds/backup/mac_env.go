package backup

import (
	"regexp"
)

func init(){
	RegisterCmd("dockerv",setup_brew_mac)
	RegisterCmd("sshmongo",sshmongo)
	RegisterCmd("upload",upload_mac)
	RegisterCmd("instbrew",instbrew)
}

var (
	instbrew = &WrappedCmd{
		CmdString:  `
`,
		Tips:       "",
		ParamsNum:  0,
		ParamsMap:  nil,
		CmdType:    Cmd_Shell,
		CmdHandler: nil,
	}
	sshmongo = &WrappedCmd{
		CmdString:  `
ssh -NCPf root@47.103.151.11 -L 3307:127.0.0.1:3306
ssh -L 3306:47.103.151.11:3306 -Nf root@47.103.151.11
spawn ssh -L 27017:139.196.98.237:27017 -Nf root@101.132.188.236
expect {
 "(yes/no)?" {send "yes\n"}
"*assword:" 
{send "$1\n"}
}
expect eof
`,
		Tips:       "",
		ParamsNum:  1,
		ParamsMap:  nil,
		CmdType:    Cmd_Expect,
		CmdHandler: func(output CmdOutput) *CmdResult {
			ret:=&CmdResult{
				Outputs: output,
				Success: false,
				Results: nil,
			}
			return ret
		},
	}
	setup_brew_mac = &WrappedCmd{
		CmdString: `docker -v`,
		Tips:      "dockerv",
		ParamsNum: 0,
		ParamsMap: nil,
		CmdType: Cmd_Shell,
		CmdHandler: func(output CmdOutput) *CmdResult {
			regexp1:=regexp.MustCompile(`Docker\s*version\s*(([.]|[0-9]|\w)+)[,]\s*build\s(\w+)\s*`)
			success:=regexp1.MatchString(output.LastOutput())
			ret:= &CmdResult{
				Outputs: output,
				Success: success,
				Results: make(map[string]interface{}),
			}
			if success {
				ret.Results["version"] = regexp1.ReplaceAllString(output.LastOutput(),"$1")
				ret.Results["build"] = regexp1.ReplaceAllString(output.LastOutput(),"$3")
			}
			return ret
		},
	}
	upload_mac = &WrappedCmd{
		CmdString: `
set PASSWORD Mima9943
set DOMAIN root@101.132.188.236

set localpath /data/muolserver/gopath/src/gitlab.18178.net/muolserver/
set remotepath $DOMAIN:/root/nginx/html/
spawn bash -c "scp -C -r $localpath $remotepath"
expect {
 "(yes/no)?" {send "yes\n"}
"*assword:" 
{send "$PASSWORD\n"}
}
expect eof

set localpath /data/muolserver/gopath/src/github.com/nomos/
set remotepath $DOMAIN:/root/nginx/html/
spawn bash -c "scp -C -r $localpath $remotepath"
expect {
 "(yes/no)?" {send "yes\n"}
"*assword:" 
{send "$PASSWORD\n"}
}
expect eof
`,
		Tips:      "dockerv",
		ParamsNum: 0,
		ParamsMap: nil,
		CmdType: Cmd_Expect,
		CmdHandler: func(output CmdOutput) *CmdResult {
			success:=regexp.MustCompile(`Docker`).MatchString(output.LastOutput())
			return &CmdResult{
				Outputs: output,
				Success: success,
				Results: make(map[string]interface{}),
			}
		},
	}
)