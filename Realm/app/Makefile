# 版本建置Makefile

# 安裝最新的專案Module
InstallModule:
	@echo "==============Start Installing Module=============="
	cd functions && npm install pixies-mygladiators@latest

# 推送App
PushApp:
	@echo "==============Start Pushing App=============="
	realm-cli push --include-package-json

# 重新連結aura-gladiatorfishin
RelinkModule:
	@echo "=============Relink npm module=============="
	cd functions && npm link pixies-mygladiators
# 完整發布流程
all: InstallModule PushApp RelinkModule