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

inputFiles =  ["inputs.json"]
outputFiles = ["outputs.json"]
modelFile = "newCoolFile.h5"
epochs = 5
batchSize = 0

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

loaded_model.fit(inputArrays,outputArrays,epochs=epochs,batch_size=batchSize, verbose=0)

weightsFileName = (''.join(random.SystemRandom().choice(string.ascii_uppercase + string.digits) for y in range(20))) + ".h5"
jsonFileName = (''.join(random.SystemRandom().choice(string.ascii_uppercase + string.digits) for x in range(20))) + ".json"

loaded_model.save_weights(weightsFileName)

h5File = h5py.File(weightsFileName, 'r')

layerToWeightsMap = {}

for dSetName in h5File.iterkeys():
    for layer in h5File[dSetName].iterkeys():
        for layerConfig in h5File[dSetName][layer].iterkeys():
            key = dSetName+"/"+layer+"/"+layerConfig
            layerToWeightsMap[key]=np.asarray(h5File[dSetName][layer][layerConfig].value).tolist()

with open(jsonFileName, 'w') as fp:
    json.dump(layerToWeightsMap, fp)

os.remove(weightsFileName)

sys.stdout.write(jsonFileName)