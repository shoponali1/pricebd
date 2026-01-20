@echo off
echo Updating Gold Prices...
go run main.go
echo.
echo Adding to git...
git add fe/src/prices.csv fe/src/silver-prices.csv
git commit -m "Manual update from Windows"
git push
echo.
echo Done! You can close this window.
pause
