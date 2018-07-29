import argparse
import numpy as np
import json
from keras.models import load_model
import codecs
import random, string, sys
import h5py, h5ToJson, os

def calculateDifference(endWeightsMap, beginningWeightsMap):
    differenceWeightsMap = {}
    for key in endWeightsMap.iterkeys():
        differenceWeightsMap[key] = np.asarray(endWeightsMap[key]-beginningWeightsMap[key]).tolist()
    return differenceWeightsMap

parser = argparse.ArgumentParser()

parser.add_argument("-i","--inputFile", nargs='+',help="Location of Neural Network Inputs As Json")
parser.add_argument("-o","--outputFile", nargs='+',help="Location of Neural Network Outputs As Json")
parser.add_argument("-m","--modelFile", help="Location of Model Json (must be .hd5)")
parser.add_argument("-b","--batchSize", type=int,help="Size of batch to train on")
parser.add_argument("-e","--epochs", type=int, help="Number of Epochs to train on")

args = parser.parse_args()

inputFiles = args.inputFile
outputFiles = args.outputFile
modelFile = args.modelFile
epochs = args.epochs
batchSize = args.batchSize


if (batchSize == 0):
    batchSize = None

inputArrays = []
outputArrays = []

for inputFile in inputFiles:
    obj_text = codecs.open(inputFile, 'r', encoding='utf-8').read()
    b_new = json.loads(obj_text)
    a_new = np.array(b_new)
    inputArrays.append(a_new)

for outputFile in outputFiles:
    obj_text = codecs.open(outputFile, 'r', encoding='utf-8').read()
    b_new = json.loads(obj_text)
    a_new = np.array(b_new)
    outputArrays.append(a_new)

loaded_model = load_model(modelFile)

weightsFileName = (''.join(random.SystemRandom().choice(string.ascii_uppercase + string.digits) for y in range(20))) + ".h5"
jsonFileName = (''.join(random.SystemRandom().choice(string.ascii_uppercase + string.digits) for x in range(20))) + ".json"

loaded_model.fit(inputArrays,outputArrays,epochs=epochs,batch_size=batchSize, verbose=0)

loaded_model.save(weightsFileName)

h5File = h5py.File(weightsFileName, 'r')

layerToWeightsMap = {}
configToTrainMap = {}

for dSetName in h5File["model_weights"].iterkeys():
    for layer in h5File["model_weights"][dSetName].iterkeys():
        for layerConfig in h5File["model_weights"][dSetName][layer].iterkeys():
            key = dSetName+"/"+layer+"/"+layerConfig
            layerToWeightsMap[key]=np.asarray(h5File["model_weights"][dSetName][layer][layerConfig].value).tolist()

finalParamsMap = {}
finalParamsMap["weights"]=layerToWeightsMap

fileAttributes = h5File.attrs
configMap = json.loads(fileAttributes['training_config'])
for k,v in configMap["optimizer_config"]["config"].iteritems():
    configToTrainMap[k]=v

finalParamsMap["config"]=configToTrainMap

with open(jsonFileName, 'w') as fp:
    json.dump(finalParamsMap, fp)

os.remove(weightsFileName)

sys.stdout.write(jsonFileName)