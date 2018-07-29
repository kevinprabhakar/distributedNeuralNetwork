package config

type ConfigFile struct{
	WeightsFileLocation		string 		`json:"weightsFileLocation"`
	ModelConfigLocation 	string		`json:"modelConfigLocation"`
	InputFileLocation 		string 		`json:"inputFileLocation"`
	OuputFileLocation 		string 		`json:"outputFileLocation"`
	GPUOptimized 			bool 		`json:"GPUOptimized"`
}

