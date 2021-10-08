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
    msg = json.loads(json.loads(line)["msg"])
    return str(msg["msg_id"]), float(msg["timestamp"])


def readDelaywithMetrics(path: Path) -> Dict:
    metrics = defaultdict(lambda: (float("inf"), -float("inf")))
    json_files = [pos_json for pos_json in os.listdir(path) if pos_json.endswith('delay.log')]
    for json_file in json_files:
        with open(os.path.join(path, json_file), "r") as f:
            for line in f:
                messageID, timestamp = parseMetrics(line)
                early, late = metrics[messageID]
                metrics[messageID] = (min(timestamp, early), max(timestamp, late))
    return metrics

def mapToMaxDelay(metrics: Dict) -> Dict:
    return {k: v[1] - v[0] for k, v in metrics.items()}


def reportMetrics():
    pathStr = sys.argv[1]
    path = Path(pathStr)
    metrics = readDelaywithMetrics(path)
    delayMetrics = mapToMaxDelay(metrics)
    pprint(delayMetrics)
    sns.set_theme()
    sns.set_context("paper")
    df = pd.DataFrame(
        {
            "message": np.arange(len(delayMetrics)),
            "delay": [x for x in delayMetrics.values()],
        }
    )
    f = sns.relplot(x="message", y="delay", kind="scatter", data=df)
    f.set_axis_labels(x_var="message: message", y_var="delay: nano seconds")
    df.style.set_caption(f"Message Delay")
    plt.show()
    plt.savefig(f"delay.png")


if __name__ == "__main__":
    reportMetrics()
