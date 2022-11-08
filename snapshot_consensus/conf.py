from .data_models import SettingsConf
import json


settings = SettingsConf.parse_file('snapshot_consensus/settings.json')
