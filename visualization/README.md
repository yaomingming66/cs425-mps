# Visualization

## Development Environment configuration

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## MP0

we want to track two metrics:

Delay from the time the event is generated to the time it shows up in the centralized logger
The amount of bandwidth used by the centralized logger

For the delay, you can just use the difference between the current time when you are about to print the event and the timestamp of the event itself. For measuring the bandwidth, you will need to track the length of all the messages received by the logger.

You should produce a graph of these two metrics over time.
For the bandwidth,
you should track the average bandwidth across each second of the experiment.
For the delay,
for each second you should plot the 
minimum, maximum, median, and 90th percentile delay at each second.
Make sure your graphs and axes are well labeled, with units.


## MP1

Record bandwith in json, named after xxx_b.json

Record delay in json, named after xxx_d.json


```json
{ "timestamp": "", "delay": "", "size": "" }
```
