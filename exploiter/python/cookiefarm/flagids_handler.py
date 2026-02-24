import requests

from typing import Any


class FlagIdsHandler:
    def __init__(self, url: str, name_service: str):
        self.url = url
        self.name_service = name_service

    def get_flag_ids(self) -> dict[int, list[Any]]:
        response = requests.get(
            f"{self.url}",
            params={"service": self.name_service},
            headers={"Content-Type": "application/json"}
        )

        if not response.ok:
            raise ValueError(f"Failed to retrieve flag IDs from {self.url}. Error: {response.text}")

        return self.__parser_flag_ids(response.json())


    def __parser_flag_ids(self, raw_flag_ids: dict) -> dict[int, list[Any]]:
        """
        Returns a dictionary mapping team IDs to their respective flag IDs.

        Returns:
            dict[int, list[Any]]: A dictionary where keys are team IDs and values are lists of flag IDs.
        """

        service_data = raw_flag_ids.get(self.name_service, None)
        flag_ids = {}

        if not service_data:
            raise ValueError(f"No data found for service: {self.name_service}")

        for team_id, rounds in service_data.items():
            try:
                team_id = int(team_id)
            except ValueError:
                continue

            team_flags = []
            for _, round_data in rounds.items():
                team_flags.append(round_data)

            flag_ids[team_id] = team_flags

        return flag_ids
