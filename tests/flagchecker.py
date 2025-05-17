#!/usr/bin/env python3

from flask import Flask, request
import time
import random

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

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5001)
