import argparse
import numpy as np
import json
from keras.models import load_model
import codecs
import random, string, sys, os
import h5py, h5ToJson, jsonToH5
from shutil import copyfile

parser = argparse.ArgumentParser()

parser.add_argument("-f","--files", nargs='+',help="Locations of HD5 File Weights as JSON")
parser.add_argument("-m","--originalmodel", help="Location of the original model HD5 File")

args = parser.parse_args()

originalModel ="J5EHUL2GLQC0D3ZFWT5I.h5"

fileNames = ["PF0QVZSZAII7YW559BF6.json"]


modelsWeightsArray = []

TotalWeightsMap = {}
TotalConfigMap = {}

for file in fileNames:
    data = json.load(open(file))
    modelsWeightsArray.append(data)

for layerName in modelsWeightsArray[0]["weights"].iterkeys():
    layerWeightsAll = []
    for model in modelsWeightsArray:
        a_new = np.asarray(model["weights"][layerName])
        layerWeightsAll.append(a_new)
    TotalWeightsMap[layerName]=layerWeightsAll

for configVar in modelsWeightsArray[0]["config"].iterkeys():
    configValuesAll = []
    for model in modelsWeightsArray:
        a_new = np.asarray(model["config"][configVar])
        configValuesAll.append(a_new)
    TotalConfigMap[configVar] = configValuesAll


def average(listOfNP):
    return np.mean(listOfNP, axis=0)

finalWeightsMap = {}
finalConfigMap = {}

for key in TotalWeightsMap.iterkeys():
    avg = average(TotalWeightsMap[key])
    finalWeightsMap[key]=avg

for key in TotalConfigMap.iterkeys():
    avg = np.mean(TotalConfigMap[key])
    finalConfigMap[key]=avg
    print avg

loaded_model = load_model(originalModel)

newModelName = (''.join(random.SystemRandom().choice(string.ascii_uppercase + string.digits) for _ in range(20))) + ".h5"

loaded_model.save(newModelName)

h5File2 = h5py.File(newModelName)

for dSetName in h5File2["model_weights"].iterkeys():
    for layer in h5File2["model_weights"][dSetName].iterkeys():
        for layerConfig in h5File2["model_weights"][dSetName][layer].iterkeys():
            key = dSetName+"/"+layer+"/"+layerConfig
            del h5File2["model_weights"][key]
            dset = h5File2["model_weights"].create_dataset(key, data=finalWeightsMap[key])

attributes = h5File2.attrs

obj = json.loads(attributes.get("training_config"))
for key in TotalConfigMap.iterkeys():
    obj["optimizer_config"]["config"][key] = TotalConfigMap[key][0].tolist()
    print TotalConfigMap[key]

attributes.modify("training_config", json.dumps(obj))

sys.stdout.write(newModelName)

