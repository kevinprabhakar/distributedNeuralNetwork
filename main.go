package main

import (
	"oneOffProjects/p2pTesting/host"
	"fmt"
	"gx/ipfs/QmbD5yKbXahNvoMqzeuNyKQA9vAs9fUvJg2GXeWU1fVqY5/go-libp2p-net"
	"bufio"
	"flag"
	"oneOffProjects/p2pLocationServer/mongo"
	"io/ioutil"
	"net/url"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"oneOffProjects/p2pTesting/util"
	"encoding/json"
	"context"
	"github.com/libp2p/go-libp2p-peerstore"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

)

var HardCodedUrl = "http://localhost:3000"

var ProtocolName protocol.ID = "echo/1.0.0"

func GetListeners()(*[]mongo.Listener, error){
	var Listeners []mongo.Listener

	resp, err := http.PostForm(HardCodedUrl+"/api/getListeners",url.Values{})
	if (err != nil){
		fmt.Println("Couldn't get post form")
		return nil, err
	}

	byteForm, err := ioutil.ReadAll(resp.Body)
	if (err != nil){
		fmt.Println("Couldn't convert to byte form")
		return nil, err
	}

	err = json.Unmarshal(byteForm, &Listeners)
	if (err != nil){
		fmt.Println("Couldn't marshall to listeners")
		return nil, err
	}

	return &Listeners, nil
}

func GenerateCLIFriendlyInputsValidation(fileHashMap *map[string]string, Inputs []string, Outputs []string)(string, string){
	inputFilesToDelete := make([]string, 0)
	outputFilesToDelete := make([]string, 0)
	inputCLI := ""
	outputCLI := ""

	for input := range Inputs{
		key := fmt.Sprintf("input%d",input)
		(*fileHashMap)[key] = util.RandomStringWithLength(20)
		util.WriteStringToFile(Inputs[input], (*fileHashMap)[key], "json")
		if (input == len(Inputs)-1){
			inputCLI += (*fileHashMap)[key]+".json"
		}else{
			inputCLI += (*fileHashMap)[key]+".json"+" "
		}
		inputFilesToDelete = append(inputFilesToDelete, (*fileHashMap)[key])
	}
	for output := range Outputs{
		key := fmt.Sprintf("output%d",output)
		(*fileHashMap)[key] = util.RandomStringWithLength(20)
		util.WriteStringToFile(Outputs[output], (*fileHashMap)[key], "json")
		if (output == len(Outputs)-1){
			outputCLI += (*fileHashMap)[key]+".json"
		}else{
			outputCLI += (*fileHashMap)[key]+".json"+" "
		}
		outputFilesToDelete = append(outputFilesToDelete, (*fileHashMap)[key])
	}

	return inputCLI, outputCLI
}

func ExecuteValidateCommand(inputCLI string, outputCLI string, modelCLI string)(bool, error){
	validateCmd := exec.Command("python", "validateDataset.py", "-i", inputCLI, "-o", outputCLI, "-m", modelCLI)


	fmt.Println("Validating dataset...")

	combinedOutput, err := validateCmd.CombinedOutput()

	if (err != nil){
		return true, err
	}

	valid, err := regexp.Match("Validation Complete", combinedOutput)
	if (err != nil){
		return true, err
	}

	return valid, nil
}

func RemovePeerFromServer(peerID string){
	body := url.Values{}
	body.Add("peerID", peerID)

	resp, err := http.PostForm(HardCodedUrl+"/api/delListener",body)
	if (err != nil){
		fmt.Println(err.Error())
		return
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(respBody))
}

func main(){
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	port := flag.Int("p",0,"what port to listen in on")
	listener := flag.Bool("l", true, "bool value of whether host is listener")
	inputFiles := flag.String("i","","list of numpy input arrays in json format")
	outputFiles := flag.String("o","","list of numpy output arrays in json format")
	generations := flag.Int("u",1,"number of generations to train on")
	modelFile := flag.String("m", "", "hd5 model file")
	epochs := flag.Int("e",1,"how many epochs to train model for")
	batchSize := flag.Int("b", 1, "batch size")

	flag.Parse()

	if *port == 0 {
		fmt.Println("Please use the -p <port> utility on command line.")
		return
	}

	ha, hostAddr, hostID := host.NewHost(*port)

	go func(peerID string) {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		RemovePeerFromServer(peerID)
		// do last actions and wait for all write operations to end
		os.Exit(0)
	}(hostID)

	ha.SetStreamHandler(ProtocolName, func(s net.Stream){
		fmt.Println("Got a new stream!")
		if err := trainNetwork(s); err != nil {
			fmt.Println("Unable to train Network. Check parameters/dataset")
			s.Reset()
		} else {
			s.Close()
		}
	})

	if *listener == true{
		fmt.Println("Listening for connections")
		body := url.Values{}
		body.Add("peerID", hostID)
		body.Add("multiAddr", hostAddr)
		resp, err := http.PostForm(HardCodedUrl+"/api/addListener",body)
		if (err != nil){
			fmt.Println(err.Error())
			return
		}
		respBody, _ := ioutil.ReadAll(resp.Body)

		fmt.Println(string(respBody))
		select{}
	}else{
		if (*modelFile == "")||(*inputFiles == "")||(*outputFiles == ""){
			fmt.Println("You forgot to specify some files")
			return
		}

		scopeModelFile := *modelFile

		fmt.Println("Beginning Training Procedure")
		inputFileList := strings.Split(*inputFiles, ",")
		outputFileList := strings.Split(*outputFiles, ",")

		inputFileBytes, err := util.GetByteFormOfFiles(inputFileList)
		if (err != nil){
			fmt.Println(err.Error())
			return
		}

		outputFileBytes, err := util.GetByteFormOfFiles(outputFileList)
		if (err != nil){
			fmt.Println(err.Error())
			return
		}

		listeners, err := GetListeners()
		if (err != nil){
			fmt.Println(err.Error())
			return
		}

		fileHashMap := make(map[string]string,0)
		inputCLI, outputCLI := GenerateCLIFriendlyInputsValidation(&fileHashMap,inputFileList, outputFileList)

		valid, err := ExecuteValidateCommand(inputCLI, outputCLI, scopeModelFile)

		if (!valid){
			fmt.Println("Dataset is not valid")
			return
		}

		fmt.Println("Splitting Input/Output Arrays")
		inputStringStream := make([][]string,0)
		outputStringStream := make([][]string, 0)

		for _, inputByteForm := range *inputFileBytes{
			stringList, err := util.SplitJsonArray(inputByteForm, len(*listeners))
			if (err != nil){
				fmt.Println(err.Error())
				return
			}

			inputStringStream = append(inputStringStream, *stringList)
		}
		for _, outputByteForm := range *outputFileBytes{
			stringList, err := util.SplitJsonArray(outputByteForm, len(*listeners))
			if (err != nil){
				fmt.Println(err.Error())
				return
			}

			outputStringStream = append(outputStringStream, *stringList)
		}

		numListeners := len(*listeners)

		fmt.Printf("Sending Train Commands to %d Peers\n", numListeners)
		modelFilesToDelete := make([]string, 0)

		for gen:=0 ; gen<*generations ; gen++ {
			modelBytes, err := util.GetByteFormOfFiles([]string{scopeModelFile})
			if (err != nil){
				fmt.Println(err.Error())
				return
			}

			wg := sync.WaitGroup{}

			resultsFileNameChannel := make(chan string, numListeners)

			for listenerIndex := range *listeners{
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					listener := (*listeners)[index]
					peerId, targetAddress := host.GetPIDandTargetAddrFromTargetURL(listener.MultiAddr)
					ha.Peerstore().AddAddr(*peerId, *targetAddress, peerstore.PermanentAddrTTL)
					stream, err := ha.NewStream(context.Background(), *peerId, ProtocolName)
					if (err != nil) {
						fmt.Println(err.Error())
						return
					}

					peerInputs := make([]string, 0)
					peerOutputs := make([]string, 0)

					for _, input := range inputStringStream {
						peerInputs = append(peerInputs, input[index])
					}
					for _, output := range outputStringStream {
						peerOutputs = append(peerOutputs, output[index])
					}

					payload, err := json.Marshal(host.ClientPayload{(*modelBytes)[0], peerInputs, peerOutputs, *batchSize, *epochs})
					if (err != nil) {
						fmt.Println(err.Error())
						return
					}

					payload = append(payload, '\n')

					_, err = stream.Write(payload)
					if (err != nil) {
						fmt.Println(err.Error())
						return
					}

					out, err := ioutil.ReadAll(stream)
					if err != nil {
						fmt.Println(err.Error())
						return
					}

					fileName := util.RandomStringWithLength(20)+".json"

					fileCreateErr := util.WriteBytesToJsonFile(fileName, out)
					if (fileCreateErr != nil) {
						fmt.Println(err.Error())
						return
					}

					resultsFileNameChannel <- fileName
				}(listenerIndex)
			}
			wg.Wait()
			close(resultsFileNameChannel)

			resultWeightsFileNames := []string{"averageJsonH5.py", "-f"}

			for fileName := range resultsFileNameChannel{
				resultWeightsFileNames = append(resultWeightsFileNames, fileName)
			}

			fmt.Println("Received weights files from all peers")
			fmt.Println("Averaging Weights Files")

			resultWeightsFileNames = append(resultWeightsFileNames, "-m",scopeModelFile)

			cmd2 := exec.Command("python", resultWeightsFileNames...)

			out, err := cmd2.Output()

			if (err != nil){
				fmt.Println(err.Error())
				return
			}

			fileName := string(out)

			fmt.Printf("Generation %d Model File: %s\n",gen, fileName)
			fmt.Print(resultWeightsFileNames[2:len(resultWeightsFileNames)-2])

			if ((gen > 0)&&(gen < (*generations-1))){
				modelFilesToDelete = append(modelFilesToDelete, scopeModelFile)
			}

			scopeModelFile = fileName

			err = CleanUpEndJSONFiles(resultWeightsFileNames[2:len(resultWeightsFileNames)-2])

			if err != nil {
				fmt.Println(err.Error())
				return
			}


		}
		for _, modelFile := range modelFilesToDelete{
			err := CleanUp(modelFile)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}



}

func CleanUp(fileName string)(error){
	err := os.Remove(fileName)
	return err
}

func CleanUpEndJSONFiles(files []string)(error){
	for _, file := range files{
		err := CleanUp(file)
		if (err != nil){
			return err
		}
	}
	return nil
}

func CleanupHashedFiles(modelFile string, inputFiles []string, outputFiles []string, weightsFile string)(error){
	for _, inputFile := range inputFiles {
		err := CleanUp(inputFile)
		if (err != nil){
			return err
		}
	}
	for _, outputFile := range outputFiles {
		err := CleanUp(outputFile)
		if (err != nil){
			return err
		}
	}
	err := CleanUp(modelFile + ".h5")
	if (err != nil){
		return err
	}

	//".h5" not needed because it's set in the python file
	err = CleanUp(weightsFile)
	return err
}

func doEcho(s net.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Println("read: "+ str)
	_, err = s.Write([]byte("Message Received"))
	return err
}

func trainNetwork(s net.Stream) error{
	buf := bufio.NewReader(s)
	byteForm, err := buf.ReadBytes('\n')
	if err != nil {
		return err
	}

	var payload host.ClientPayload

	json.Unmarshal(byteForm, &payload)

	fileHashMap := make(map[string]string, 0)

	fileHashMap["modelName"] = util.RandomStringWithLength(20)
	err = util.WriteBytesToFile(payload.Model, fileHashMap["modelName"],"h5")
	if err != nil {
		return err
	}

	inputCLI, outputCLI := GenerateCLIFriendlyInputsValidation(&fileHashMap,payload.Inputs,payload.Outputs)
	modelCLI := fileHashMap["modelName"] + ".h5"

	inputFilesToDelete := strings.Split(inputCLI, " ")
	outputFilesToDelete := strings.Split(outputCLI, " ")

	fmt.Println("Beginning Training")
	batchSize := strconv.Itoa(payload.BatchSize)
	epochs := strconv.Itoa(payload.Epochs)
	cmd:= exec.Command("python", "ASGDTrainParamAve.py", "-i", inputCLI, "-o", outputCLI, "-m", modelCLI, "-b", batchSize, "-e", epochs)
	out, err := cmd.Output()

	if (err != nil){
		return err
	}

	fileName := string(out)

	fileHashMap["weightsName"] = fileName

	byteFormOfWeights, err := ioutil.ReadFile(fileName)

	if (err != nil){
		return err
	}

	_, err = s.Write([]byte(byteFormOfWeights))

	if (err != nil){
		return err
	}

	fmt.Println("Successfully finished training")

	err = CleanupHashedFiles(fileHashMap["modelName"],inputFilesToDelete, outputFilesToDelete, fileHashMap["weightsName"])

	return nil


	return nil
}