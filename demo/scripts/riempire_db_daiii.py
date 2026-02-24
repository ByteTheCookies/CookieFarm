import random
import uuid
import time
import json
import string
from shitcurl import send_post_request, login

def random_flag_code(length=30):
    charset = string.ascii_uppercase + string.digits
    return "FLAG{" + ''.join(random.choices(charset, k=length)) + "}"

# login
s = login("password")

headers = {
    'Content-Type': 'application/json',
    'Cookie': f"token={s.cookies['token']}"
}

batch_size = 4_000
total_flags = 100_000

for i in range(total_flags // batch_size):
    print(f"Processing batch {i+1}")

    # crea flags freschi per ogni batch
    flags_batch = []
    for _ in range(batch_size):
        flags_batch.append({
            "status": "unsubmitted",
            "id": str(uuid.uuid4()),
            "team_id": random.randint(1, 80),
            "port_service": random.randint(1, 65535),
            "service_name": "diocane",
            "flag_code": random_flag_code(50),
            "response_time": 0,
            "submit_time": random.randint(1, 1000000000)
        })

    body = json.dumps({"flags": flags_batch})  # JSON corretto per il batch

    res = send_post_request("api/v1/submit-flags", headers, body)
    if res and res.status_code == 200:
        print("Batch sent successfully!")
    else:
        print(f"Failed to send batch {i+1}: {res.status_code if res else 'No Response'}")
