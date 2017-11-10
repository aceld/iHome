/**
* @file fdfs_op.go
* @brief  fdfs go语言接口集成 目前支持 上传和删除
* @author

Aceld(LiuDanbing)

email: danbing.at@gmail.com
Blog: http://www.gitbook.com/@aceld

* @version 1.0
* @date 2017-11-05
*/
package utils

import (
	"fmt"
	"github.com/weilaihui/fdfs_client"
)

func FDFSUploadByFileName(filename string) (groupName string, fileId string, err error) {
	fdfsClient, err := fdfs_client.NewFdfsClient("./conf/client.conf")
	if err != nil {
		fmt.Printf("New FdfsClient error %s", err.Error())
		return "", "", err
	}
	uploadResponse, err := fdfsClient.UploadByFilename(filename)
	if err != nil {
		fmt.Printf("UploadByfilename error %s", err.Error())
		return "", "", err
	}
	/*
		fmt.Println(uploadResponse.GroupName)
		fmt.Println(uploadResponse.RemoteFileId)
	*/

	//fdfsClient.DeleteFile(uploadResponse.RemoteFileId)

	return uploadResponse.GroupName, uploadResponse.RemoteFileId, nil
}

func FDFSUploadByBuffer(buffer []byte, suffix string) (gourpName string, fileId string, err error) {

	fdfsClient, err := fdfs_client.NewFdfsClient("./conf/client.conf")
	if err != nil {
		fmt.Printf("New FdfsClient error %s", err.Error())
		return "", "", err
	}

	uploadResponse, err := fdfsClient.UploadByBuffer(buffer, suffix)
	if err != nil {
		fmt.Println("TestUploadByBuffer error %s", err.Error())
		return "", "", err
	}

	/*
		fmt.Println(uploadResponse.GroupName)
		fmt.Println(uploadResponse.RemoteFileId)
	*/
	//fdfsClient.DeleteFile(uploadResponse.RemoteFileId)

	return uploadResponse.GroupName, uploadResponse.RemoteFileId, nil
}

func FDFSDeleteByFileId(fileId string) error {

	fdfsClient, err := fdfs_client.NewFdfsClient("./conf/client.conf")
	if err != nil {
		fmt.Printf("New FdfsClient error %s", err.Error())
		return err
	}

	fdfsClient.DeleteFile(fileId)

	return nil
}
