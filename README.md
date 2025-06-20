<div align="center">
  <img width="360px" height="auto" src="assets/logo_mucca.png" alt="CookieFarm Logo">
</div>

<p align="center">
  <img src="https://img.shields.io/badge/relase-1.2.0-red?style=flat-square" alt="Version">
  <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/ByteTheCookies/CookieFarm?filename=.%2Fclient%2Fgo.mod&style=flat-square">
  <img alt="GitHub code size in bytes" src="https://img.shields.io/github/languages/code-size/ByteTheCookies/CookieFarm?color=7289DA&style=flat-square">
  <img alt="GitHub License" src="https://img.shields.io/github/license/ByteTheCookies/CookieFarm?color=orange&style=flat-square">
</p>

# 🍪 CookieFarm

**CookieFarm** is an *Attack/Defense CTF* framework inspired by DestructiveFarm, developed by the Italian team **ByteTheCookies**.
Its strength lies in a hybrid **Go + Python** architecture and a **zero-distraction philosophy**:
> 🎯 *Your only task is to write the exploit!*

CookieFarm automates exploit distribution, flag submission, and result monitoring — allowing you to focus entirely on building powerful exploits.

---

## 🔧 Prerequisites

Make sure you have the following installed:

- ✅ Python 3+
- ✅ Docker

---

## 📁 Repository Structure

| Directory       | Description |
|------------------|-------------|
| [`client/`](./client/) | Handles exploit creation and flag submission |
| [`server/`](./server/) | Manages exploit distribution, flag collection, and monitoring |

---

## ⚙️ Architecture Overview

<div align="center">
  <img width="800px" height="auto" src="assets/arch_farm.png" alt="Architecture Diagram">
</div>

---

## ▶️ Getting Started

### 🖥️ Starting the Server

1. Create an `.env` file in the server directory to configure the environment settings:

    ```bash
      # Server configuration
      DEBUG=false                   # Enable debug mode for verbose logging
      PASSWORD=SuperSecret  # Set a strong password for authentication
      CONFIG_FROM_FILE=config.yml  # Set if the server takes the config from config.yml in the filesystem; otherwise, do not set the variable
      PORT=8080            # Define the port the server will listen on
    ```

  > ⚠️ For production environments, set `DEBUG=false` and use a strong, unique password

2. Start the server with Docker Compose:
   ```bash
   docker compose up --build
   ```

📘 For more configuration details, refer to the [server documentation](./docs/server/README.md).

---

### 💻 Using the Client & Running Exploits

1. Run the installation :
  ```bash
  pip install cookiefarm
  ```

  > After installation, the `ckc` command is available globally in your terminal (or in your virtual environment if you are using one).

2. Log in and configure the client:
   ```bash
   ckc config login -P SuperSecret -h 192.168.1.10 -p 8000 -u your_username
   ```

3. Install the Python helper module and create a new exploit template:
   ```bash
   ckc exploit create -n your_exploit_name
   ```

   This will generate `your_exploit_name.py` in `~/.cookiefarm/exploits/`.

4. Run your exploit:
   ```bash
   ckc exploit run -e your_exploit_name.py -p 1234 -t 120 -T 40
   ```

📘 For more usage examples, check out the [client documentation](./docs/client/README.md).

---

## 🤝 Contributing

We welcome contributions, suggestions, and bug reports!
See [CONTRIBUTING.md](./CONTRIBUTING.md) for details on how to get involved.


## 📈 Star History

<a href="https://star-history.com/#ByteTheCookies/CookieFarm&Date&secret=Z2hwX1AzVkd6OTFZR2h1RkZWNjJHZnplTTFZZU1Yb3pHMTFKeHlDdw==">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=ByteTheCookies/CookieFarm&type=Date&theme=dark&secret=Z2hwX1AzVkd6OTFZR2h1RkZWNjJHZnplTTFZZU1Yb3pHMTFKeHlDdw==" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=ByteTheCookies/CookieFarm&type=Date&secret=Z2hwX1AzVkd6OTFZR2h1RkZWNjJHZnplTTFZZU1Yb3pHMTFKeHlDdw==" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=ByteTheCookies/CookieFarm&type=Date&secret=Z2hwX1AzVkd6OTFZR2h1RkZWNjJHZnplTTFZZU1Yb3pHMTFKeHlDdw==" />
 </picture>
</a>

<div align="center">
  Built with ❤️ by <strong>ByteTheCookies</strong>
</div>
