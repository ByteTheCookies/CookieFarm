#!/usr/bin/env python3

import time
import random
import string

from flask import Flask
from flask import request
from flask.templating import render_template

app = Flask(__name__)
server_number = random.randint(1, 100)

def generate_random_flag():
    allowed_chars = string.ascii_uppercase + string.digits
    random_part = ''.join(random.choice(allowed_chars) for _ in range(31))
    return random_part + "="

# ============ Service #1 ============

@app.route('/get-flag/')
def get_flag():
    flag = generate_random_flag()
    return flag

# ============ Service #2 ============

@app.route('/get-flag-sleep/')
def get_flag_sleep():
    flag = generate_random_flag()
    time.sleep(random.randint(1, 10))
    return flag

# ============ Service #3 ============

@app.route('/get-flag-guess/', methods=['POST'])
def get_flag_guess_post():
    user_guess = request.form.get('number', type=int)

    if user_guess is None or not (1 <= user_guess <= 100):
        return "Please send a valid number between 1 and 100"

    if user_guess == server_number:
        flag = generate_random_flag()
        return f"Congratulations! The number was {server_number}. Here's your flag: {flag}"
    else:
        return f"Sorry, the number was {server_number}. Try again!"

# ============ Service #4 ============

@app.route('/get-flag-maze/')
def get_flag_maze():
    chain_length = random.randint(2, 10)
    chain_id = ''.join(random.choice(string.ascii_lowercase) for _ in range(8))
    next_url = f"/get-flag-maze/{chain_id}/1/{chain_length}"
    return f"<html><body><p>Starting maze with {chain_length} steps.</p><p><a href='{next_url}'>Continue to step 1</a></p></body></html>"


@app.route('/get-flag-maze/<chain_id>/<int:step>/<int:chain_length>')
def get_flag_maze_step(chain_id, step, chain_length):
    if step >= chain_length:
        flag = generate_random_flag()
        return f"<html><body><p>Congratulations! You've reached the end of the maze.</p><p>Here's your flag: {flag}</p></body></html>"
    else:
        next_url = f"/get-flag-maze/{chain_id}/{step+1}/{chain_length}"
        return f"<html><body><p>You are at step {step} of {chain_length}.</p><p><a href='{next_url}'>Continue to step {step+1}</a></p></body></html>"

# ============ Service #5 ============

@app.route('/get-flag-somewhere/')
def get_flag_somewhere():
    try:
        with open('text.txt', 'r') as file:
            text = file.read()

        flag = generate_random_flag()

        if text:
            insert_position = random.randint(0, len(text))
            text_with_flag = text[:insert_position] + flag + text[insert_position:]
            return text_with_flag
        else:
            return "Error: Text file is empty"
    except FileNotFoundError:
        return "Error: Text file not found"


# ============ Service #6 ============
@app.route('/get-more-flags/', methods=['GET'])
def get_more_flag():
    flags = [generate_random_flag() for _ in range(5)]
    return "<br>".join(flags)

@app.route('/')
def index():
    return render_template("index.html")

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8081, debug=False)
