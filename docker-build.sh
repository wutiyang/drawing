#/bin/bash

mode=$1
if [ "$mode" ==  "" ];then
        echo "help args  dev|test|stg|stage"
        exit 0
fi
pathName=`pwd`
echo ${pathName}

dockerDir=/app/go-drawing
appName=go-drawing-${mode}
appVersion=1.0

name=${appName}:${appVersion}
docker build -t ${name} -f Dockerfile_${mode} .

containerList=`docker ps -a | grep "${appName}" |awk '{print $1}'`
echo containerList

for containerId in ${containerList[@]}
do
  echo "id={$containerId}"
  docker rm -f ${containerId}
done

docker run -d --name ${appName} -p 9018:8007 -v ${pathName}/runtime:${dockerDir}/runtime -v ${pathName}/upload:${dockerDir}/upload ${name}