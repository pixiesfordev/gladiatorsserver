@echo off
REM 使用此批次檔切換到Dev環境 在powershell中執行.\Dev_SwitchProject.bat
@echo on
gcloud config set project fourth-waters-410202
gcloud config set container/cluster cluster-gladiators
gcloud container clusters get-credentials cluster-gladiators --zone=asia-east1-c
kubectl config current-context



