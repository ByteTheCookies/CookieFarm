# 🍪 CookieFarm - Exploiter Manager

![Language](https://img.shields.io/badge/languages-Python-yellowgreen)
![Keywords](https://img.shields.io/badge/keywords-CTF%2C%20Exploiting%2C%20Attack%20Defense-red)
![License](https://img.shields.io/badge/license-MIT-blue)

> Python decorator for automating exploit execution in CTF Attack & Defense competitions

---

## 📦 What is it?

This package provides a **`@exploit_manager` decorator** designed to automate the parallel execution of exploits in CTF (*Attack & Defense*) settings, specifically for use with the [CookieFarm](https://github.com/BytesTheCookies/CookieFarm) project.

It handles:

* Authentication with the central server
* Retrieving team configuration
* Automatic flag parsing from `stdout`

> ⚠️ **Note:** This package is **not standalone**. It must be used together with the [CookieFarm client](https://github.com/BytesTheCookies/CookieFarm). The client provides the required APIs and team configurations.

---

## 📦 Installation

To install the package:
```bash
pip install cookiefarm
```


## ⚙️ How it works

The `@exploit_manager` decorator takes care of:

* Calling your `exploit(ip, port, name_service, flag_ids)` function
* Capturing your exploit's `stdout`
* Parsing flags via regex
* Logging the result in JSON format, including: team ID, port, service name, and the flag found


---

## 🚀 Example usage

```python
from cookiefarm import exploit_manager
import requests

@exploit_manager
def exploit(ip, port, name_service, flag_ids):
    # Run your exploit here
    response = requests.get(f"http://{ip}:{port}/")

    # Just print the flag to stdout
    print(response.text)

# Run from the command line with arguments from CookieFarm
# python3 myexploit.py <ip_server> <password> <tick_time> <thread_number> <port> <name_service>
```

For execution, you have to pass the required arguments from the command line, which are provided by the CookieFarm client. The decorator will handle the rest.

```bash

python3 myexploit.py -s <server_address> -t <tick_time> -T <thread_number> -p <port> -n <name_service> -x [test mode]
```

| Argument        | Description                                      |
|------------------|--------------------------------------------------|
| `-s` or `--server` | The address of the CookieFarm server             |
| `-t` or `--tick_time` | The time interval for the exploit execution      |
| `-T` or `--thread_number` | The number of threads to use for the exploit     |
| `-p` or `--port` | The port to target for the exploit                |
| `-n` or `--name_service` | The name of the service to exploit               |
| `-x` or `--test` | Run in test mode (does not execute the exploit)   |

---

## 🛠️ Requirements

* Python ≥ 3.12
* Working CookieFarm client installed


---

## 📝 License

Distributed under the [MIT License](LICENSE). Feel free to use, modify, and contribute.

---

For any questions, suggestions, or issues, feel free to open a [GitHub issue](https://www.github.com/BytesTheCookies/CookieFarmExploiter)!

**Created with ❤️ by ByteTheCookies (feat. @0xMatte)**
