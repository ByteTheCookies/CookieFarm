import re

from typing import Callable, Any
from .logger import StatusCode, log_status
from .flagids_handler import FlagIdsHandler


class TestExploitExecutor:
    def __init__(
        self,
        exploit_function: Callable,
        port_service: int,
        name_service: str,
        flag_regex: re.Pattern,
        nop_team: tuple[int, str],
        url_flag_ids: str
    ):
        self.exploit_function = exploit_function
        self.port_service = port_service
        self.name_service = name_service
        self.flag_regex = flag_regex
        self.nop_team = nop_team
        self.handler_flag_ids = FlagIdsHandler(
            url=url_flag_ids,
            name_service=name_service
        )


    def __extract_flag(self, source: str | list[str]) -> set[str]:
        if isinstance(source, list):
            if len(source) == 0 or any(isinstance(item, str) for item in source):
                return set()

            return set(self.flag_regex.findall("\n".join(source)))

        elif isinstance(source, str):
            return set(self.flag_regex.findall(source))

        return set()


    def __single_exploit_execution(
        self,
        team_id: int,
        ip: str,
        flag_ids: list[dict[str, Any]]
    ):
        try:
            result = self.exploit_function(ip, self.port_service, self.name_service, flag_ids)

            if result != "":
                log_status(StatusCode.EXPLOIT_INFO, result)

            return self.__extract_flag(result)

        except Exception as e:
            log_status(
                StatusCode.ERROR,
                f"Error while exploiting team {team_id} ({ip}:{self.port_service}): {str(e)}",
                team_id=team_id,
                port_service=self.port_service,
                name_service=self.name_service,
            )
            return None


    def execute(self):
        flag_id = self.handler_flag_ids.get_flag_ids()
        flags = self.__single_exploit_execution(self.nop_team[0], self.nop_team[1], flag_id.get(self.nop_team[0], []))

        if flags:
            for flag in flags:
                log_status(
                    StatusCode.SUCCESS,
                    "Flag found in NOP team",
                    team_id=self.nop_team[0],
                    port_service=self.port_service,
                    name_service=self.name_service,
                    flag_code=flag,
                )

            log_status(
                StatusCode.INFO,
                message="Exploit test completed successfully",
                team_id=self.nop_team[0],
                port_service=self.port_service,
                name_service=self.name_service,
            )
        else:
            log_status(
                StatusCode.FAILED,
                "No flags found for NOP team",
                team_id=self.nop_team[0],
                port_service=self.port_service,
                name_service=self.name_service,
            )
