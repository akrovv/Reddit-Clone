#!/bin/bash

test1=$(curl -s -i -I -X 'GET' "http://localhost:8080/")
test2=$(curl -s -i -I -X 'GET' "http://localhost:8080/static/")
test3=$(curl -s -i -I -X 'GET' "http://localhost:8080/signup")
test4=$(curl -s -i -I -X 'GET' "http://localhost:8080/login")
test5=$(curl -s -i -I -X 'GET' "http://localhost:8080/u/akro")
test6=$(curl -s -i -I -X 'GET' "http://localhost:8080/a/music")
test7=$(curl -s -i -I -X 'GET' "http://localhost:8080/a/music/id")
test8=$(curl -s -i -I -X 'GET' "http://localhost:8080/api/posts/")
test9=$(curl -s -i -I -X 'GET' "http://localhost:8080/api/posts/music")
test10=$(curl -s -i -I -X 'GET' "http://localhost:8080/api/user/akrov")
test11=$(curl -s -i -I -X 'GET' "http://localhost:8080/createpost")


if ! curl -IsSf "http://localhost:8080" &> /dev/null; then
    echo "Ошибка: curl не работает или не удалось получить ответ от сервера."
    exit 1
fi

i=1

while [ $i -le 10 ]; do
    result=$(eval "echo \$test$i" | grep "HTTP/1.1" | awk '{print $2}')

    if [ $result -ne "200" ]; then
        echo $i $result 
        exit 1
    fi

    i=$((i+1))
done

test1=$(curl -s -i -I -X 'POST' "http://localhost:8080/api/register")
test2=$(curl -s -i -I -X 'POST' "http://localhost:8080/api/login")
test3=$(curl -s -i -I -X 'POST' "http://localhost:8080/api/posts/")
test4=$(curl -s -i -I -X 'POST' "http://localhost:8080/api/post/id")
test5=$(curl -s -i -I -X 'DELETE' "http://localhost:8080/api/post/id")
test6=$(curl -s -i -I -X 'DELETE' "http://localhost:8080/api/post/id/id")

i=1

while [ $i -le 6 ]; do
    result=$(eval "echo \$test$i" | grep "HTTP/1.1" | awk '{print $2}')

    if [ $result -eq "200" ]; then
        echo $i $result 
        exit 1
    fi

    i=$((i+1))
done


