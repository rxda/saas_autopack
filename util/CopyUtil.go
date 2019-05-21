package util

import (
	"io"
	"os"
)

func CopyFolder(source string, dest string) (err error) {

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceInfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	defer directory.Close()
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourceFilePointer := source + "/" + obj.Name()

		destinationFilePointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			err = CopyFolder(sourceFilePointer, destinationFilePointer)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(sourceFilePointer, destinationFilePointer)
			if err != nil {
				return err
			}
		}

	}
	return
}

func CopyFile(source string, dest string) (err error) {
	fromFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	if err != nil {
		return err
	}
	return nil
}
