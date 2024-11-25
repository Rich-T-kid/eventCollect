## Slack Bot

### What It Does
When you run the script, it performs the following actions:
1. Print a sucess message to the terminal: This ensures that the script is running correctly.
2. Sends a message to the designated Slack channel: The bot posts a text message to a Slack channel via the Slack API.

### Requirements
- Python 3.x
- pip to install dependencies
- Slack workspace and bot token

###Set up
1. Clone the repository
2. Install dependencies
Run the following command to install the necessary libraries:
pip3 install slack-sdk python-dotenv
3. Set up the Environment file
Create a .env file and add your Slack API token:
SLACK_API_TOKEN= your_slack_api_token
Replace your_slack_api_token with your actual token

- Token can be found after creating an app in Slack and navigating to OAuth & Permissions in left sidebar in the app settings
- Under Scopes add the scope chat:write. Then install the app to workspace 
- The Bot User OAuth Token will appear in the OAuth & 

4. Invite the bot to the specific channel on slack:
/invite @your-bot-name
5. Run the Script:
python3 slack_bot.py
