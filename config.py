import json
from settings_model import Settings

f = open('settings.json','r')
settings_dict = json.load(f)

settings = Settings(**settings_dict)
