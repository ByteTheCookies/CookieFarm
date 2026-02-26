#!/usr/bin/env python3

import requests
import json
import logging

s = requests.Session()

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
BASE_URL = "http://localhost:8080"

def send_post_request(endpoint, headers=None, data=None, files=None):
    url = f"{BASE_URL}/{endpoint}"
    try:
        logging.info(f"Sending POST request to {url} with data: {data}")
        response = s.post(url, headers=headers, data=data, files=files,cookies=s.cookies)
        response.raise_for_status()
        logging.info(f"Response received: {response.status_code}")
        return response
    except requests.exceptions.HTTPError as http_err:
        logging.error(f"HTTP error occurred: {http_err}")
    except Exception as err:
        logging.error(f"Other error occurred: {err}")
    return None

def login(password):
    logging.info("Logging in with provided password...")
    payload = {'password': password}
    headers = {'Content-Type': 'application/x-www-form-urlencoded'}
    response = send_post_request('api/v1/auth/login', headers=headers, data=payload)

    if response:
        if response.status_code == 200:
            logging.info("Login successful, token received.")
            return response
        else:
            logging.warning("Token not found in the response.")
            return None
    else:
        logging.error("Failed to send login request.")
        return None

def configure(config_data):
    logging.info("Configuring the system with provided configuration data...")
    headers = {
        'Content-Type': 'application/json',
        'Cookie': f"token={s.cookies['token']}"
    }
    payload = json.dumps(config_data)
    response = send_post_request('api/v1/config', headers=headers, data=payload)

    if response:
        logging.info("Configuration updated successfully.")
        return response.json()
    return None

if __name__ == '__main__':
    password = 'password'
    login(password)


    config_data = json.load(open('config.json', 'r'))
    config_data = {
        "config": config_data
    }
    configure(config_data)
