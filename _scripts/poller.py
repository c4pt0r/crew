import os
import time
import json
import base64
import requests
import traceback
import jinja2

from datetime import datetime
from dateutil import tz

API_URL = 'http://<dbq-endpoint>/q/cat/pull'  # Replace with your dbq API endpoint, go to github.com/c4pt0r/dbq for details
MARKDOWN_FILE = '../logs.md'  # Replace with your Markdown file path
IMG_DIR = '../_static/imgs'   # Replace with your static directory path

MESSAGE_TEMPLATE = jinja2.Template("""
<!--- {{ id }} | {{ created_at }} --->
{{ data }}
<p style='text-align: right;'>@{{ created_at }}</p>

--------
<!--- {{ id }} | EOF --->

""")

def utc_to_local(utc_time_str):
    utc_time = datetime.strptime(utc_time_str, "%Y-%m-%dT%H:%M:%SZ")
    local_timezone = tz.tzlocal() 
    local_time = utc_time.replace(tzinfo=tz.gettz('UTC')).astimezone(local_timezone)
    return local_time

def render_message(data, id = None, created_at = None):
    return MESSAGE_TEMPLATE.render(data=data, id=id, created_at=created_at)

def fetch_data():
    try:
        response = None
        auth_token = os.getenv('AUTH_TOKEN')
        if auth_token != None and len(auth_token) > 0:
            headers =  {"Authorization": "Bearer " + auth_token}
            response = requests.get(API_URL, timeout=5, headers=headers)
        else:
            response = requests.get(API_URL, timeout=5)
    except requests.exceptions.Timeout:
        return None
    if response.status_code == 200:
        ret = response.json()
        if len(ret['payload']) > 0:
            payloads = ret['payload']
            if len(payloads) > 0:
                return payloads
    return None

def parse_message_data(data : dict):
    tp = data.get('type')
    if tp == 'text':
        return data.get('content')
    elif tp == 'image':
        fname = data.get('filename')
        content = base64.b64decode(data.get('content'))
        # write image to static directory
        with open(os.path.join(IMG_DIR, fname), 'wb') as file:
            file.write(content)
        # return markdown image link
        return f"![{fname}]({os.path.join(IMG_DIR, fname)})"
    elif tp == 'url':
        return f"[{data.get('content')}]({data.get('content')})"
    else:
        return None

def append_to_markdown(data, id, created_at):
    if not os.path.exists(MARKDOWN_FILE):
        with open(MARKDOWN_FILE, 'w', encoding='utf-8') as file:
            file.write(render_message(data, id, created_at))
        return

    with open(MARKDOWN_FILE, 'r', encoding='utf-8') as file:
        content = file.readlines()

    with open(MARKDOWN_FILE, 'w', encoding='utf-8') as file:
        file.write(render_message(data, id, created_at))
        file.writelines(content)

def main():
    while True:
        try:
            payload = fetch_data()
            if payload:
                for msg in payload:
                    decoded_data = base64.b64decode(msg['data']).decode()
                    # make sure the data is a json dict
                    data = json.loads(decoded_data)
                    decoded_data = parse_message_data(data)
                    if decoded_data:
                        append_to_markdown(decoded_data, msg['id'], msg['created_at'])
                        print("New data appended to Markdown file.")
            else:
                pass
        except Exception:
            traceback.print_exc()
            time.sleep(3)
        time.sleep(3)  # Wait for 3 seconds before the next API call

if __name__ == '__main__':
    main()
