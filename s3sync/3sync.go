package s3sync

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

const debug = true

func hashFile(reader io.Reader) []byte {
	hasher := md5.New()
	io.Copy(hasher, reader)
	return hasher.Sum(nil)
}

func hashLocal(fname string) []byte {
	file, err := os.Open(fname)
	if err != nil {
		panic(err.Error())
	}
	return hashFile(file)
}

func hashRemote(bucket *s3.Bucket, path string) []byte {
	file, err := bucket.GetReader(path)
	if err != nil {
		return nil
	}
	return hashFile(file)
}

func syncPath(acl, localPath, path string, bucket *s3.Bucket) {
	s3Path := strings.Replace(path, localPath, "", -1)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening local file, Failed to sync '" + path + "'")
		return
	}
	info, err := file.Stat()
	if err != nil {
		fmt.Println("Error stating local file, Failed to sync '" + path + "'")
		return
	}
	length := info.Size()
	pathSplit := strings.Split(path, ".")
	ext := pathSplit[len(pathSplit)-1]
	mtype := mime.TypeByExtension(ext)
	puterr := bucket.PutReader(s3Path, file, length, mtype, s3.ACL(acl))
	if puterr != nil {
		fmt.Println("Failed to sync: " + s3Path)
		return
	}
	fmt.Println("Synced: " + s3Path)

}

func compareAndSync(acl, localPath, path string, bucket *s3.Bucket, doneChan chan int) {
	s3Path := strings.Replace(path, localPath, "", -1)
	localHash := hashLocal(path)
	s3Hash := hashRemote(bucket, s3Path)
	if s3Hash == nil || !bytes.Equal(s3Hash, localHash) {
		syncPath(acl, localPath, path, bucket)
	}
	doneChan <- 1
}

func doTheWalk(localPath string, pathChan chan string) {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			pathChan <- path
		}
		return err
	}
	filepath.Walk(localPath, walkFunc)
	pathChan <- ""
	return
}

func Sync(localPath, bucketName, awsRegion, acl string) {
	if localPath == "" || bucketName == "" {
		flag.PrintDefaults()
		return
	}
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}
	s3Conn := s3.New(auth, aws.Regions[awsRegion])
	bucket := s3Conn.Bucket(bucketName)
	pathChan := make(chan string)
	doneChan := make(chan int)
	go doTheWalk(localPath, pathChan)
	count := 0
	for path := range pathChan {
		if path == "" {
			break
		}
		go compareAndSync(acl, localPath, path, bucket, doneChan)
		count++
	}
	for i := 0; i < count; i++ {
		<-doneChan
	}
}
