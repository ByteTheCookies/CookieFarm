import re
import sys
import requests

from typing import Any
from array import array
from pathlib import Path
from dataclasses import dataclass
from .logger import StatusCode, log_status


@dataclass
class ServerConfigInfo:
    regex_flag: re.Pattern
    teams: dict[int, str]
    nop_team: tuple[int, str]
    url_flag_ids: str


class ServerConfig:
    def __init__(self, server_address: str, config_api_path: str = "/api/v1/config"):
        self.server_address = server_address
        self.config_api_path = config_api_path
        self.raw_config = self.__get_raw_config()

    def __get_auth_token(self) -> str:
        """
        Get the authentication token from the session file.

        Returns:
            str: The authentication token.
        """
        session_path = Path.home().joinpath(".config", "cookiefarm", "session")

        if not session_path.exists():
            log_status(StatusCode.FATAL, "Session file not found. Please login first.")
            sys.exit(1)

        with open(session_path) as session_file:
            token = session_file.read().strip()

        return token

    def __get_raw_config(self) -> dict[str, Any]:
        """
        Get the configuration from the server.

        Returns:
            dict[str, Any]: The configuration dictionary.
        Raises:
            ValueError: If the response from the server is not ok.
        """
        response = requests.get(
            f"http://{self.server_address}{self.config_api_path}",
            headers={
                "Content-Type": "application/json",
                "Cookie": f"token={self.__get_auth_token()}",
            },
        )
        if not response.ok:
            log_status(
                StatusCode.FATAL, "Failed to retrieve configuration from the server."
            )
            sys.exit(1)

        return response.json()

    def __generate_teams_ip(self, range_team: int, nop_team_id: int, my_team_id: int, format_ip_teams: str) -> dict[int, str]:
        """
        Generate a dictionary of team IDs and their corresponding IP addresses.

        Args:
            range_team (int): The total number of teams.
            nop_team_id (int): The NOP team ID.
            my_team_id (int): My team ID.
            format_ip_teams (str): The format string for IP addresses.

        Returns:
            dict[int, str]: A dictionary mapping team IDs to their IP addresses.
        """

        if my_team_id < nop_team_id:
            x_1 = my_team_id
            x_2 = nop_team_id
        else:
            x_1 = nop_team_id
            x_2 = my_team_id

        ip_teams = array("H", range(x_1))
        for i in range(x_1 + 1, x_2):
            ip_teams.append(i)

        for i in range(x_2 + 1, range_team + 1):
            ip_teams.append(i)

        return {i: format_ip_teams.format(i) for i in ip_teams}


    def config(self) -> ServerConfigInfo:
        """
        Get the configuration information from the server.

        Returns:
            ServerConfigInfo: A ServerConfigInfo object containing regex_flag and teams.
        Raises:
            ValueError: If the response from the server is not ok.
        """
        config_json = self.raw_config

        regex_flag = re.compile(config_json["client"]["regex_flag"])
        if not regex_flag:
            log_status(StatusCode.FATAL, "Regex flag is not defined in the configuration.")
            sys.exit(1)

        log_status(
            StatusCode.DEBUG,
            f"Regex flag: {regex_flag.pattern}"
        )

        format_ip_teams: str = config_json["client"]["format_ip_teams"]
        if not format_ip_teams:
            log_status(StatusCode.FATAL, "Format IP teams is not defined in the configuration.")
            sys.exit(1)

        log_status(
            StatusCode.DEBUG,
            f"Format IP teams: {format_ip_teams}"
        )

        my_team = int(config_json["client"]["my_team_id"])
        if not my_team:
            log_status(StatusCode.FATAL, "My team ID is not defined in the configuration.")
            sys.exit(1)

        log_status(
            StatusCode.DEBUG,
            f"My team ID: {my_team}"
        )

        nop_team: int = int(config_json["client"]["nop_team"])
        if not nop_team and nop_team != 0:
            log_status(StatusCode.FATAL, "NOP team is not defined in the configuration.")
            sys.exit(1)

        log_status(
            StatusCode.DEBUG,
            f"NOP team ID: {nop_team}"
        )

        url_flag_ids: str = config_json["client"]["url_flag_ids"]
        if not url_flag_ids:
            log_status(StatusCode.FATAL, "URL flag IDs is not defined in the configuration.")
            sys.exit(1)

        log_status(
            StatusCode.DEBUG,
            f"URL flag IDs: {url_flag_ids}"
        )

        range_team: int = int(config_json["client"]["range_ip_teams"])
        if not range_team:
            log_status(StatusCode.FATAL, "Range IP teams is not defined in the configuration.")
            sys.exit(1)

        log_status(
            StatusCode.DEBUG,
            f"Range IP teams: {range_team}"
        )

        ip_teams = self.__generate_teams_ip(
            range_team, nop_team, my_team, format_ip_teams
        )

        return ServerConfigInfo(
            re.compile(regex_flag),
            ip_teams,
            (nop_team, format_ip_teams.format(nop_team)),
            url_flag_ids
        )
