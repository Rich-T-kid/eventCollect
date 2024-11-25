import os
from slack_sdk import WebClient
from dotenv import load_dotenv

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
        print("error sending message:", e)

if __name__ == "__main__":
    send_message()