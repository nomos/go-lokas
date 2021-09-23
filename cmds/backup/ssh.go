package backup

func init(){
	RegisterCmd("sshx",sshexec)
}

var (
	//远程执行命令
	sshexec  = &WrappedCmd{
		CmdString: `
ip=$1
passwd=$2
cmd=$3
/usr/bin/expect << EOF
proc remote_exec {ip passwd cmd} {
  spawn ssh root@\${ip}
  exp_internal 0
  expect {
    "*no':" { send "yes\n";exp_continue}
    "*password:" {send "${passwd}\n"}
    }
  expect "*#"
  send "${cmd}\n"
  expect "*#"
  send "exit"
  close
}
remote_exec ${ip} ${passwd} ${cmd}
EOF
`,
		Tips:      "sshx [ip] [pass] [cmd]",
		ParamsNum: 3,
		ParamsMap: nil,
	}
)