from .data_models import SettingsConf
import json


settings = SettingsConf.parse_file('./settings.json')
