"""
Logger module for logging status messages.

"""

import json
from enum import Enum


class StatusCode(Enum):
    """Enumeration of possible status codes for logging."""

    SUCCESS = "success"  # Case: get flag successfully
    ERROR = "error"  # Case: get flag unsuccessfully [error in exploit returns an error]
    FAILED = "failed"  # Case: not get the flag [exploit returns no flag]
    FATAL = "fatal"  # Case: fatal error that requires immediate attention. Stop script's execution
    DEBUG = "debug"  # Case: debug message, useful for developers
    INFO = "info"  # Case: informative message
    STATS = "stats"  # Case: statistics message
    EXPLOIT_INFO = "exploit_info"  # Case: information about the exploit process


def log_status(
    status: StatusCode,
    message: str,
    team_id: int = 0,
    port_service: int = 0,
    flag_code: str = "",
    name_service: str = "",
) -> bool:
    """
    Logs a status message with additional information.

    Args:
        status (StatusCode): The status code to log.
        message (str): The message to log.
        team_id (int, optional): The ID of the team. Defaults to 0.
        port_service (int, optional): The port number of the service. Defaults to 0.
        flag_code (str, optional): The flag code. Defaults to "".
        name_service (str, optional): The name of the service. Defaults to "".

    Returns:
        bool: True if the status was logged successfully, False otherwise.
    """

    if not isinstance(status, StatusCode):
        return False

    print(
        json.dumps(
            {
                "status": status.value,
                "message": message,
                "team_id": team_id,
                "port_service": port_service,
                "flag_code": flag_code,
                "name_service": name_service,
            }
        ),
        flush=True,
    )
    return True


def log_stats(
    message: str,
    port_service: int,
    name_service: str,
    stats: dict[str, int]
):

    """
    Logs statistics message.

    Args:
        message (str): The message to log.
        port_service (int): The port number of the service.
        name_service (str): The name of the service.
        stats (dict[str, int]): A dictionary containing statistics.

    Returns:
        bool: True if the statistics were logged successfully, False otherwise.
    """
    print(
        json.dumps(
            {
                "status": StatusCode.STATS.value,
                "message": message,
                "port_service": port_service,
                "name_service": name_service,
                "total_flag": stats.get("total_flag", 0),
                "success_team": stats.get("success_team", 0),
                "failed_team": stats.get("failed_team", 0),
            }
        ),
        flush=True,
    )
    return True
