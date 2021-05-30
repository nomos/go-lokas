package shell

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas/util/stringutil"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Options contains ShellCommand configuration
type Options struct {
	CheckPath bool
	UseShell  bool
	Timeout   int
}

func getPATH() []string {
	return strings.Split(os.Getenv("PATH"), ":")
}

// this regex tests if the cmd starts with: ./, ../, ~/ or /
var partialPathRegex = regexp.MustCompile(`^((\~|\.{1,})?\/)`)

func findInPath(cmd string) (path string, found bool) {

	// stops validation when a full or
	// partial path was inputed
	log.Warnf("findInPath", partialPathRegex.Match([]byte(cmd)))
	if partialPathRegex.Match([]byte(cmd)) {
		found = true
		return
	}

	for _, dir := range getPATH() {

		fullPath := fmt.Sprintf("%s/%s", dir, cmd)
		log.Warnf("fullPath", fullPath)
		if fileExist(fullPath) {
			log.Warnf("fullPath exist", fullPath)
			path = fullPath
			found = true
			break
		}
	}

	return
}

func fileExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

type CommandLine struct {
	Command string
	Args    []string
}

func NewCommandLine(s string) *CommandLine {
	ret := &CommandLine{
		Command: "",
		Args:    nil,
	}
	ret.Parse(s)
	return ret
}

func (this *CommandLine) String() string {
	str := strings.TrimRight(strings.Join([]string{this.Command, strings.Join(this.Args, " ")}, " "), " ")
	//str = "bash -c "+str
	return str
}

func (this *CommandLine) Parse(s string) {
	splits := strings.Split(s, " ")
	this.Command = splits[0]
	if len(splits) > 1 {
		this.Args = splits[1:]
	}
	if this.Command == "" {
		log.Error("Missing command name")
		return
	}
	//path, success := findInPath(this.Command)
	//if !success {
	//	log.Error("ShellCommand not found in PATH")
	//	return
	//}
	//log.Warnf("before", this.Command, "args", this.Args)
	//this.Command = path

	//if regexp.MustCompile(`\/bin\/([a-z]+)\s*`).FindString(this.Command) == this.Command {
	//	this.Command = regexp.MustCompile(`\/bin\/([a-z]+)`).ReplaceAllString(this.Command, "$1")
	//}
	log.Warnf(this.Command, "args", this.Args)
}

// ShellCommand defines how to call a program
type ShellCommand struct {
	isExpect bool
	CommandStr  string
	Commands    []*CommandLine
	Options     Options
	Description string
	Path        string
	cmd         *exec.Cmd
	timer       *time.Timer
	writer      io.Writer
	writeChan   chan string
	stdIn  io.Reader
	outputs []string
	outputChan chan string
	done chan struct{}
}

// New creates a ShellCommand
func New(useShell bool, cmd string,isExpect bool) *ShellCommand {
	ret := &ShellCommand{
		Options:  Options{UseShell: useShell, CheckPath: true},
		Commands: []*CommandLine{},
		outputs:make([]string,0),
		outputChan:make(chan string),
		done:make(chan struct{}),
		isExpect: isExpect,
	}
	ret.cmd = exec.Command("pwd")
	data, _ := ret.cmd.Output()
	ret.Path = strings.TrimRight(string(data), "\n")
	ret.CommandStr = cmd
	log.Warnf("pwd", ret.Path)
	ret.start()
	return ret
}

func (this *ShellCommand)start() {
	go func() {
		for {
			select {
			case msg:=<-this.outputChan:
				this.outputs = append(this.outputs, msg)
			case <-this.done:
				break
			}
		}
	}()
}

func (this *ShellCommand)destroy() {
	this.done<- struct{}{}
}

func (this *ShellCommand) SetWriter(writer io.Writer) {
	this.writer = writer
}

func (this *ShellCommand) SetStdIn(reader io.Reader) {
	this.stdIn = reader
}

func (this *ShellCommand) GetOutputs()[]string {
	return this.outputs
}

func (this *ShellCommand) Set(cmd string) {
	this.CommandStr = cmd
	this.Commands = make([]*CommandLine, 0)
}

// Run a ShellCommand
func (this *ShellCommand) Start() error {
	//if this.Options.CheckPath&&!path2.IsAbs(this.Command) {
	err := this.Parse()
	//	if  err != nil {
	//		log.Error(err.Error())
	//		return fmt.Errorf("Check PATH failed: %v", err)
	//	}
	//	this.Command = path
	//}
	if err!=nil {
		log.Error(err.Error())
		return err
	}
	this.cmd.Stdin = this.stdIn
	stdPipe, err := this.cmd.StdoutPipe()
	stdErrPipe, err := this.cmd.StderrPipe()
	go func() {
		reader := bufio.NewReader(stdErrPipe)
		for {
			line, err2 := reader.ReadString('\n')
			line = strings.TrimRight(line, "\n")
			if line == "" {
				continue
			}
			if this.outputChan!=nil {
				this.outputChan<-stringutil.CopyString(line)
			}
			if this.writer != nil {
				this.writer.Write([]byte(line))
			}
			if err2 != nil || io.EOF == err2 {
				log.Error(err2.Error())
				break
			}
		}
	}()
	go func() {
		reader := bufio.NewReader(stdPipe)
		for {
			line, err2 := reader.ReadString('\n')
			line = strings.TrimRight(line, "\n")
			if line == "" {
				continue
			}
			if this.outputChan!=nil {
				this.outputChan<-stringutil.CopyString(line)
			}
			if this.writer != nil {
				this.writer.Write([]byte(line))
			}
			if err2 != nil || io.EOF == err2 {
				log.Error(err2.Error())
				break
			}
		}
	}()
	err = this.cmd.Start()
	if err != nil {
		err = fmt.Errorf("Error starting a command: %v", err)
		return err
	}

	if this.Options.Timeout > 0 {

		execLimit := time.Duration(this.Options.Timeout) * time.Second

		this.timer = time.AfterFunc(execLimit, func() {
			this.cmd.Process.Kill()
		})
	}
	return nil
}

func (this *ShellCommand) Wait() error {
	err := this.cmd.Wait()
	if err != nil {
		err = fmt.Errorf("Error running a command: %v", err)
	}

	if this.Options.Timeout > 0 {
		this.timer.Stop()
	}
	this.destroy()
	return nil
}

func (this *ShellCommand) Parse() error {
	splits := strings.Split(this.CommandStr, ";")
	for _, split := range splits {
		this.Commands = append(this.Commands, NewCommandLine(split))
	}
	if len(this.Commands)==0 {
		return errors.New("no command input")
	} else {
		cmdprefix:="bash"
		if this.isExpect {
			cmdprefix = "/usr/bin/expect"
		}
		this.cmd = exec.Command(cmdprefix, "-c",this.String())
		//this.cmd = exec.Command(this.Commands[0].String())
		//if len(this.Commands)>1 {
		//	log.Warnf(">1")
		//	stdin, _ := this.cmd.StdinPipe()
		//	for _, cmd := range this.Commands[1:] {
		//		cmdstr1 := cmd.String()
		//		log.Warnf(">1",cmdstr1)
		//		stdin.Write([]byte(cmdstr1))
		//	}
		//}
	}
	return nil
}

func (this *ShellCommand) String()string {
	arr:=make([]string,0)
	for _,cmd:=range this.Commands{
		arr = append(arr, cmd.String())
	}
	ret:= strings.Join(arr,";")
	log.Warnf("finalstr",ret)
	return ret
}