FOR /f "tokens=*" %%i IN ('docker ps -q') DO docker kill %%i
docker build -t companion_ai --build-arg value=%cd% .
docker run -v /var/run/docker.sock:/var/run/docker.sock -v %cd%:/companionAI/mnt -d -p 8080:8080 companion_ai