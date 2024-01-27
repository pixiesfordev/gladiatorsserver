@REM 部屬完server後可以查看pod部屬狀況 kubectl get pods -n gladiators-gameserver -o wide
@REM 查看部屬的描述(如果部屬失敗可以用來查原因) kubectl describe gameserver gladiators-matchgame -n gladiators-gameserver

call Dev_SwitchProject.bat
@if ERRORLEVEL 1 exit /b 1
kubectl apply -f .\CICD_Matchgame_Dev\Dev_fleet.yaml
@if ERRORLEVEL 1 exit /b 1
kubectl apply -f .\CICD_Matchgame_Dev\Dev_fleetautoscaler.yaml
@if ERRORLEVEL 1 exit /b 1
