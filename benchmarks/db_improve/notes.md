# Benchmark Notes

## 1-05-2025

## Goal

Check the perfomance differences between the current implementation of the Database and the new one using mattn + some query tweks. We will run the same exploit on both implementations and compare the results.

### Summary

- 40 hosts
- sha: 9c2ed70be7e5248f3f7fd2c4ff9ca35b7c6a17e5
- branch: improve/database
- cks version: 1.3.0
- ckc version: 1.3.0

The original is the dev branch with hash 2d9a804d4bb855d93c7f4eb68dd20b6ee8d7c6da

## CK config

### 9c2ed70be7e5248f3f7fd2c4ff9ca35b7c6a17e5

32 GB RAM
32 vCPU
LAN

- Build command: `just server-build-prod`, `just server-build-plugins-prod`
- Command run: `../bin/cks -c`
- Run command: `time python3 riempire_db_daiii.py`

### 2d9a804d4bb855d93c7f4eb68dd20b6ee8d7c6da

32 GB RAM
32 vCPU
LAN

- Build command: `just client-build-prod`
- Command run: `./bin/ckc exploit run -e benchmark -n CookieService -t 5 -W 10`
- Run command: `time python3 riempire_db_daiii.py`

### Cks config used:

```yaml
configured: true

# Server
server:
  url_flag_checker: "http://localhost:5001/flags"
  team_token: "pippo"
  submit_flag_checker_time: 30
  max_flag_batch_size: 5000
  protocol: "cc_http"
  tick_time: 120
  flag_ttl: 0 # in ticks
  start_time: "2023-10-01T00:00:00Z"
  end_time: "2023-10-31T23:59:59Z"

# Client
shared:
  services:
    CookieService: 8081
    vulnify: 1337
    app-nc: 1338
  range_ip_teams: 40
  format_ip_teams: "10.10.{}.1"
  my_team_id: 1
  regex_flag: "[A-Z0-9]{31}="
  nop_team: 0
  url_flag_ids: "http://localhost:5001/flagIds"
```

### Test config

```python
tickets_to_emulate = 150
window_seconds = 120
min_flags_per_window = 99
max_flags_per_window = 299

services = ["http", "ssh", "dns", "smtp", "ftp", "redis", "mysql", "postgres"]

exploits = [
    "sqli_blind",
    "rce_template_injection",
    "path_traversal",
    "auth_bypass",
    "deserialization_rce",
    "command_injection",
    "ssrf",
    "buffer_overflow",
]

batch_size = 4_000
base_submit_time = random.randint(1_700_000_000, 1_750_000_000)
```

### Results

#### Time of inserting flags in the database:

- 2ce6d275998a6cc73d6ec39ce378810c61ed1771:

CPU	5,3%
user	0,676
system	0,075
total	1,413

Total flags 35448 in 1,413 seconds so 25.008 flags/s

Fetch random page 40ms

- 2d9a804d4bb855d93c7f4eb68dd20b6ee8d7c6da:

CPU	0%
user	0,705
system	0,093
total	4:25,05

Total flags 37339 in 265,05 seconds so 140.8 flags/s

Fetch random page 0.23ms

But in the old implementation i use limit and offset to fetch the flags, so the time to fetch a random page grows with the number of flags in the database, while in the new implementation i use a id to fetch the flags, so the time to fetch a random page is constant. This is why in the new implementation it takes 40ms to 70ms to fetch a random page, while in the old implementation it takes 0.23ms but grows with the number of flags.

Old: 230k flags like from 170ms to 220ms
New: 250k flags from 70ms to 77ms
