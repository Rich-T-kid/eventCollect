import ssl
import certifi
from slack_sdk import WebClient
from dotenv import load_dotenv
import os


ssl_context = ssl.create_default_context(cafile=certifi.where())
ssl._create_default_https_context = ssl._create_unverified_context 

load_dotenv()
slack_token = os.getenv('SLACK_API_TOKEN')

client = WebClient(token=slack_token)

def send_message():
    try:
        response = client.chat_postMessage(
            channel='bot-test', 
            text="Hello from SLACK BOT!"
        )
        print("Message sent successfully: ", response['message']['text'])
    
    except Exception as e:
        print("Error sending message:", e)

if __name__ == "__main__":
    send_message()
