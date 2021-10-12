import json
import pathlib
from pathlib import Path
from typing import Dict, List
import statistics
import numpy as np
import math
from pprint import pprint
import sys
import os
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt
from collections import *

def parseMetrics(line: str) -> Dict:
    obj = json.loads(json.loads(line)["msg"])
    rst = {
        "node_id": str(obj["node_id"]),
        "timestamp": float(obj["timestamp"]) // 1e9,
        "bytes_size": int(obj["bytes_size"]),
    }
    return rst


def readBandwithMetrics(path: Path) -> Dict:
    metrics = defaultdict(list)
    json_files = [pos_json for pos_json in os.listdir(path) if pos_json.endswith('bandwidth.log')]
    for json_file in json_files:
        with open(os.path.join(path, json_file), "r") as f:
            for line in f:
                parsed_json = parseMetrics(line)
                metrics[parsed_json["node_id"]].append(parsed_json)
    for val in metrics.values():
        val.sort(key=lambda x: x["timestamp"])
    return metrics


def mapToSize(wins: List) -> List:
    return [*map(lambda x: x["bytes_size"], wins)]


def calcBandwidth(sizes: List[int]) -> int:
    return sum(sizes)


def windowMetrics(metrics: Dict) -> Dict:
    rstDict = {}
    for node_id, js in metrics.items():
        if len(js) == 0 or len(js) == 1:
            rstDict[node_id] = js
            continue

        start, end = js[0], js[-1]
        startTime, endTime = start["timestamp"], end["timestamp"]
        timeDiff = endTime - startTime
        bucketsSize = math.ceil(timeDiff) + 1

        rst = [[] for _ in range(bucketsSize)]

        for metric in js:
            timestamp = metric["timestamp"]
            bucketsIndex = int(timestamp - startTime)
            rst[bucketsIndex].append(metric)
            
        rstDict[node_id] = rst
    return rstDict


def transformMetrics(metricWindows: Dict) -> Dict:
    rstDict = {}
    for node_id, windows in metricWindows.items():
        rst = []
        for metricWindow in windows:
            sizes = mapToSize(metricWindow)
            rst.append(
                {
                    "bandwidth": calcBandwidth(sizes),
                }
            )
        rstDict[node_id] = rst
    return rstDict


def reportMetrics():
    pathStr = sys.argv[1]
    path = Path(pathStr)
    metrics = readBandwithMetrics(path)
    metricsWindows = windowMetrics(metrics)
    reportDict = transformMetrics(metricsWindows)
    for node_id, reports in reportDict.items():
        pprint(reports)
        sns.set_theme()
        sns.set_context("paper")
        df = pd.DataFrame(
            {
                "time": np.arange(len(reports)),
                "bandwidth": [x["bandwidth"] for x in reports],
            }
        )
        f = sns.relplot(x="time", y="bandwidth", kind="line", data=df)
        f.set_axis_labels(x_var="time: second", y_var="bandwidth: bytes per second")
        df.style.set_caption(f"{node_id} Bandwith")
        plt.show()
        plt.savefig(f"{node_id}_bandwith.png")


if __name__ == "__main__":
    reportMetrics()
