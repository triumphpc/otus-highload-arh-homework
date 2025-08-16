- добавлен контейнер с HAPROXY
- проверил работу master/slave

```shell
psql -h localhost -p 5001 -U postgres -d app_db
psql (14.18 (Homebrew), server 17.2 (Debian 17.2-1.pgdg120+1))
WARNING: psql major version 14, server major version 17.
         Some psql features might not work.
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

app_db=# 



psql -h localhost -p 5002 -U postgres -d app_db
psql (14.18 (Homebrew), server 17.2 (Debian 17.2-1.pgdg120+1))
WARNING: psql major version 14, server major version 17.
         Some psql features might not work.
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

app_db=# 
```

- поднял два приложения и nginx над ними в отдельном контейнер
```shell

docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}" | (read -r; printf "\e[1;33m%s\e[0m\n" "$REPLY"; column -t -s $'\t')
CONTAINER ID   NAMES                    STATUS                             PORTS                                                                                          IMAGE
7f77204d3c91   nginx                    Up 1 second                        0.0.0.0:80->80/tcp, [::]:80->80/tcp, 0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp               nginx:1.21
bb79a8a87d80   dialog                   Up 7 seconds (health: starting)    0.0.0.0:50051->50051/tcp, [::]:50051->50051/tcp                                                docker-dialog
35ce31d15b94   feed-updater             Up 7 seconds                       0.0.0.0:8081->8081/tcp, [::]:8081->8081/tcp                                                    docker-feed-updater
ed4742528033   grafana                  Up 17 seconds (health: starting)   0.0.0.0:3000->3000/tcp, [::]:3000->3000/tcp                                                    grafana/grafana:latest
6d7dd8ab8965   app1                     Up 7 seconds (healthy)             0.0.0.0:8082->8080/tcp, [::]:8082->8080/tcp                                                    docker-app1
38e5a3edac08   app2                     Up 7 seconds (healthy)             0.0.0.0:8083->8080/tcp, [::]:8083->8080/tcp                                                    docker-app2
7aa24d3b2e9c   haproxy                  Up 18 seconds (health: starting)   0.0.0.0:5000-5002->5000-5002/tcp, [::]:5000-5002->5000-5002/tcp                                haproxy:latest
0eea91eca89c   prometheus               Up 18 seconds                      0.0.0.0:9090->9090/tcp, [::]:9090->9090/tcp                                                    prom/prometheus
40425974cadc   docker-worker1-1         Up 18 seconds (healthy)            5432/tcp                                                                                       citusdata/citus:13.0.3
16736d48e68b   docker-worker2-1         Up 18 seconds (healthy)            5432/tcp                                                                                       citusdata/citus:13.0.3
e678a27845e0   manager                  Up 18 seconds (healthy)                                                                                                           citusdata/membership-manager:0.3.0
37699b7e7073   redis                    Up 18 seconds (healthy)            0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp                                                    redis:latest
bbaf430e0a15   kafka                    Up 18 seconds (healthy)            0.0.0.0:9092->9092/tcp, [::]:9092->9092/tcp, 0.0.0.0:29093->29093/tcp, [::]:29093->29093/tcp   confluentinc/cp-kafka:latest
a83b98c0bba8   postgres-exporter-2      Up 18 seconds                      0.0.0.0:9188->9187/tcp, [::]:9188->9187/tcp                                                    prometheuscommunity/postgres-exporter
2b0ba4812764   docker-node-exporter-1   Up 18 seconds                      0.0.0.0:9100->9100/tcp, [::]:9100->9100/tcp                                                    prom/node-exporter
563d5dd29fc8   postgres-exporter-3      Up 18 seconds                      0.0.0.0:9189->9187/tcp, [::]:9189->9187/tcp                                                    prometheuscommunity/postgres-exporter
628b72f6d026   postgres-exporter-1      Up 18 seconds                      0.0.0.0:9187->9187/tcp, [::]:9187->9187/tcp                                                    prometheuscommunity/postgres-exporter
97cabb277e3b   master                   Up 18 seconds (healthy)            0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp                                                    citusdata/citus:13.0.3          citusdata/citus:13.0.3
```


- проверим запросы к приложению
```shell

curl --location 'http://localhost:8080/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data-raw '{
  "birth_date": "2006-01-02T15:04:05Z",
  "city": "Moscow",
  "email": "1234@tes.ru",
  "first_name": "First",
  "gender": "female",
  "last_name": "Last",
  "password": "password",
  "interests": [
    "some",
    "up"
  ]
}'
{"details":"user already exists","error":"Internal server error"}%                                                                                                        
curl --location 'http://localhost:8080/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data-raw '{
  "birth_date": "2006-01-02T15:04:05Z",
  "city": "Moscow",
  "email": "1235@tes.ru",
  "first_name": "First",
  "gender": "female",
  "last_name": "Last",
  "password": "password",
  "interests": [
    "some",
    "up"
  ]
}'
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTUzNTA5MTgsImlhdCI6MTc1NTM0OTExOCwic3ViIjowfQ.e2QPegFbVrVTaibUH5kOYQ3-Iw2MAezFPHmNceCOfQ8","user":{"id":0,"first_name":"First","last_name":"Last","email":"1235@tes.ru","birth_date":"2006-01-02T15:04:05Z","gender":"female","interests":["some","up"],"city":"Moscow","created_at":"2025-08-16T12:58:38.936710386Z","is_adult":true}}%     
```

- потушим один слэйв, проверим что работает приложение
```shell

docker stop 16736d48e68b
16736d48e68b
docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}" | (read -r; printf "\e[1;33m%s\e[0m\n" "$REPLY"; column -t -s $'\t')
CONTAINER ID   NAMES                    STATUS                     PORTS                                                                                          IMAGE
7f77204d3c91   nginx                    Up 2 minutes               0.0.0.0:80->80/tcp, [::]:80->80/tcp, 0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp               nginx:1.21
bb79a8a87d80   dialog                   Up 2 minutes (healthy)     0.0.0.0:50051->50051/tcp, [::]:50051->50051/tcp                                                docker-dialog
35ce31d15b94   feed-updater             Up 2 minutes               0.0.0.0:8081->8081/tcp, [::]:8081->8081/tcp                                                    docker-feed-updater
ed4742528033   grafana                  Up 3 minutes (healthy)     0.0.0.0:3000->3000/tcp, [::]:3000->3000/tcp                                                    grafana/grafana:latest
6d7dd8ab8965   app1                     Up 2 minutes (healthy)     0.0.0.0:8082->8080/tcp, [::]:8082->8080/tcp                                                    docker-app1
38e5a3edac08   app2                     Up 2 minutes (healthy)     0.0.0.0:8083->8080/tcp, [::]:8083->8080/tcp                                                    docker-app2
7aa24d3b2e9c   haproxy                  Up 3 minutes (unhealthy)   0.0.0.0:5000-5002->5000-5002/tcp, [::]:5000-5002->5000-5002/tcp                                haproxy:latest
0eea91eca89c   prometheus               Up 3 minutes               0.0.0.0:9090->9090/tcp, [::]:9090->9090/tcp                                                    prom/prometheus
40425974cadc   docker-worker1-1         Up 3 minutes (healthy)     5432/tcp                                                                                       citusdata/citus:13.0.3
e678a27845e0   manager                  Up 3 minutes (healthy)                                                                                                    citusdata/membership-manager:0.3.0
37699b7e7073   redis                    Up 3 minutes (healthy)     0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp                                                    redis:latest
bbaf430e0a15   kafka                    Up 3 minutes (healthy)     0.0.0.0:9092->9092/tcp, [::]:9092->9092/tcp, 0.0.0.0:29093->29093/tcp, [::]:29093->29093/tcp   confluentinc/cp-kafka:latest
a83b98c0bba8   postgres-exporter-2      Up 3 minutes               0.0.0.0:9188->9187/tcp, [::]:9188->9187/tcp                                                    prometheuscommunity/postgres-exporter
2b0ba4812764   docker-node-exporter-1   Up 3 minutes               0.0.0.0:9100->9100/tcp, [::]:9100->9100/tcp                                                    prom/node-exporter
563d5dd29fc8   postgres-exporter-3      Up 3 minutes               0.0.0.0:9189->9187/tcp, [::]:9189->9187/tcp                                                    prometheuscommunity/postgres-exporter
628b72f6d026   postgres-exporter-1      Up 3 minutes               0.0.0.0:9187->9187/tcp, [::]:9187->9187/tcp                                                    prometheuscommunity/postgres-exporter
97cabb277e3b   master                   Up 3 minutes (healthy)     0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp                                                    citusdata/citus:13.0.3

curl --location 'http://localhost:8080/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data-raw '{
  "birth_date": "2006-01-02T15:04:05Z",
  "city": "Moscow",
  "email": "555@tes.ru",
  "first_name": "First",
  "gender": "female",
  "last_name": "Last",
  "password": "password",
  "interests": [
    "some",
    "up"
  ]
}'
{"details":"user already exists","error":"Internal server error"}%                                                                                                        
curl --location 'http://localhost:8080/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data-raw '{
  "birth_date": "2006-01-02T15:04:05Z",
  "city": "Moscow",
  "email": "556@tes.ru",
  "first_name": "First",
  "gender": "female",
  "last_name": "Last",
  "password": "password",
  "interests": [
    "some",
    "up"
  ]
}'
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTUzNTEwMTQsImlhdCI6MTc1NTM0OTIxNCwic3ViIjowfQ.WlB343VW9QktcqGqvCG21Lmoz8EoJ9-HUXWwDL-6BQs","user":{"id":0,"first_name":"First","last_name":"Last","email":"556@tes.ru","birth_date":"2006-01-02T15:04:05Z","gender":"female","interests":["some","up"],"city":"Moscow","created_at":"2025-08-16T13:00:14.937894333Z","is_adult":true}}%   
```

- теперь положим один app1, и проверим, что запросы все еще обрабатываются
```shell

docker stop 6d7dd8ab8965
6d7dd8ab8965
docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}" | (read -r; printf "\e[1;33m%s\e[0m\n" "$REPLY"; column -t -s $'\t')
CONTAINER ID   NAMES                    STATUS                     PORTS                                                                                          IMAGE
7f77204d3c91   nginx                    Up 4 minutes               0.0.0.0:80->80/tcp, [::]:80->80/tcp, 0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp               nginx:1.21
bb79a8a87d80   dialog                   Up 4 minutes (healthy)     0.0.0.0:50051->50051/tcp, [::]:50051->50051/tcp                                                docker-dialog
35ce31d15b94   feed-updater             Up 4 minutes               0.0.0.0:8081->8081/tcp, [::]:8081->8081/tcp                                                    docker-feed-updater
ed4742528033   grafana                  Up 4 minutes (healthy)     0.0.0.0:3000->3000/tcp, [::]:3000->3000/tcp                                                    grafana/grafana:latest
38e5a3edac08   app2                     Up 4 minutes (healthy)     0.0.0.0:8083->8080/tcp, [::]:8083->8080/tcp                                                    docker-app2
7aa24d3b2e9c   haproxy                  Up 4 minutes (unhealthy)   0.0.0.0:5000-5002->5000-5002/tcp, [::]:5000-5002->5000-5002/tcp                                haproxy:latest
0eea91eca89c   prometheus               Up 4 minutes               0.0.0.0:9090->9090/tcp, [::]:9090->9090/tcp                                                    prom/prometheus
40425974cadc   docker-worker1-1         Up 4 minutes (healthy)     5432/tcp                                                                                       citusdata/citus:13.0.3
e678a27845e0   manager                  Up 4 minutes (healthy)                                                                                                    citusdata/membership-manager:0.3.0
37699b7e7073   redis                    Up 4 minutes (healthy)     0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp                                                    redis:latest
bbaf430e0a15   kafka                    Up 4 minutes (healthy)     0.0.0.0:9092->9092/tcp, [::]:9092->9092/tcp, 0.0.0.0:29093->29093/tcp, [::]:29093->29093/tcp   confluentinc/cp-kafka:latest
a83b98c0bba8   postgres-exporter-2      Up 4 minutes               0.0.0.0:9188->9187/tcp, [::]:9188->9187/tcp                                                    prometheuscommunity/postgres-exporter
2b0ba4812764   docker-node-exporter-1   Up 4 minutes               0.0.0.0:9100->9100/tcp, [::]:9100->9100/tcp                                                    prom/node-exporter
563d5dd29fc8   postgres-exporter-3      Up 4 minutes               0.0.0.0:9189->9187/tcp, [::]:9189->9187/tcp                                                    prometheuscommunity/postgres-exporter
628b72f6d026   postgres-exporter-1      Up 4 minutes               0.0.0.0:9187->9187/tcp, [::]:9187->9187/tcp                                                    prometheuscommunity/postgres-exporter
97cabb277e3b   master                   Up 4 minutes (healthy)     0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp                                                    citusdata/citus:13.0.3


curl --location 'http://localhost:8080/api/v1/auth/register' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data-raw '{
  "birth_date": "2006-01-02T15:04:05Z",
  "city": "Moscow",
  "email": "665@tes.ru",
  "first_name": "First",
  "gender": "female",
  "last_name": "Last",
  "password": "password",
  "interests": [
    "some",
    "up"
  ]
}'
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTUzNTExMDEsImlhdCI6MTc1NTM0OTMwMSwic3ViIjowfQ.vjnQNxYGPR5LKnFgnKk49cRwLeq4Gf4gYubHvZRKsp0","user":{"id":0,"first_name":"First","last_name":"Last","email":"665@tes.ru","birth_date":"2006-01-02T15:04:05Z","gender":"female","interests":["some","up"],"city":"Moscow","created_at":"2025-08-16T13:01:41.503874054Z","is_adult":true}}%  
```

- логи из haproxy
```shell
2025-08-16T12:56:51.072557003Z Connect from 172.27.0.19:35390 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.334300586Z Connect from 172.27.0.17:41372 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.334313420Z Connect from 172.27.0.17:41360 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.334467378Z Connect from 172.27.0.17:41382 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.334700878Z Connect from 172.27.0.17:41402 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.334705128Z Connect from 172.27.0.17:41404 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.334706086Z Connect from 172.27.0.17:41394 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:56:51.355894128Z Connect from 172.27.0.17:41408 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T12:59:57.706711214Z [WARNING]  (8) : Backup Server postgres_read/slave2 is DOWN, reason: Layer4 timeout, info: " at step 1 of tcp-check (connect)", check duration: 3004ms. 1 active and 0 backup servers left. 0 sessions active, 0 requeued, 0 remaining in queue.
2025-08-16T12:59:57.706868131Z Backup Server postgres_read/slave2 is DOWN, reason: Layer4 timeout, info: " at step 1 of tcp-check (connect)", check duration: 3004ms. 1 active and 0 backup servers left. 0 sessions active, 0 requeued, 0 remaining in queue.
2025-08-16T13:01:51.424256586Z Connect from 172.27.0.16:45938 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T13:01:51.619967336Z Connect from 172.27.0.19:56836 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T13:01:51.620021545Z Connect from 172.27.0.18:55610 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T13:01:51.955053628Z Connect from 172.27.0.16:45952 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T13:01:52.152073378Z Connect from 172.27.0.19:56844 to 172.27.0.13:5001 (postgres_write/TCP)
2025-08-16T13:01:52.152127170Z Connect from 172.27.0.18:55614 to 172.27.0.13:5001 (postgres_write/TCP)
```

- nginx просто балансит трафик

В целом достигли что требовалось