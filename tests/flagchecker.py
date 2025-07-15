#!/usr/bin/env python3

from flask import Flask, request

import sys
import time
import random
from string import ascii_letters

app = Flask(__name__)

# Define constants for status codes
ACCEPTED = "flag claimed"
DENIED = ("invalid flag", "flag from nop team", "flag is your own", "flag too old","flag already claimed","the check which dispatched this flag didn't terminate successfully")
RESUBMIT = "the flag is not active yet, wait for next round"
ERROR = "notify the organizers and retry later"

status = {
    "ACCEPTED": ACCEPTED,
    "DENIED": DENIED,
    "RESUBMIT": RESUBMIT,
    "ERROR": ERROR,
}

flag_store = set()
num_team = sys.argv[1] if len(sys.argv) > 1 else 10

def generate_random_string(length=16):
    """
    Generates a random string of fixed length.
    """
    return ''.join(random.choice(ascii_letters) for _ in range(length))


@app.route("/flags", methods=['PUT'])
def check_flags():
    responses = []

    time.sleep(random.randint(0, 2))

    if request.headers.get('X-Team-Token'):
        data = request.get_json()
        if data:
            for flag in data:
                if flag in flag_store:
                    s = "RESUBMIT"
                    message = status["RESUBMIT"]
                else:
                    s = random.choice(list(status.keys()))
                    match s:
                        case "ACCEPTED":
                            message = status["ACCEPTED"]
                            flag_store.add(flag)
                        case "DENIED":
                            message = status["DENIED"]
                        case _:
                            message = status["ERROR"]

                responses.append({
                    "msg": f"[{flag}] {random.choice(message) if s == "DENIED" else  message}",
                    "flag": flag,
                    "status": s
                })
    return responses


@app.route("/flagIds", methods=['GET'])
def get_flag_ids():
    """
    Returns a list of flag IDs.
    """
    example_flag_ids = {
        "CookieService": {},
    }

    for service_name, service_data in example_flag_ids.items():
        for i in range(int(num_team)):
            if random.randint(0, 1) == 0:
                example_flag_ids[service_name].update({
                    f"{i}": {
                        "0": {
                            "username": generate_random_string(8),
                            "password": generate_random_string(16)
                        },
                        "1": {
                            "username": generate_random_string(8),
                            "password": generate_random_string(16)
                        },
                        "2": {
                            "username": generate_random_string(8),
                            "password": generate_random_string(16)
                        }
                    }
                })

    return example_flag_ids


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5001)
