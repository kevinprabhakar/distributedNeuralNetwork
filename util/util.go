package util

import(

	"io/ioutil"
	"errors"
	"encoding/json"
	"fmt"
	"os"
	"encoding/gob"
	"math/rand"
	"bytes"
	"time"
)



type JsonData []interface{}

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)



func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func NewSplitJsonArray(file []byte, peers int)(){
	var fullData []interface{}

	err := json.Unmarshal(file, &fullData)
	if err != nil {
		return

	}

	fmt.Println(len(fullData))

	stringForm := fmt.Sprintf("%v", fullData)
	fmt.Println(stringForm)

}

func SplitJsonArray(file []byte, peers int) (*[]string, error){
	var fullData JsonData

	err := json.Unmarshal(file, &fullData)
	if err != nil {
		return nil, err

	}

	if (len(fullData) < peers){
		return nil, errors.New("More peers requested than data points available")
	}

	returnData := make([]string,0)

	incrFactor := len(fullData)/peers

	for i:=0 ; i < (peers-1) ;i++ {
		batch, err := json.Marshal(fullData[(i*incrFactor):((i+1)*incrFactor)])
		if (err != nil){
			return nil, err
		}
		stringForm := fmt.Sprintf("%v", string(batch))

		returnData = append(returnData, stringForm)
	}

	batch, err := json.Marshal(fullData[((peers-1)*incrFactor):])
	if (err != nil){
		return nil, err
	}
	stringForm := fmt.Sprintf("%v", string(batch))

	returnData = append(returnData, stringForm)


	return &returnData, nil
}

func GetByteFormOfFiles(fileList []string)(*[][]byte, error){
	inputFileStrings := make([][]byte, 0)

	for _, fileName := range fileList{
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		data, err := ioutil.ReadAll(file)
		if (err != nil){
			fmt.Println("Input File is Invalid")
			return nil, err
		}
		inputFileStrings = append(inputFileStrings, data)
	}

	return &inputFileStrings, nil
}

var src = rand.NewSource(time.Now().UnixNano())

func RandomStringWithLength(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func WriteBytesToFile(data []byte, fileName string, filetype string)(error){
	fullFile := fmt.Sprintf("%s.%s", fileName, filetype)

	file, err := os.Create(fullFile) // Truncates if file already exists, be careful!
	if err != nil {
		return err
	}
	_, err = file.Write(data)

	if err != nil {
		return err
	}
	return nil
}

func WriteStringToFile(data string, fileName, filetype string)(error){
	fullFile := fmt.Sprintf("%s.%s", fileName, filetype)

	file, err := os.Create(fullFile) // Truncates if file already exists, be careful!
	if err != nil {
		return err
	}
	_, err = file.WriteString(data)

	if err != nil {
		return err
	}
	return nil
}

func WriteBytesToJsonFile(fileName string, data []byte)(error){
	file, err := os.Create(fileName)
	if (err != nil){
		return err
	}

	_, writeErr := file.Write(data)
	if (writeErr != nil){
		return writeErr
	}

	return nil
}