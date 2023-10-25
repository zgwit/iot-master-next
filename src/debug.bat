@echo off

cls

echo.
echo --- --- --- --- --- compile.. --- --- --- --- --- ---
echo.

go build -o iot-master-pro.exe

    if %errorlevel% neq 0 (

    echo.
    echo --- --- --- --- --- compile fail! --- --- --- --- ---
    echo.

    exit /b 1

)
move iot-master-pro.exe ..

cd ..

echo.
echo --- --- --- --- --- running --- --- --- --- --- ---
echo.

iot-master-pro.exe

echo.
echo --- --- --- --- --- end --- --- --- --- --- ---
echo.

cd %~dp0
