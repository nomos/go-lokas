package sshc

import (
	"bufio"
	"fmt"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/shell"
	"github.com/nomos/go-lokas/util/events"
	"github.com/nomos/go-lokas/util/gzip"
	"github.com/nomos/go-lokas/util/promise"
	"github.com/nomos/go-lokas/util/zip"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ShellSession struct {
	ssh       *SshClient
	cmd       *shell.ShellCommand
	writer    io.Writer
	sysWriter io.Writer
}

func (this *ShellSession) SetWriter(writer io.Writer) {
	this.writer = writer
}

func (this *ShellSession) SetCmd(cmd *shell.ShellCommand) {
	this.cmd = cmd
}

func (this *ShellSession) Run(cmd string, isExpect bool) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		if this.cmd == nil {
			this.cmd = shell.New(true, cmd, isExpect)
			this.cmd.SetWriter(this.ssh)
		} else {
			this.cmd.Set(cmd)
		}
		err := this.Start()
		if err != nil {
			log.Error(err.Error())
			reject(err)
			return
		}
		resolve(this.cmd.GetOutputs())
	})
}

func (this *ShellSession) Start() error {
	if this.cmd != nil {
		err := this.cmd.Start()
		if err != nil {
			this.ssh.Error(err.Error())
			return err
		}

		err = this.cmd.Wait()
		if err != nil {
			this.ssh.Error(err.Error())
			return err
		}
		this.ssh.Write([]byte(">"))
	}
	return nil
}

type StringWriter interface {
	WriteString(string)
}

type SshClient struct {
	*log.ComposeLogger
	events.EventEmmiter
	client              *ssh.Client
	sftp                *sftp.Client
	sftps               []*sftp.Client
	sessions            []*ssh.Session
	shellSessions       []*ShellSession
	addr                string
	User                string
	password            string
	defaultShellSession *ShellSession
	defaultSession      *ssh.Session
	done                chan struct{}
	connected           bool
	sftpMutex           sync.Mutex
	stringWriter        StringWriter
}

func NewSshClient(user, password, addr string, console bool) *SshClient {
	ret := &SshClient{
		EventEmmiter: events.New(),
		addr:         addr,
		User:         user,
		password:     password,
		connected:    false,
		sftps:        make([]*sftp.Client, 0),
		sessions:     make([]*ssh.Session, 0),
	}
	ret.defaultShellSession = ret.NewShellSession()
	if console {
		ret.ComposeLogger = log.NewComposeLogger(true, log.ConsoleConfig(""), 1)
	} else {
		ret.ComposeLogger = log.NewComposeLogger(true, log.DefaultConfig(""), 1)
	}
	return ret
}

func (this *SshClient) NewShellSession() *ShellSession {
	ret := &ShellSession{
		ssh:       this,
		cmd:       nil,
		sysWriter: this,
	}
	this.shellSessions = append(this.shellSessions, ret)
	return ret
}

func (this *SshClient) Clear() {

}

func (this *SshClient) Write(p []byte) (int, error) {
	this.Info(string(p))
	this.stringWriter.WriteString(string(p))
	return 0, nil
}

func (this *SshClient) SetStringWriter(writer StringWriter) {
	this.stringWriter = writer
}

func (this *SshClient) SetAddr(user, password, addr string) {
	this.addr = addr
	this.User = user
	this.password = password
}

func (this *SshClient) GetConnStr() string {
	return this.User + "@" + this.addr
}

func (this *SshClient) ClearSftpClients() {
	for _, c := range this.sftps {
		c.Close()
	}
	this.sftps = make([]*sftp.Client, 0)
}

func (this *SshClient) fetchSftp() (*sftp.Client, error) {
	this.sftpMutex.Lock()
	defer this.sftpMutex.Unlock()
	if len(this.sftps) == 0 {
		ret, err := this.createSftp()

		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	ret := this.sftps[len(this.sftps)-1]
	this.sftps = this.sftps[:len(this.sftps)-1]
	return ret, nil
}

func (this *SshClient) createSftp() (*sftp.Client, error) {
	ret, err := sftp.NewClient(this.client)
	if err != nil {
		return nil, err
	}
	return ret, nil

}

func (this *SshClient) recycleSftp(client *sftp.Client) {
	this.sftpMutex.Lock()
	defer this.sftpMutex.Unlock()
	this.sftps = append(this.sftps, client)
}

func (this *SshClient) Connect() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		var err error
		this.client, err = ssh.Dial("tcp", this.addr, &ssh.ClientConfig{
			User:            this.User,
			Auth:            []ssh.AuthMethod{ssh.Password(this.password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
		if err != nil {
			reject(err)
			return
		}
		this.defaultSession, err = this.client.NewSession()
		if err != nil {
			reject(err)
			return
		}
		this.sftp, err = this.createSftp()
		if err != nil {
			reject(err)
			return
		}
		this.connected = true
		this.ComposeLogger.Info("Connected")
		resolve(nil)
	})
}

func (this *SshClient) IsConnect() bool {
	return this.connected
}

func (this *SshClient) Disconnect() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		if this.client == nil {
			resolve(nil)
			return
		}
		err := this.defaultSession.Close()
		if err != nil {
			log.Error(err.Error())
			reject(err)
			return
		}
		for _, session := range this.sessions {
			log.Warnf("CLOSE Session")
			if session != nil {
				session.Close()
			}
		}
		this.sessions = make([]*ssh.Session, 0)
		err = this.client.Close()
		if err != nil {
			reject(err)
			return
		}
		this.client = nil
		resolve(nil)
	})
}

func (this *SshClient) runShellCmd(cmd string, expect bool) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		go func() {
			_, err := this.defaultShellSession.Run(cmd, expect).Await()
			if err != nil {
				reject(err)
				return
			}
			resolve(this.defaultShellSession)
		}()
	})
}

func (this *SshClient) RunShellCmd(cmd string, expect bool) *promise.Promise {
	return this.runShellCmd(cmd, expect)
}

func (this *SshClient) RunShellCmdPwd() *promise.Promise {
	return this.runShellCmd("pwd", false)
}

func (this *SshClient) runCmd(cmd string, pwd bool) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		session, err := this.client.NewSession()
		if err != nil {
			reject(err)
			return
		}
		this.sessions = append(this.sessions, session)
		if pwd {
			this.readPump(session, "pwd", true)
			err = session.Run("pwd")
			if err != nil {
				reject(err)
				return
			}
			resolve(nil)
		} else {
			log.Warnf("runCmd", zap.String("cmd", cmd))
			this.readPump(session, cmd, false)
			err = session.Run(cmd)
			if err != nil {
				reject(err)
				return
			}
			resolve(session)
		}
	})
}

func (this *SshClient) RunCmd(cmd string) *promise.Promise {
	return this.runCmd(cmd, false).Then(func(interface{}) interface{} {
		return this.RunCmdPwd()
	})
}

func (this *SshClient) RunCmdPwd() *promise.Promise {
	return this.runCmd("", true)
}

func (this *SshClient) close() {
	this.Disconnect()
}

func (this *SshClient) serveIO() {
	this.done = make(chan struct{})
	this.defaultSession.Wait()
}

func (this *SshClient) readPump(session *ssh.Session, cmd string, pwd bool) {
	cmdReader, err := session.StdoutPipe()
	if err != nil {
		log.Error(err.Error())
		return
	}
	scanner := bufio.NewScanner(cmdReader)
	ticker := time.NewTicker(time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				if ok := scanner.Scan(); ok {
					text := scanner.Text()
					if pwd {
						this.WriteConsole(zapcore.Entry{}, []byte(text))
					} else {
						this.WriteConsole(zapcore.Entry{}, []byte(text))
					}
				} else {
					log.Info("onSessionClose", zap.Any("session", session))
					return
				}
			}
		}
	}()
}

func (this *SshClient) Upload(localPath string, remotePath string) error {
	s, err := os.Stat(localPath)
	if err != nil {
		text := "文件路径不存在"
		this.Error(text)
		return err
	}
	//判断是否是文件夹
	if s.IsDir() {
		err := this.uploadDirectory(localPath, remotePath)
		if err != nil {
			return err
		}
	} else {
		err := this.uploadFile(localPath, remotePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *SshClient) Remove(remotePath string) error {

	f, err := this.sftp.Stat(remotePath)
	if err != nil {
		this.Error(remotePath + ":原始文件不存在,无需删除")
		return err
	}
	if f.IsDir() {
		err := this.removeDirectory(remotePath)
		if err != nil {
			return err
		}
	} else {
		err := this.removeFile(remotePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *SshClient) UploadSub(localPath string, remotePath string) error {

	s, err := os.Stat(localPath)
	if err != nil {
		this.Error(remotePath + ":文件路径不存在")
		return err
	}
	//判断是否是文件夹
	if s.IsDir() {
		localFiles, err := ioutil.ReadDir(localPath)
		if err != nil {
			this.Error(fmt.Sprintf("读取本地文件错误:%s, %s", localPath, err.Error()))
			return err
		}
		if err != nil {
			return err
		}
		this.sftp.Mkdir(remotePath)
		//遍历文件夹内容
		for _, backupDir := range localFiles {
			localFilePath := path.Join(localPath, backupDir.Name())
			remoteFilePath := path.Join(remotePath, backupDir.Name())
			//判断是否是文件,是文件直接上传.是文件夹,先远程创建文件夹,再递归复制内部文件
			if backupDir.IsDir() {
				this.sftp.Mkdir(remoteFilePath)
				this.uploadDirectory(localFilePath, remoteFilePath)
			} else {
				this.uploadFile(path.Join(localPath, backupDir.Name()), remotePath)
			}
		}
		this.uploadDirectory(localPath, remotePath)
	} else {
		text := "没有此目录"
		this.Errorf(text)
	}

	this.Warnf("全部结束!")
	return nil
}

func (this *SshClient) RemoveSub(remotePath string) error {

	f, err := this.sftp.Stat(remotePath)
	if err != nil {
		this.Info(remotePath + ":原始文件不存在,无需删除")
		return err
	}
	if f.IsDir() {
		remoteFiles, err := this.sftp.ReadDir(remotePath)
		if err != nil {
			this.Error(remotePath + ":文件不存在,无需移除 " + err.Error())
			return err
		}
		for _, backupDir := range remoteFiles {
			remoteFilePath := path.Join(remotePath, backupDir.Name())
			if backupDir.IsDir() {
				err := this.removeDirectory(remoteFilePath)
				if err != nil {
					return err
				}
			} else {
				this.sftp.Remove(path.Join(remoteFilePath))
			}
		}
	} else {
		this.Warnf("没有此目录")
	}
	return nil
}

func (this *SshClient) removeFile(remotePath string) error {

	this.sftp.Remove(path.Join(remotePath))
	log.Info(remotePath + "  delete file")
	return nil
}

func (this *SshClient) removeDirectory(remotePath string) error {
	//打不开,说明要么文件路径错误了,要么是第一次部署
	remoteFiles, err := this.sftp.ReadDir(remotePath)
	if err != nil {
		this.Error(remotePath + ":文件不存在,无需移除 " + err.Error())
		return err
	}
	//和上传文件逻辑差不多,用递归的方法删除
	defer this.sftp.RemoveDirectory(remotePath)
	for _, backupDir := range remoteFiles {
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			err := this.removeDirectory(remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			this.sftp.Remove(path.Join(remoteFilePath))
		}
	}

	this.Info(remotePath + "  delete directory")
	return nil
}

func winBase(path string) string {
	if path == "" {
		return "."
	}
	// Strip trailing slashes.
	for len(path) > 0 && path[len(path)-1] == '\\' {
		path = path[0 : len(path)-1]
	}
	// Find the last element
	if i := strings.LastIndex(path, "\\"); i >= 0 {
		path = path[i+1:]
	}
	if i := strings.LastIndex(path, "/"); i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if path == "" {
		return "/"
	}
	return path
}

func (this *SshClient) uploadFile(localFilePath string, remotePath string) error {
	//打开本地文件流
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		this.Error("os.Open error : " + localFilePath)
		this.Error(err.Error())
		return err
	}
	//关闭文件流
	defer srcFile.Close()
	//上传到远端服务器的文件名,与本地路径末尾相同
	var remoteFileName string
	if runtime.GOOS == "windows" {
		remoteFileName = winBase(localFilePath)
	} else {
		remoteFileName = path.Base(localFilePath)
	}
	//打开远程文件,如果不存在就创建一个

	if err != nil {
		return err
	}
	dstFile, err := this.sftp.Create(path.Join(remotePath, remoteFileName))
	log.Warnf("distFile", zap.Reflect("dstFile", dstFile))
	if err != nil {
		this.Error(fmt.Sprintf("sftpClient.OnCreate error : %s", path.Join(remotePath, remoteFileName)))
		this.Error(err.Error())

	}
	//关闭远程文件
	defer dstFile.Close()
	//读取本地文件,写入到远程文件中(这里没有分快穿,自己写的话可以改一下,防止内存溢出)
	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		this.Error("ReadAll error :" + localFilePath)
		this.Error(err.Error())
		return err
	}
	dstFile.Write(ff)
	this.Info(localFilePath + "  copy file to remote " + path.Join(remotePath, remoteFileName) + " finished!")
	return nil
}

func (this *SshClient) uploadDirectory(localPath string, remotePath string) error {
	//打开本地文件夹流
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		this.Error(fmt.Sprintf("路径错误 %s", err.Error()))
		return err
	}
	//先创建最外层文件夹
	this.sftp.Mkdir(remotePath)
	//遍历文件夹内容
	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		//判断是否是文件,是文件直接上传.是文件夹,先远程创建文件夹,再递归复制内部文件
		if backupDir.IsDir() {
			this.sftp.Mkdir(remoteFilePath)
			this.uploadDirectory(localFilePath, remoteFilePath)
		} else {
			this.uploadFile(path.Join(localPath, backupDir.Name()), remotePath)
		}
	}
	this.Info(localPath + "  copy directory to remote " + remotePath + " finished!")
	return nil
}

func (this *SshClient) Gzip(files []*os.File, dest string) error {
	return gzip.Compress(files, dest)
}

func (this *SshClient) Zip(files []*os.File, dest string) error {
	return gzip.Compress(files, dest)
}

func (this *SshClient) UnZip(file, dest string) error {
	return zip.DeCompress(file, dest)
}

func (this *SshClient) UnGzip(file, dest string) error {
	return gzip.DeCompress(file, dest)
}

func (this *SshClient) getDir(path string) string {
	return this.subString(path, 0, strings.LastIndex(path, "/"))
}

func (this *SshClient) subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < start || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}
