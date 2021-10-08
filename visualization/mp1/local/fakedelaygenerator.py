
from random import choice
import json

c = "ABCDEFGHIJKLMN"
li = []

with open("3_node_0_5_hz_100_d.json", "r") as f:
    for l in f:
        j = json.loads(l)
        j["messageID"] = choice([*c])
        li.append(j)
        # f.write(json.dumps(j))
        # json.dump(j, f)
        
with open("3_node_0_5_hz_100_d.json", "w") as f:
    for j in li:
        json.dump(j, f)
        f.write("\n")