# 📜 CookieFarm Client Exploitation Guide

## 🔧 Prerequisites

Before you begin, make sure you have:
- **Python 3+**
- **pip** for installing Python modules
- **Modern terminal** that supports ANSI colors and Unicode (for TUI mode)

---

## ⚙️ Client Installation

To install the CookieFarm client, run the following command:

```bash
pip install cookiefarm
```

After installation, the `ckc` command will be globally available in your terminal. By default, it will launch in interactive TUI (Text User Interface) mode.

To uninstall the client:

```bash
pip uninstall cookiefarm
```

---

## 📺 Interface Options

The CookieFarm client has two interface modes:

1. **Interactive TUI Mode** (default): A colorful, user-friendly interface with menus and keyboard navigation
   ```bash
   ckc
   # TUI starts automatically
   ```

2. **Traditional CLI Mode**: For scripts, automation, or environments where TUI isn't supported
   ```bash
   ckc --help
   ```

> [!NOTE]
> If you add params at the end of the command, they will be parsed as CLI commands. For example:
> ```bash
> ckc config show
> # This will run the `config show` command in CLI mode.
> ```
## 🚀 Client Command Overview

### Configuration Commands

| Command                     | Description                                     | Detailed Usage |
|----------------------------|-------------------------------------------------| -------------------------------------------------|
| `config login`             | Log in to the server with a password            | [Section](#config-login-command) |
| `config update`            | Update client config with IP, port, user, etc.  | [Section](#config-update-command) |
| `config show`              | Show current configuration                      | [Section](#config-show-command) |
| `config reset`             | Reset configuration to default                  | [Section](#config-reset-command) |
| `config logout`            | Log out from the server                         | [Section](#config-logout-command) |

### Exploit Management Commands

| Command                      | Description                                        | Detailed Usage |
|-----------------------------|----------------------------------------------------| -------------------------------------------------|
| `exploit create`            | Create a new exploit template                      | [Section](#exploit-create-command) |
| `exploit test`              | Test an exploit template (exploit the NOP team)     | [Section](#exploit-test-command) |
| `exploit run`               | Run an exploit against a target                   | [Section](#exploit-run-command) |
| `exploit list`              | List all currently running exploits                | [Section](#exploit-list-command) |
| `exploit remove`            | Remove an exploit template                         | [Section](#exploit-remove-command) |
| `exploit stop`              | Stop a running exploit                             | [Section](#exploit-stop-command) |

---

## 🧪 Exploitation Workflow

1. **Install the client** using pip:
   ```bash
   pip install cookiefarm
   ```

2. **Log in** to the server:
   ```bash
   # In CLI mode (with no environment variable setted):
   ckc config login -P SuperSecret -H 192.168.1.10 -p 8000 -u CookieMonster

   # In TUI mode:
   # Navigate to: Configuration → Login → Enter credentials
   ```

3. **Create a new exploit template**:
   ```bash
   # In CLI mode (with no environment variable setted):
   ckc exploit create -n my_exploit

   # In TUI mode:
   # Navigate to: Exploits → Create Exploit → Enter name
   ```

4. **Write your exploit** in the created file `~/.config/cookiefarm/exploits/my_exploit.py`.

5. **Run the exploit**:
   ```bash
   # In CLI mode (with no environment variable setted):
   ckc exploit run -e my_exploit.py -p 1234 -t 120 -T 40
   ```

---

## 🌟 TUI Navigation

The interactive TUI provides easy navigation with the following keyboard shortcuts:

| Key          | Action                         |
|-------------|--------------------------------|
| ↑/↓ or j/k  | Navigate menu items            |
| Enter       | Select item or submit form     |
| ESC         | Go back to previous screen     |
| Tab         | Navigate between input fields  |
| q or Ctrl+C | Quit the application           |

The TUI offers these main views:
- **Main Menu**: Choose between Configuration and Exploit operations
- **Config Menu**: Configuration management commands
- **Exploit Menu**: Exploit management commands
- **Input Forms**: Fill in required parameters for commands
- **Output View**: See command results with syntax highlighting

---

## Detailed Command Usage

### Config Login Command
Authenticate with the server using a password:
```bash
# In CLI mode (with no environment variable setted):
ckc config login -P <password> -H <server_ip> -p <port> -u <username>

# In TUI mode:
# Navigate to: Configuration → Login → Enter password
```
Parameters:
- `-P <password>`: The password for the server. This is required for authentication.
- `-H <server_ip>`: IP address of the server.
- `-p <port>`: Port of the server.
- `-u <username>`: Username for the client. (default is `cookieguest`)

### Config Update Command
Update the client configuration (all fields optional, at least one required):
```bash
# In CLI mode (with no environment variable setted):
ckc config update -H <server_ip> -p <port> -u <username> [-s]

# In TUI mode:
# Navigate to: Configuration → Update Config → Fill the form
```
Parameters:
- `-H <server_ip>`: IP address of the server.
- `-p <port>`: Port of the server.
- `-u <username>`: Username for the client. (default is `guest`)
- `-s`: Use secure connection (HTTPS).

### Config Show Command
Display the current configuration:
```bash
# In CLI mode (with no environment variable setted):
ckc config show

# In TUI mode:
# Navigate to: Configuration → Show Config
```

### Config Reset Command
Reset the configuration to default:
```bash
# In CLI mode (with no environment variable setted):
ckc config reset

# In TUI mode:
# Navigate to: Configuration → Reset Config
```

### Config Logout Command
Log out and clear the current session:
```bash
# In CLI mode (with no environment variable setted):
ckc config logout

# In TUI mode:
# Navigate to: Configuration → Logout
```

---

### Exploit Create Command
Create a new exploit template:
```bash
# In CLI mode (with no environment variable setted):
ckc exploit create -n <exploit_name>

# In TUI mode:
# Navigate to: Exploits → Create Exploit → Enter name
```
Parameters:
- `-n <exploit_name>`: Name of the exploit template. This can be a path to a file or just a name.
> If user does not specify the path in the exploit name, it will be created in the `~/.config/cookiefarm/exploits/` directory. Otherwise, it will be created in the specified path.

*Example:*
```bash
# In CLI mode (with no environment variable setted):
ckc exploit create -n ./my_exploit

# In TUI mode:
# Navigate to: Exploits → Create Exploit → Enter name as `./my_exploit`
```
In this case, the exploit will be created in the current directory.


### Exploit Test Command
Test an exploit against the NOP team:
```bash
# In CLI mode (with no environment variable setted):
ckc exploit test -e <exploit_file> -n <service_name> [-t <timeout>] [-T <threads>]
```
Parameters:
- `-e <exploit_file>`: Path to the exploit file (Python script).
- `-n <service_name>`: Name of the service to test the exploit against.
- `-t <timeout>`: Timeout for the exploit in seconds (default is 120).
- `-T <threads>`: Number of threads to use (default is 10).

### Exploit Run Command
Run an exploit:
```bash
# In CLI mode (with no environment variable setted):
ckc exploit run -e <exploit_file> -n <service_name> [-t <timeout>] [-T <threads>] [-D]

```
Parameters:
- `-e <exploit_file>`: Path to the exploit file (Python script).
- `-n <service_name>`: Name of the service to run the exploit against.
- `-t <timeout>`: Timeout for the exploit in seconds (default is 120).
- `-T <threads>`: Number of threads to use (default is 10).
- `-D`: Enable debug mode for more verbose output.

*Example:*
```bash
# In CLI mode (with no environment variable setted):
ckc exploit run -e my_exploit.py -n CookieService -t 120 -T 40
# This will return the PID of the running exploit.
```

> [!IMPORTANT]
> With the command `ckc exploit test` you can see all the print statements from the exploit script, which is useful for debugging and understanding how the exploit works, in the `ckc exploit run` command no.

### Exploit List Command
List all running exploits:
```bash
# In CLI mode (with no environment variable setted):
ckc exploit list

# In TUI mode:
# Navigate to: Exploits → List Running Exploits
```

### Exploit Remove Command
Remove a saved exploit template:
```bash
# In CLI mode (with no environment variable setted):
ckc exploit remove -n <exploit_name>

# In TUI mode:
# Navigate to: Exploits → Remove Exploit → Enter exploit name
```
Parameters:
- `-n <exploit_name>`: Name of the exploit template to remove. This can be a path to a file or just a name.

### Exploit Stop Command
Stop a running exploit:
```bash
# In CLI mode (with no environment variable setted):
ckc exploit stop -p <pid>

# In TUI mode:
# Navigate to: Exploits → Stop Exploit → Enter PID
```
Parameters:
- `-p <pid>`: Process ID of the exploit to stop.
