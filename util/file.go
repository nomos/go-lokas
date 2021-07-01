package util

import (
	"encoding/json"
	"errors"
	"github.com/nomos/go-lokas/log"
	"io/ioutil"
	"os"
	"path"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type WalkDirFunc func(filePath string, file os.FileInfo) bool

func CreateFile(filePath string,perms... int)  error  {
	perm := 0644
	if len(perms)>0 {
		perm = perms[0]
	}
	dirPath := path.Dir(filePath)
	if !IsFileExist(dirPath) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err!= nil {
			return err
		}
		_, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
		return err
	}
	return nil
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func FilterFileWithExt(files []string,ext... string)[]string {
	ret:=make([]string,0)
	for _,file:=range files {
		for _,v:=range ext {
			if v==path.Ext(file) {
				ret = append(ret, file)
			}
		}
	}
	return ret
}

func FilterFileWithFunc(files []string,f func(string)bool)[]string {
	ret:=make([]string,0)
	for _,file:=range files {
		if f(file) {
			ret = append(ret, file)
		}
	}
	return ret
}

func WalkDir(dirPath string, recursive bool) ([]string, error) {
	ret := make([]string, 0)
	return walkDir(dirPath, ret, 3, nil, recursive)
}

func WalkDirWithFunc(dirPath string, walkFunc WalkDirFunc, recursive bool) ([]string, error) {
	ret := make([]string, 0)
	return walkDir(dirPath, ret, 3, walkFunc, recursive)
}

func FindFile(dirPath string,name string,recursive bool)string {
	ret:=""
	WalkDirFilesWithFunc(dirPath, func(filePath string, file os.FileInfo) bool {
		if file.Name() == name {
			ret = filePath
			return true
		}
		return false
	},recursive)
	return ret
}

func WalkDirFiles(dirPath string, recursive bool) ([]string, error){
	ret := make([]string, 0)
	return walkDir(dirPath, ret, 1, nil, recursive)
}

func WalkDirFilesWithFunc(dirPath string, walkFunc WalkDirFunc, recursive bool) ([]string, error) {
	ret := make([]string, 0)
	return walkDir(dirPath, ret, 1, walkFunc, recursive)
}

func WalkDirDirs(dirPath string, recursive bool) ([]string, error) {
	ret := make([]string, 0)
	return walkDir(dirPath, ret, 2, nil, recursive)
}

func WalkDirDirsWithFunc(dirPath string, walkFunc WalkDirFunc, recursive bool) ([]string, error) {
	ret := make([]string, 0)
	return walkDir(dirPath, ret, 2, walkFunc, recursive)
}

func walkDir(dirPath string, files []string, typ int, walkFunc WalkDirFunc, recursive bool) ([]string, error) {
	read, err := ioutil.ReadDir(dirPath) //读取文件夹
	if err != nil {
		log.Error(err.Error())
		return files, errors.New("文件夹不可读取")
	}
	// 如果是文件夹，进行递归处理
	for _, fi := range read {
		if fi.IsDir() { // 判断是否是文件夹
			fullDir := dirPath + "/" + fi.Name() //构造新的路径
			if typ>>1&1 == 1 {
				files = append(files, fullDir) //追加路径
				if walkFunc != nil {
					if walkFunc(fullDir, fi) {
						return files, nil
					}
				}
			}
			if recursive {
				files, _ = walkDir(fullDir, files, typ, walkFunc, recursive) //文件夹递归
			}
		} else {
			fullDir := dirPath + "/" + fi.Name() //构造新的路径
			if typ>>0&1 == 1 {
				files = append(files, fullDir) //追加路径
				if walkFunc != nil {
					if walkFunc(fullDir, fi) {
						return files, nil
					}
				}
			}
		}
	}
	return files, nil
}


func LoadJson (path string,v interface{}) {
	data,err := ioutil.ReadFile(path)
	if err!=nil {
		return
	}
	err = json.Unmarshal(data,v)
	if err!=nil {
		return
	}
}
