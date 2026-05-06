#!/bin/bash

set -euo pipefail
set -o errtrace

# ══════════════════════════════════════════════════════════════════════════════
#  Colors (truecolor)
# ══════════════════════════════════════════════════════════════════════════════

BOLD=$'\033[1m'
RESET=$'\033[0m'

C_TITLE=$'\033[38;2;205;161;87m'
C_DIM=$'\033[38;2;136;136;136m'
C_FLAG=$'\033[38;2;33;150;243m'
C_GREEN=$'\033[38;2;33;155;84m'
C_ARGUMENT=$'\033[38;2;237;237;237m'
C_ERROR_FG=$'\033[38;2;237;237;237m'
C_ERROR_BG=$'\033[48;2;231;76;60m'
C_ERROR=$'\033[38;2;231;76;60m'

# Gum hex palette
GC_TITLE="#CDA157"
GC_DIM="#888888"
GC_BASE="#E9E9E9"
GC_ARGUMENT="#EDEDED"
GC_ERROR="#E74C3C"

# ══════════════════════════════════════════════════════════════════════════════
#  Error handling
# ══════════════════════════════════════════════════════════════════════════════

last_command=""
current_command=""
failed_wrapped_title=""
failed_wrapped_command=""
error_reported=0
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG

err_report() {
    local exit_code="${3:-$?}"
    local lineno="${1:-${BASH_LINENO[0]:-0}}"
    local command="${2:-${current_command:-${last_command:-unknown}}}"
    [[ "$exit_code" -eq 0 ]] && return 0
    error_reported=1
    printf "%b ERROR %b %s%b (exit %d, line %d)\n" \
        "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" \
        "${command:-unknown}" "$RESET" \
        "$exit_code" "$lineno" >&2
    if [[ -n "${failed_wrapped_command:-}" ]]; then
        printf "%b       %bSpinner task: %s\n" "$C_DIM" "$RESET" "${failed_wrapped_title:-unknown}" >&2
        printf "%b       %bWrapped command: %s\n" "$C_DIM" "$RESET" "$failed_wrapped_command" >&2
    fi
    [[ -x "${GUM_BIN:-}" ]] &&
        "$GUM_BIN" log --structured --level error --time rfc822 \
            "Command failed" \
            exit_code "$exit_code" line "$lineno" command "${command:-unknown}" \
            wrapped_title "${failed_wrapped_title:-}" wrapped_command "${failed_wrapped_command:-}" >&2 || true
}

trap 'err_report "$LINENO" "$BASH_COMMAND"' ERR
trap 'code=$?; [[ "$code" -ne 0 ]] && [[ "${error_reported:-0}" -eq 0 ]] && err_report "$LINENO" "${current_command:-unknown}" "$code"' EXIT

require_cmd() {
    local cmd="$1"
    local hint="${2:-}"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        printf "%b ERROR %b Missing required command: %b%s%b%s\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" \
            "${BOLD}${C_FLAG}" "$cmd" "$RESET" \
            "${hint:+ — $hint}" >&2
        exit 2
    fi
}

# ══════════════════════════════════════════════════════════════════════════════
#  Gum bootstrap
# ══════════════════════════════════════════════════════════════════════════════

gum_bin() {
    local repo="charmbracelet/gum"
    local api="https://api.github.com/repos/${repo}/releases/227167086"
    local os arch version asset url tmpdir tarball extracted_bin bindir

    os="$(uname -s)"
    arch="$(uname -m)"

    case "$os" in
    Linux) os="Linux" ;;
    Darwin) os="Darwin" ;;
    *)
        printf "%b ERROR %b Unsupported OS: %s\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" "$os" >&2
        return 1
        ;;
    esac

    case "$arch" in
    x86_64 | amd64) arch="x86_64" ;;
    aarch64 | arm64) arch="arm64" ;;
    armv7l | armv7) arch="armv7" ;;
    *)
        printf "%b ERROR %b Unsupported architecture: %s\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" "$arch" >&2
        return 1
        ;;
    esac

    require_cmd curl "Install curl to download gum"
    require_cmd tar "Install tar to extract the archive"
    require_cmd mktemp
    require_cmd find

    version="$(curl -fsSL "$api" | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p' | head -n1)"
    if [[ -z "${version:-}" ]]; then
        printf "%b ERROR %b Could not fetch gum version from GitHub\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" >&2
        return 1
    fi

    asset="gum_${version}_${os}_${arch}.tar.gz"
    url="https://github.com/${repo}/releases/download/v${version}/${asset}"

    tmpdir="$(mktemp -d /tmp/gum.XXXXXX)"
    tarball="${tmpdir}/${asset}"
    bindir="${tmpdir}/bin"
    mkdir -p "$bindir"

    printf "%b  →%b Downloading gum %b%s%b...\n" \
        "$C_DIM" "$RESET" "${BOLD}${C_TITLE}" "$version" "$RESET" >&2
    curl -fL --progress-bar "$url" -o "$tarball"
    tar -xzf "$tarball" -C "$tmpdir"

    extracted_bin="$(find "$tmpdir" -maxdepth 3 -type f -name gum -print -quit 2>/dev/null || true)"
    if [[ -z "${extracted_bin:-}" || ! -f "$extracted_bin" ]]; then
        printf "%b ERROR %b gum binary not found in downloaded archive\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" >&2
        return 1
    fi

    cp "$extracted_bin" "${bindir}/gum"
    chmod +x "${bindir}/gum"
    printf '%s\n' "${bindir}/gum"
}

ensure_gum() {
    # Fast path: already cached
    if [[ -x /tmp/gum-bin/gum ]]; then
        printf '%s\n' /tmp/gum-bin/gum
        return 0
    fi

    # Search for a previously extracted binary
    local found
    found="$(find /tmp -maxdepth 4 -type f -name gum -executable -path '/tmp/gum.*' \
        -print -quit 2>/dev/null || true)"
    if [[ -n "${found:-}" ]]; then
        mkdir -p /tmp/gum-bin
        cp "$found" /tmp/gum-bin/gum
        chmod +x /tmp/gum-bin/gum
        printf '%s\n' /tmp/gum-bin/gum
        return 0
    fi

    # Download fresh
    local bin
    bin="$(gum_bin)"
    if [[ -z "${bin:-}" || ! -x "$bin" ]]; then
        printf "%b ERROR %b gum is not available and could not be installed\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" >&2
        return 1
    fi
    mkdir -p /tmp/gum-bin
    cp "$bin" /tmp/gum-bin/gum
    chmod +x /tmp/gum-bin/gum
    printf '%s\n' /tmp/gum-bin/gum
}

# ══════════════════════════════════════════════════════════════════════════════
#  Gum themed helpers
# ══════════════════════════════════════════════════════════════════════════════

gum_section() {
    "$GUM_BIN" style \
        --border normal \
        --border-foreground "$GC_TITLE" \
        --foreground "$GC_BASE" \
        --bold \
        --padding "0 1" \
        "$1"
}

gum_input() {
    local header="$1"
    shift
    "$GUM_BIN" input \
        --header "$header" \
        --header.foreground "$GC_TITLE" \
        --prompt "> " \
        --prompt.foreground "$GC_DIM" \
        --cursor.foreground "$GC_TITLE" \
        "$@"
}

# Prompts for a numeric value and re-asks until the input is a valid integer.
gum_input_int() {
    local header="$1"
    local default="$2"
    local value
    while true; do
        value="$(gum_input "$header" --value "$default" --placeholder "$default")"
        if printf '%s' "$value" | grep -Eq '^[0-9]+$'; then
            printf '%s' "$value"
            return 0
        fi
        printf "%b⚠%b  '%s' is not a valid number — please try again\n" \
            "$C_TITLE" "$RESET" "$value" >&2
    done
}

gum_confirm() {
    "$GUM_BIN" confirm \
        --prompt.foreground "$GC_BASE" \
        --selected.background "$GC_TITLE" \
        --selected.foreground "#0A0C0D" \
        --unselected.foreground "$GC_DIM" \
        "$@"
}

gum_choose() {
    "$GUM_BIN" choose \
        --header.foreground "$GC_TITLE" \
        --cursor.foreground "$GC_DIM" \
        --cursor="🍪 " \
        "$@"
}

format_command() {
    local quoted="" part q
    for part in "$@"; do
        printf -v q '%q' "$part"
        quoted="${quoted}${quoted:+ }${q}"
    done
    printf '%s' "$quoted"
}

gum_spin() {
    local title="$1"
    shift
    local cmd_display exit_code

    cmd_display="$(format_command "$@")"
    failed_wrapped_title=""
    failed_wrapped_command=""

    set +e
    "$GUM_BIN" spin \
        --spinner minidot \
        --title "$title" \
        --spinner.foreground "$GC_TITLE" \
        --title.foreground "$GC_BASE" \
        -- "$@"
    exit_code=$?
    set -e

    if [[ "$exit_code" -ne 0 ]]; then
        failed_wrapped_title="$title"
        failed_wrapped_command="${cmd_display:-<empty>}"
        printf "%b ERROR %b Spinner task failed (exit %d)\n" \
            "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" "$exit_code" >&2
        printf "%b       %bTask: %s\n" "$C_DIM" "$RESET" "$failed_wrapped_title" >&2
        printf "%b       %bWrapped command: %s\n" "$C_DIM" "$RESET" "$failed_wrapped_command" >&2
        [[ -x "${GUM_BIN:-}" ]] &&
            "$GUM_BIN" log --structured --level error --time rfc822 \
                "Spinner task failed" \
                exit_code "$exit_code" title "$failed_wrapped_title" command "$failed_wrapped_command" >&2 || true
        return "$exit_code"
    fi
}

gum_log_info() {
    local msg="$1"
    shift || true
    "$GUM_BIN" log --structured --level info "$msg" "$@" >&2
}

gum_log_debug() {
    local msg="$1"
    shift || true
    "$GUM_BIN" log --structured --level debug "$msg" "$@" >&2
}

gum_log_error() {
    local msg="$1"
    shift || true
    "$GUM_BIN" log --structured --level error --time rfc822 "$msg" "$@" >&2
}

# ══════════════════════════════════════════════════════════════════════════════
#  State globals
# ══════════════════════════════════════════════════════════════════════════════

PORT=""
PASSWORD=""
CONFIG_FILE=""
DEBUG=""
VERSION="v$(git ls-remote --tags --refs https://github.com/ByteTheCookies/CookieFarm.git 2>/dev/null | sed -n 's#.*refs/tags/##p' | sed 's/^v//' | sort -V | tail -n1)"

preconfigured="false"

url_flag_checker=""
team_token=""
submit_flag_checker_time=""
max_flag_batch_size=""
protocol=""
tick_time=""
flag_ttl=""
start_time=""
end_time=""
services_yaml=""
range_ip_teams=""
format_ip_teams=""
my_team_id=""
regex_flag=""
nop_team=""
url_flag_ids=""
flagids_format=""

# Tracked output paths (set during write_env / gum_ask_config)
ENV_FILE=""
CFG_FILE=""

# ══════════════════════════════════════════════════════════════════════════════
#  Section 1 — Basic .env
# ══════════════════════════════════════════════════════════════════════════════

gum_ask_basic() {
    gum_section "  BASIC CONFIGURATION  "

    PORT="$(gum_input_int "Server PORT" "8080")"
    PASSWORD="$(gum_input "Server PASSWORD" --value "password" --placeholder "password")"

    if gum_confirm "Use a config file (CONFIG_FILE)?"; then
        CONFIG_FILE="$(gum_input "CONFIG_FILE path" --value "config.yml" --placeholder "config.yml")"
        if [[ -z "${CONFIG_FILE:-}" ]]; then
            CONFIG_FILE="config.yml"
        fi
    else
        CONFIG_FILE="false"
    fi

    if gum_confirm "Enable DEBUG mode?"; then
        DEBUG="true"
    else
        DEBUG="false"
    fi

    gum_log_debug "Basic configuration captured" \
        port "$PORT" config_file "$CONFIG_FILE" debug "$DEBUG"
}

# ══════════════════════════════════════════════════════════════════════════════
#  Section 2 — config.yml
# ══════════════════════════════════════════════════════════════════════════════

gum_ask_config() {
    local dest="${1:-cookiefarm/config.yml}"

    # Normalise: if dest is a directory or ends in /, append filename
    if [[ -d "$dest" || "${dest: -1}" = "/" || "$dest" = "." ]]; then
        dest="${dest%/}/config.yml"
    fi
    CFG_FILE="$dest"

    [[ "${CONFIG_FILE}" != "false" ]] || return 0

    gum_section "  CONFIGURATION FILE  "

    if gum_confirm "Do you already have a preconfigured config.yml?"; then
        local cfg_path
        cfg_path="$(gum_input "Path to your existing config.yml" \
            --value "./config.yml" --placeholder "./config.yml")"
        if [[ -f "$cfg_path" ]]; then
            mkdir -p "$(dirname "$dest")"
            cp "$cfg_path" "$dest"
            printf "%b✔%b  Copied config → %b%s%b\n" \
                "$C_GREEN" "$RESET" "${BOLD}${C_ARGUMENT}" "$dest" "$RESET"
            gum_log_debug "Copied preconfigured config" source "$cfg_path" dest "$dest"
            preconfigured="true"
        else
            printf "%b⚠%b  Path not found — switching to interactive setup\n" \
                "$C_TITLE" "$RESET"
            gum_log_error "Config path not found" path "$cfg_path"
            preconfigured="false"
        fi
    fi

    [[ "$preconfigured" = "true" ]] && return 0

    # ── Server settings ──────────────────────────────────────────────────────
    gum_section "  SERVER SETTINGS  "

    local checker_default="http://10.10.10.1:8081/flags"
    url_flag_checker="$(gum_input "url_flag_checker" \
        --value "$checker_default" --placeholder "$checker_default")"
    team_token="$(gum_input "team_token" --placeholder "your-team-token")"
    submit_flag_checker_time="$(gum_input_int "submit_flag_checker_time (seconds)" "120")"
    max_flag_batch_size="$(gum_input_int "max_flag_batch_size" "1000")"
    protocol="$(gum_input "protocol" --value "cc_http" --placeholder "cc_http")"
    tick_time="$(gum_input_int "tick_time (seconds)" "120")"
    flag_ttl="$(gum_input_int "flag_ttl (ticks)" "0")"

    local start_time_default
    start_time_default="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    start_time="$(gum_input "start_time (ISO 8601 UTC)" \
        --value "$start_time_default" --placeholder "$start_time_default")"

    local end_time_default
    if command -v python3 >/dev/null 2>&1; then
        end_time_default="$(python3 -c "
from datetime import datetime, timezone, timedelta
print((datetime.now(timezone.utc) + timedelta(hours=8)).strftime('%Y-%m-%dT%H:%M:%SZ'))
")"
    else
        case "$(uname -s)" in
        Linux) end_time_default="$(date -u -d "+8 hours" +"%Y-%m-%dT%H:%M:%SZ")" ;;
        Darwin) end_time_default="$(date -u -v+8H +"%Y-%m-%dT%H:%M:%SZ")" ;;
        *) end_time_default="$start_time_default" ;;
        esac
    fi
    end_time="$(gum_input "end_time (ISO 8601 UTC, default now+8h)" \
        --value "$end_time_default" --placeholder "$end_time_default")"

    # ── Services ─────────────────────────────────────────────────────────────
    gum_section "  SERVICES  "

    local num_services
    num_services="$(gum_input_int "How many services to configure?" "3")"

    services_yaml=""
    if [[ "$num_services" -gt 0 ]]; then
        local i=1
        while [[ "$i" -le "$num_services" ]]; do
            local svc_name svc_port
            svc_name="$(gum_input "Service #${i} name" \
                --value "Service${i}" --placeholder "ServiceName")"
            svc_port="$(gum_input_int "Port for ${svc_name}" "808${i}")"
            services_yaml="${services_yaml}      ${svc_name}: ${svc_port}"$'\n'
            i=$((i + 1))
        done
    fi

    # ── Teams / Network ───────────────────────────────────────────────────────
    gum_section "  TEAMS & NETWORK  "

    range_ip_teams="$(gum_input_int "range_ip_teams" "40")"
    format_ip_teams="$(gum_input "format_ip_teams" \
        --value "10.10.{}.1" --placeholder "10.10.{}.1")"
    my_team_id="$(gum_input_int "my_team_id" "1")"
    regex_flag="$(gum_input "regex_flag" \
        --value "[A-Z0-9]{31}=" --placeholder "[A-Z0-9]{31}=")"
    nop_team="$(gum_input_int "nop_team" "0")"
    url_flag_ids="$(gum_input "url_flag_ids" \
        --value "http://10.10.10.1:8081/flagIds" \
        --placeholder "http://<ip>:8081/flagIds")"

    # Flag IDs Format Selection
    flagids_format_choice="$(gum_choose \
        --header "Select Flag IDs Format Template" \
        --limit 1 \
        "CyberChallenge Template" \
        "Faust Template" \
        "Custom")"

    case "$flagids_format_choice" in
    "CyberChallenge Template")
        flagids_format="[service].[team].[id]"
        ;;
    "Faust Template")
        flagids_format="flag_ids.[service].[team].[id]"
        ;;
    "Custom")
        flagids_format="$(gum_input "flagids_format" \
            --value "[service].[team].[id]" \
            --placeholder "Enter custom template (e.g., custom.[service].[teams].[id])")"
        ;;
    esac

    # ── Write config.yml ──────────────────────────────────────────────────────
    mkdir -p "$(dirname "$dest")"
    gum_log_debug "Creating config file" path "$dest"

    cat >"$dest" <<EOF
configured: true

# Server
server:
  url_flag_checker: "${url_flag_checker}"
  team_token: "${team_token}"
  submit_flag_checker_time: ${submit_flag_checker_time}
  max_flag_batch_size: ${max_flag_batch_size}
  protocol: "${protocol}"
  tick_time: ${tick_time}
  flag_ttl: ${flag_ttl}
  start_time: "${start_time}"
  end_time: "${end_time}"

# Client
shared:
  services:
${services_yaml}  range_ip_teams: ${range_ip_teams}
  format_ip_teams: "${format_ip_teams}"
  my_team_id: ${my_team_id}
  regex_flag: "${regex_flag}"
  nop_team: ${nop_team}
  url_flag_ids: "${url_flag_ids}"
  flagids_format: "${flagids_format}"
EOF
    printf "%b✔%b  Wrote %b%s%b\n" \
        "$C_GREEN" "$RESET" "${BOLD}${C_ARGUMENT}" "$dest" "$RESET"
}

# ══════════════════════════════════════════════════════════════════════════════
#  Write .env
# ══════════════════════════════════════════════════════════════════════════════

write_env() {
    local dest="${1:-cookiefarm/.env}"

    if [[ -d "$dest" || "${dest: -1}" = "/" ]]; then
        dest="${dest%/}/.env"
    fi
    ENV_FILE="$dest"

    mkdir -p "$(dirname "$dest")"
    gum_log_debug "Creating .env file" path "$dest"

    cat >"$dest" <<EOF
PORT=${PORT}
PASSWORD=${PASSWORD}
CONFIG_FILE=${CONFIG_FILE}
DEBUG=${DEBUG}
VERSION=${VERSION}
EOF

    printf "%b✔%b  .env written → %b%s%b  PORT=%b%s%b  CONFIG_FILE=%b%s%b  DEBUG=%b%s%b\n" \
        "$C_GREEN" "$RESET" \
        "${BOLD}${C_ARGUMENT}" "$dest" "$RESET" \
        "$C_FLAG" "$PORT" "$RESET" \
        "$C_FLAG" "$CONFIG_FILE" "$RESET" \
        "$C_FLAG" "$DEBUG" "$RESET"
}

# ══════════════════════════════════════════════════════════════════════════════
#  Review & confirm before deployment
# ══════════════════════════════════════════════════════════════════════════════

confirm() {
    gum_section "  REVIEW CONFIGURATION  "

    if [[ -f "$CFG_FILE" ]]; then
        "$GUM_BIN" style '── config.yml ──' --italic --bold --foreground "$GC_TITLE"
        cat "$CFG_FILE" | "$GUM_BIN" format -t code -l yaml
    fi

    "$GUM_BIN" style '── .env ──' --italic --bold --foreground "$GC_TITLE"
    cat "$ENV_FILE" | "$GUM_BIN" format -t code -l env

    printf "\n"
    "$GUM_BIN" style \
        --foreground "$GC_TITLE" \
        --italic \
        "You can edit the files directly and re-run this installer if needed."
    printf "\n"

    # Optional: open editor
    if gum_confirm "Open a file in ${EDITOR:-vi} before deploying?"; then
        local chosen_file
        chosen_file="$("$GUM_BIN" file . --all --cursor="🍪")"
        if [[ -n "${chosen_file:-}" && -f "$chosen_file" ]]; then
            "${EDITOR:-vi}" "$chosen_file"
        else
            printf "%b⚠%b  No file selected — skipping editor\n" "$C_TITLE" "$RESET"
        fi
    fi

    # Final go/no-go
    if ! gum_confirm "Proceed with deployment?"; then
        printf "%b✘%b  Deployment cancelled\n" "$C_ERROR" "$RESET"
        exit 0
    fi
}

# ══════════════════════════════════════════════════════════════════════════════
#  Entry point
# ══════════════════════════════════════════════════════════════════════════════

printf "%b%b  🍪  CookieFarm Installer  %b\n\n" "$BOLD" "$C_TITLE" "$RESET"

require_cmd bash
require_cmd printf
require_cmd sed
require_cmd grep
require_cmd find
require_cmd cp
require_cmd curl "Install curl to enable downloading dependencies"

GUM_BIN="$(ensure_gum)" || {
    printf "%b ERROR %b Failed to obtain gum binary\n" \
        "${BOLD}${C_ERROR_BG}${C_ERROR_FG}" "$RESET" >&2
    exit 1
}

mode=$(gum_choose remote build --header Mode)

# ── Build mode: clone + docker build ─────────────────────────────────────────
if [[ "$mode" = "build" ]]; then
    require_cmd git "Install git to clone the repository"
    require_cmd docker "Install Docker to run containers"

    if [[ -d CookieFarm ]]; then
        printf "%b⚠%b  CookieFarm directory already exists — skipping clone\n" \
            "$C_TITLE" "$RESET"
    else
        gum_spin "Cloning CookieFarm repository..." \
            git clone https://github.com/ByteTheCookies/CookieFarm.git
    fi

    cd CookieFarm
    gum_ask_basic
    write_env      # → cookiefarm/.env
    gum_ask_config # → cookiefarm/config.yml
    confirm

    cd cookiefarm
    gum_spin "Building and starting containers (this may take a while)..." \
        sh -c 'docker compose -f compose.build.yml up --build -d 2>&1'

# ── Compose-only mode: download + docker up ───────────────────────────────────
else
    require_cmd wget "Install wget to download compose.yml"
    require_cmd docker "Install Docker to run containers"

    mkdir -p CookieFarm
    cd CookieFarm

    gum_spin "Downloading compose.yml..." \
        wget -q -O compose.yml https://raw.githubusercontent.com/ByteTheCookies/CookieFarm/dev/cookiefarm/compose.yml

    gum_ask_basic
    write_env .      # → ./.env
    gum_ask_config . # → ./config.yml
    confirm

    gum_spin "Starting containers..." \
        sh -c 'docker compose -f compose.yml pull && docker compose -f compose.yml up --build --pull always -d 2>&1'
fi

# ── Success banner ────────────────────────────────────────────────────────────
printf "\n"
"$GUM_BIN" style \
    --border double \
    --border-foreground "$GC_TITLE" \
    --foreground "$GC_ARGUMENT" \
    --bold \
    --padding "1 4" \
    --align center \
    "🍪  CookieFarm is up and running!"
printf "\n"
