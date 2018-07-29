package main

import (
	"testing"
	"oneOffProjects/p2pTesting/util"
	"fmt"
	"net/url"
)

func TestStuff(t *testing.T){
	byteForm, _ := util.GetByteFormOfFiles([]string{"foo.json"})
	stringForm, _ := util.SplitJsonArray((*byteForm)[0], 4)

	fmt.Println((*stringForm)[2])

}

func TestStuff2(t *testing.T){
	byteForm, _ := util.GetByteFormOfFiles([]string{"foo.json"})
	util.NewSplitJsonArray((*byteForm)[0],12)
}

func TestFileCreation(t *testing.T){
	byteForm, _ := util.GetByteFormOfFiles([]string{"foo.json"})
	stringForm, _ := util.SplitJsonArray((*byteForm)[0], 4)
	fileName := util.RandomStringWithLength(20)
	fileType := "json"
	fmt.Println((*stringForm)[0])
	util.WriteStringToFile((*stringForm)[0], fileName, fileType)
}

func createApiCall(t *testing.T){
	//url := "https://api.deepai.org/api/fast-style-transfer"

	body := url.Values{}
	body.Set("api-key","6279d655-ab5e-4c1e-aca0-7a6b519f7a21")
	body.Set("input")
}