# 🛠️ CookieFarm Server Guide

Welcome to the documentation of the **CookieFarm server**!
This guide describes how to launch, configure, and use the server responsible for flag management.

---

## 🌐 Server Overview

The **CookieFarm server** performs the following tasks:

- Receives and stores flags submitted by clients into a database.
- Sends flags to the `flagchecker` and handles the results.
- Displays all flags via a web interface:
  - Whether they have been submitted.
  - Whether they have been accepted.

The server is written in **Go** and is designed for easy deployment in both development and production environments.

---

## 🚀 Running the Server

To run the server use docker with:
```bash
docker compose up --build
```

---

## ⚙️ Execution Options

To configure the server, you need to create a `.env` file with the following parameters. You can use the `.env.example` file as a reference to set up your configuration.

| Environment Variable | Description                                                          | Default      |
|----------------------|----------------------------------------------------------------------|--------------|
| `DEBUG`              | Enables debug mode when set to `true`                                | `false`      |
| `CONFIG_PATH`        | Path to a YAML config file (instead of using the web form)           | N/A          |
| `PASSWORD`    | Password to access the server web interface                          | `"password"` |
| `PORT`        | Sets the port the server will listen on                              | `8080`       |

The YAML config file as be like that:
```YAML
configured: true

server:
  host_flagchecker: "<ip_flagchecker>:<port_flagchecker>"
  team_token: "<your_team_token>"
  submit_flag_checker_time: 120
  max_flag_batch_size: 1000
  protocol: "cc_http"
  tick_time: 120
  start_time: <start_time>
  end_time: <end_time>
  flag_ttl: 5 # in ticks (if the ttl is 0, the flag will never expire)

client:
  services:
    - name: "CookieService"
      port: 8081
  format_ip_teams: "10.10.{}.1"
  regex_flag: "[A-Z0-9]{31}="
  range_ip_teams: 29
  my_team_id: 1
  nop_team: 0
  url_flag_ids: "http://localhost:5001/flagIds"
```


> **⚠️ WARNING: Security Risk!**
>
> You are **strongly encouraged** to change the default password (`"password"`) to a strong, unique password. Using the default password poses a significant security risk as it could allow unauthorized access to your flag management system!

---

## 🌐 Web Interface

The **web interface** is accessible at:

```
http://<your_server_ip>:<port>
```

Through the UI you can:

- View all received flags.
- Check the submission and acceptance status of each flag.
- Configure the server (unless you setup the configuration from YAML file).

---

## 📂 Example Usage

### Setting up .env file

Create a `.env` file with the following content:

```
DEBUG=true
CONFIG_PATH=./config.yml
PASSWORD=SuperSecret
PORT=9090
```

### Running with Docker

```bash
docker compose up
```

This configuration runs the server:

- In debug mode.S
- With password `SuperSecret`.
- On port `9090`.
- Using the configuration file `./config.yml`.

---

Happy flag cathing! 🎯
