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
import subprocess
import matplotlib.pyplot as plt
from collections import *
from subprocess import Popen

def main():
    pathStr = sys.argv[1]
    path = Path(pathStr)
    json_files = [pos_json for pos_json in os.listdir(path) if pos_json.endswith('.log')]
    for json_file in json_files:
        abspath_json = os.path.join(path, json_file)
        cmd = f"grep 'metrics.bandwidth' {abspath_json}"
        with open(os.path.join(path, json_file[0] + "_bandwidth.log"), "wb") as f:
            process_grep_and_save = Popen(cmd, stdout=f, shell=True)
            process_grep_and_save.wait()
            f.flush()
        print(f"bandwidth: {json_file}")
        
        cmd = f"grep 'metrics.delay' {abspath_json}"
        with open(os.path.join(path, json_file[0] + "_delay.log"), "wb") as f:
            process_grep_and_save = Popen(cmd, stdout=f, shell=True)
            process_grep_and_save.wait()
            f.flush()
        print(f"delay: {json_file}")
            
if __name__ == "__main__":
    main()