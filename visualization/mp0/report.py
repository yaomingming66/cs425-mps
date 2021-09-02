import json
import pathlib
from pathlib import Path
from typing import Dict, List
import statistics
import numpy as np
import math
from pprint import pprint
import sys
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt


def parseMetrics(line: str) -> Dict:
    obj = json.loads(line)
    rst = {
        "timestamp": float(obj["timestamp"]),
        "delay": float(obj["delay"]),
        "size": int(obj["size"]),
    }
    return rst


def readMetrics(path: Path) -> List[Dict]:
    metrics = []
    with path.open() as f:
        for line in f:
            metrics.append(parseMetrics(line))
    metrics.sort(key=lambda x: x["timestamp"])
    return metrics


def mapToDelay(wins: List) -> List:
    return [*map(lambda x: x["delay"], wins)]


def mapToSize(wins: List) -> List:
    return [*map(lambda x: x["size"], wins)]


def calcMaxDelay(delays: List[float]) -> float:
    return math.nan if len(delays) == 0 else max(delays)


def calcMinDelay(delays: List[float]) -> float:
    return math.nan if len(delays) == 0 else min(delays)


def calcMedianDelay(delays: List[float]) -> float:
    return math.nan if len(delays) == 0 else statistics.median(delays)


def calc90PercentileDelay(delays: List[float]) -> float:
    arr = np.array(delays)
    return math.nan if len(delays) == 0 else np.percentile(arr, 90)


def calcBandwidth(sizes: List[int]) -> int:
    return sum(sizes)


def windowMetrics(metrics: List) -> List[List]:
    if len(metrics) == 0 or len(metrics) == 1:
        return metrics

    start, end = metrics[0], metrics[-1]
    startTime, endTime = start["timestamp"], end["timestamp"]
    timeDiff = endTime - startTime
    bucketsSize = math.ceil(timeDiff)

    rst = [[] for _ in range(bucketsSize)]

    for metric in metrics:
        timestamp = metric["timestamp"]
        bucketsIndex = int(timestamp - startTime)
        rst[bucketsIndex].append(metric)

    return rst


def transformMetrics(metricWindows: List[List]) -> List:
    rst = []
    for metricWindow in metricWindows:
        delays = mapToDelay(metricWindow)
        sizes = mapToSize(metricWindow)
        rst.append(
            {
                "minimumDelay": calcMinDelay(delays),
                "maximumDelay": calcMaxDelay(delays),
                "medianDelay": calcMedianDelay(delays),
                "90thPercentileDelay": calc90PercentileDelay(delays),
                "bandwidth": calcBandwidth(sizes),
            }
        )
    return rst


def reportMetrics():
    pathStr = sys.argv[1]
    path = Path(pathStr)
    metrics = readMetrics(path)
    metricsWindows = windowMetrics(metrics)
    reports = transformMetrics(metricsWindows)
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

    df = pd.DataFrame(
        {
            "time": np.arange(len(reports)),
            "90thPercentileDelay": [x["90thPercentileDelay"] for x in reports],
            "maximumDelay": [x["maximumDelay"] for x in reports],
            "minimumDelay": [x["minimumDelay"] for x in reports],
            "medianDelay": [x["medianDelay"] for x in reports],
        }
    )
    df = pd.melt(df, ["time"])
    g = sns.relplot(
        x="time", y="value", hue="variable", style="variable", kind="line", data=df
    )
    g.set_axis_labels(x_var="time: second", y_var="delay: second")
    plt.show()


if __name__ == "__main__":
    reportMetrics()
