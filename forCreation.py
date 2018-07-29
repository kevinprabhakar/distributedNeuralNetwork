import argparse
import numpy as np
import json
from keras.models import load_model
import codecs
import random, string, sys
import h5py, h5ToJson
from numpy import linalg as LA
import time, os

inputFiles = ["inputs.json"]
outputFiles = ["outputs.json"]
modelFile = "5ST53WCXLV6X4XFS131I.h5"

def calculateDifference(endWeightsMap, beginningWeightsMap):
    differenceWeightsMap = {}
    for key in endWeightsMap.iterkeys():
        differenceWeightsMap[key] = np.asarray(endWeightsMap[key]-beginningWeightsMap[key]).tolist()
    return differenceWeightsMap


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

print LA.norm(loaded_model.predict_on_batch(inputArrays) - outputArrays) / LA.norm(outputArrays)