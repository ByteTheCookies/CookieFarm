## Benchmark average cpu/ram

## 1 test

## 30 machines

- Tool: SysScope
- Duration: 30 minutes
- Client: 1
- Exploits: example_1.py
- Notes: example_1.py is a simple example of an exploit that can be used to test the system.
- Debug: False
- Client Args: -e example_1.py -p password -b http://localhost:8080 -t 5

### Results

Cookieserver:
  - Max RAM: 49MB
  - Max CPU: 1.1%
  - Average RAM: 46MB
  - Average CPU: 0.0%

Cookieclient:
  - Max RAM: 21MB
  - Max CPU: 0.0%
  - Average RAM : 20MB
  - Average CPU: 0.0%

Exploiter:
  - Max RAM: 36MB
  - Max CPU: 3.9%
  - Average RAM : 36MB
  - Average CPU: 0.1%

- Total flags: 10962
- Unsubmitted flags: 232
- Percent of not submitted flags: 2.1%

## 2 test

## 30 machines

- Tool: SysScope
- Duration: 30 minutes
- Client: 5
- Exploits: example_1.py example_2.py example_3.py example_4.py example_5.py
- Notes: example_1.py is a simple example of an exploit that can be used to test the system.
- Debug: False
- Client Args: -e example_1.py -p password -b http://localhost:8080 -t 5


Cookieserver:
  - Max RAM: 55MB
  - Max CPU: 1.2%
  - Average RAM: 51MB
  - Average CPU: 0.0%

Cookieclient (Media 5 client):
  - Max RAM: 21MB
  - Max CPU: 0.0%
  - Average RAM : 20MB
  - Average CPU: 0.0%

- Total flags: 52838
- Unsubmitted flags: 18243
- Percent of not submitted flags: 34.5%
