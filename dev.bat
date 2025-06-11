@echo off
cd frontend
echo Running npm build...
call npm run build
cd ..
echo Starting Wails dev server...
call wails dev
