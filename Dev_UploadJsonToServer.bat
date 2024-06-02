@echo off

set LOCAL_DIR=.\Realm\mymodule\JsonData\*
set GCS_BUCKET=gs://gladiators_test1_gamejson_dev

gsutil -m cp -r %LOCAL_DIR% %GCS_BUCKET%

pause
