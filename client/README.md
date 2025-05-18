# 📜 CookieFarm Client Exploitation Guide

Welcome to the CookieFarm client documentation! This guide explains how to create, run, and manage your exploits with CookieFarm's client and exploit manager.



# 👨‍💻 Client Overview

The **client** component of CookieFarm is responsible for:
- Retrieving flags through exploits
- Communicating with the server to submit obtained flags
- Handling multithreading and execution timing automatically

It is designed to let you **focus purely on writing exploits**, without worrying about concurrent execution, scheduling, or networking.



# 📊 Exploit Manager

The **exploit manager** is a Python utility that provides a decorator to simplify the exploitation process. It handles:

- Executing your exploit across all adversary machines
- Running exploits periodically (every "tick")
- Managing concurrency automatically with Python coroutines
- Proper output flushing and error handling

## 🔍 How it Works

Here is the basic structure of an exploit:

```python
#!/usr/bin/env python3

from utils.exploiter_manager import exploit_manager
import requests

@exploit_manager
def exploit(ip: str, port: int, name: str):
    r = requests.get(f"http://{ip}:{port}/get-flag")
    flag = r.text()
    return flag

if __name__ == "__main__":
    port = 4512  # Example port
    exploit(port=port)
```

✅ The `exploit_manager` automatically:
- Supplies IPs and ports
- Sends retrieved flags to the CookieFarm server
- Manages execution cycles and threads

> **Your only task: write the exploit logic!**

---

# 🚀 Running Your Exploit

Follow these steps to run your exploit with the client:

1. Navigate to the client directory:

   ```bash
   cd CookieFarm/client
   ```

2. Inside the `exploits/` folder, create your exploit script following the structure explained above.

3. Execute the following command:

   ```bash
   cookieclient -e ./<exploit_name>.py -b <server_url> -P <server_password> -p <service_port> -t <tick_time> -T <thread_count> -d
   ```

### 🔍 Command Arguments

| Argument | Description | Deafult |
|:---------|:------------|:--------|
| `-e`, `--exploit` | Path to your exploit file (must be inside `exploits/` folder) | N/A |
| `-b`, `--base_url_server` | Base URL and port of the CookieFarm server | N/A |
| `-P`, `--password` | Password for server authentication | N/A |
| `-p`, `--port` | Port of the service | N/A |
| `-t`, `--tick` | Frequency in seconds to re-run the exploit and submit flags | 120 |
| `-T`, `--threads` | Number of threads to use for concurrent execution | 5 |
| `-D`, `--debug` | Enable debug mode | False |
| `-d`, `--detach` | Detach mode | False |

### 📂 Example Run

```bash
cookieclient -e ./my_exploit.py -b http://10.10.23.1:8080 -P Str0ng_p4ssw0rd -p 8080 -t 120 -T 5 -d
```

This example runs `my_exploit.py` in debug mode every 120 seconds using 5 threads, sending the obtained flags to `http://10.10.23.1:8080`, using the password `Str0ng_p4ssw0rd`.
