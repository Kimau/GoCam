@echo off
for /r %%a in (_camA*.jpg) do (
    echo file '%%a' >> images.txt
)
E:\ffmpeg\bin\ffmpeg.exe -r 4 -f concat -i images.txt -c:v libx264 -pix_fmt yuv420p camA.mp4
del /q images.txt
del _camA*.jpg

@echo off
for /r %%a in (_camB*.jpg) do (
    echo file '%%a' >> images.txt
)
E:\ffmpeg\bin\ffmpeg.exe -r 4 -f concat -i images.txt -c:v libx264 -pix_fmt yuv420p camB.mp4
del /q images.txt
del _camB*.jpg